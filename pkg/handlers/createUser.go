package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gilcrest/go-API-template/pkg/appUser"
	"github.com/gilcrest/go-API-template/pkg/config/db"
	"github.com/gilcrest/go-API-template/pkg/config/env"
)

// Error represents a handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type Error interface {
	error
	Status() int
}

// HTTPStatusError represents an error with an associated HTTP status code.
type HTTPStatusError struct {
	Code int
	Err  error
}

// Allows HTTPStatusError to satisfy the error interface.
func (hse HTTPStatusError) Error() string {
	return hse.Err.Error()
}

// Returns our HTTP status code.
func (hse HTTPStatusError) Status() int {
	return hse.Code
}

// The Handler struct that takes a configured Env and a function matching
// our useful signature.
type Handler struct {
	Env *env.Env
	H   func(e *env.Env, w http.ResponseWriter, r *http.Request) error
}

// ServeHTTP allows our Handler type to satisfy http.Handler.
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.H(h.Env, w, r)
	if err != nil {
		switch e := err.(type) {
		case Error:
			// We can retrieve the status here and write out a specific
			// HTTP status code.
			log.Printf("HTTP %d - %s", e.Status(), e)
			http.Error(w, e.Error(), e.Status())
		default:
			// Any error types we don't specifically look out for default
			// to serving a HTTP 500
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
		}
	}
}

/*
Creates a user in the database, but also:
	- writes a log of the request and response
	- "pretty prints" the request
*/
func CreateUserHandler(env *env.Env, w http.ResponseWriter, req *http.Request) error {
	// retrieve the context from the http.Request
	ctx := req.Context()
	logger := env.Logger
	logger.Debug("handleMbrLog started")

	defer env.Logger.Sync()
	defer logger.Debug("handleMbrLog ended")

	//logRequest(req)
	//prettyPrintRequest(req)

	// db.NewContext function creates and begins a new sql.Tx, which pulls from the
	// previously opened database (postgres) connection pool and starts a database
	// transaction.  In addition, the pointer to this "started" sql.Tx is included
	// in the above created context
	ctx = db.AddDBTx2Context(ctx, env, nil)

	var usr *appUser.User
	err := json.NewDecoder(req.Body).Decode(&usr)
	if err != nil {
		return HTTPStatusError{500, err}
	}
	defer req.Body.Close()

	log.Println(usr)

	// Call the create method of the appUser object to validate data and write to db
	logsWritten, err := usr.Create(ctx)

	tx, ok := db.DBTxFromContext(ctx)

	if ok && logsWritten > 0 {
		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}
	} else if logsWritten <= 0 {
		log.Fatal(err)
	}

	if err != nil {
		// We return a status error here, which conveniently wraps the error
		// returned from our DB queries. We can clearly define which errors
		// are worth raising a HTTP 500 over vs. which might just be a HTTP
		// 404, 403 or 401 (as appropriate). It's also clear where our
		// handler should stop processing by returning early.
		return HTTPStatusError{500, err}
	}

	fmt.Fprintf(w, "logsWritten = %d\n", logsWritten)

	return nil

}
