package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gilcrest/go-API-template/appUser"
	"github.com/gilcrest/go-API-template/env"
	"github.com/gilcrest/go-API-template/server/errorHandler"
)

// CreateUser creates a user in the database
func CreateUser(env *env.Env, w http.ResponseWriter, req *http.Request) error {

	// Get a new logger instance
	log := env.Logger

	log.Debug().Msg("Start CreateUserHandler")
	defer log.Debug().Msg("Finish CreateUserHandler")

	// retrieve the context from the http.Request
	ctx := req.Context()

	var err error

	// Declare usr as an instance of appUser.User
	// Decode JSON HTTP request body into a Decoder type
	//  and unmarshal that into usr
	var usr *appUser.User
	err = json.NewDecoder(req.Body).Decode(&usr)
	if err != nil {
		return errorHandler.HTTPErr{
			Code: http.StatusBadRequest,
			Type: "invalid_request",
			Err:  err,
		}
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
			return errorHandler.HTTPErr{
				Code: http.StatusBadRequest,
				Type: "database_error",
				Err:  err,
			}
		}
	} else {
		return errorHandler.HTTPErr{
			Code: http.StatusBadRequest,
			Type: "database_error",
			Err:  errors.New("Create Date not set"),
		}
	}

	// Encode usr struct to JSON for the response body
	json.NewEncoder(w).Encode(*usr)

	return nil

}
