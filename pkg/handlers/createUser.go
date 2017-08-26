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

/*
CreateUserHandler creates a user in the database, but also:
	- writes a log of the request and response
	- "pretty prints" the request
*/
func CreateUserHandler(env *env.Env, w http.ResponseWriter, req *http.Request) error {
	// retrieve the context from the http.Request
	ctx := req.Context()

	logger := env.Logger
	defer env.Logger.Sync()

	logger.Debug("CreateUserHandler started")
	defer logger.Debug("CreateUserHandler ended")

	err := LogRequest(env, req)

	if err != nil {
		return err
	}

	// db.NewContext function creates and begins a new sql.Tx, which pulls from the
	// previously opened database (postgres) connection pool and starts a database
	// transaction.  In addition, the pointer to this "started" sql.Tx is included
	// in the above created context
	ctx = db.Tx2Context(ctx, env, nil)

	var usr *appUser.User
	err = json.NewDecoder(req.Body).Decode(&usr)
	if err != nil {
		return HTTPStatusError{500, err}
	}
	defer req.Body.Close()

	// Call the create method of the appUser object to validate data and write to db
	logsWritten, err := usr.Create(ctx)

	tx, ok := db.TxFromContext(ctx)

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
		return HTTPStatusError{http.StatusInternalServerError, err}
	}

	fmt.Fprintf(w, "logsWritten = %d\n", logsWritten)

	return nil

}
