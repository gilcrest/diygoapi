package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gilcrest/go-API-template/pkg/domain/appUser"
	"github.com/gilcrest/go-API-template/pkg/env"
)

// CreateUserHandler creates a user in the database
func CreateUserHandler(env *env.Env, w http.ResponseWriter, req *http.Request) error {

	// Get a new logger instance
	logger := env.Logger
	defer env.Logger.Sync()

	logger.Debug("Start CreateUserHandler")
	defer logger.Debug("Finish CreateUserHandler")

	// retrieve the context from the http.Request
	ctx := req.Context()

	logger.Debug("CreateUserHandler started")
	defer logger.Debug("CreateUserHandler ended")

	var err error

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
	tx, err := usr.Create(ctx, env)

	// If we have successfully written rows to the db, we commit the transaction
	if !usr.CreateDate.IsZero() {
		err = tx.Commit()
		if err != nil {
			// We return a status error here, which conveniently wraps the error
			// returned from our DB queries. We can clearly define which errors
			// are worth raising a HTTP 500 over vs. which might just be a HTTP
			// 404, 403 or 401 (as appropriate). It's also clear where our
			// handler should stop processing by returning early.
			return HTTPStatusError{http.StatusInternalServerError, err}
		}
	} else {
		return HTTPStatusError{http.StatusInternalServerError, err}
	}

	// TODO - Get a unique Request ID and add it to the header and logs via
	// a middleware
	w.Header().Set("Request-Id", "123456789")

	// Encode usr struct to JSON for the response body
	json.NewEncoder(w).Encode(*usr)

	return nil

}
