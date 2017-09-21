package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gilcrest/go-API-template/pkg/appUser"
	"github.com/gilcrest/go-API-template/pkg/config/db"
	"github.com/gilcrest/go-API-template/pkg/config/env"
)

// CreateUserHandler creates a user in the database
func CreateUserHandler(env *env.Env, w http.ResponseWriter, req *http.Request) error {

	// retrieve the context from the http.Request
	ctx := req.Context()

	// Get a new logger instance
	logger := env.Logger
	defer env.Logger.Sync()

	logger.Debug("CreateUserHandler started")
	defer logger.Debug("CreateUserHandler ended")

	var err error

	// db.NewContext function creates and begins a new sql.Tx, which pulls from the
	// previously opened database (postgres) connection pool and starts a database
	// transaction.  In addition, the pointer to this "started" sql.Tx is added to
	// the above created context
	ctx = db.Tx2Context(ctx, env, nil)

	// Declare usr as an instance of appUser.User
	// Decode JSON HTTP request body into a Decoder type
	//  and unmarshal that into usr
	var usr *appUser.User
	err = json.NewDecoder(req.Body).Decode(&usr)
	if err != nil {
		return HTTPStatusError{http.StatusInternalServerError, err}
	}
	defer req.Body.Close()

	// Call the create method of the appUser object to validate data and write to db
	rows, err := usr.Create(ctx)

	// TxFromContext extracts the database transaction from the context, if present.
	tx, ok := db.TxFromContext(ctx)

	// If we have successfully written rows to the db, we commit the transaction
	if ok && rows > 0 {
		err = tx.Commit()
		if err != nil {
			// We return a status error here, which conveniently wraps the error
			// returned from our DB queries. We can clearly define which errors
			// are worth raising a HTTP 500 over vs. which might just be a HTTP
			// 404, 403 or 401 (as appropriate). It's also clear where our
			// handler should stop processing by returning early.
			return HTTPStatusError{http.StatusInternalServerError, err}
		}
	} else if rows <= 0 {
		return HTTPStatusError{http.StatusInternalServerError, err}
	}

	// TODO - Get a unique Request ID and add it to the header and logs via
	// a middleware
	w.Header().Set("Request-Id", "123456789")

	// Encode usr struct to JSON for the response body
	json.NewEncoder(w).Encode(*usr)

	return nil

}
