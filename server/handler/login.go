package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gilcrest/go-API-template/auth"
	"github.com/gilcrest/go-API-template/db"
	"github.com/gilcrest/go-API-template/env"
	"github.com/gilcrest/go-API-template/errors"
	"github.com/gilcrest/go-API-template/server/errorHandler"
)

// LoginHandler is for user login
func LoginHandler(env *env.Env, w http.ResponseWriter, req *http.Request) error {

	// Get a new logger instance
	log := env.Logger

	log.Debug().Msg("Start handler.LoginHandler")
	defer log.Debug().Msg("Finish handler.LoginHandler")

	// retrieve the context from the http.Request
	ctx := req.Context()

	// Declare creds as an instance of auth.Credentials
	// Decode JSON HTTP request body into a Decoder type
	//  and unmarshal that into creds
	creds := new(auth.Credentials)
	err := json.NewDecoder(req.Body).Decode(&creds)
	if err != nil {
		return errorHandler.HTTPErr{
			Code: http.StatusBadRequest,
			Type: "invalid_request",
			Err:  err,
		}
	}
	defer req.Body.Close()

	tx, err := env.DS.BeginTx(ctx, nil, db.AppDB)
	if err != nil {
		return errorHandler.HTTPErr{
			Code: http.StatusBadRequest,
			Type: "database_error",
			Err:  err,
		}
	}

	// Pass user credentials to auth.Authorise
	// If an error is passed back, send a generic error so as not
	// to allow end user to know if it was the username or password
	// we can tell via logs what it was, but don't want end user to
	// know (bad security practice)
	usr, err := auth.Authorize(ctx, log, tx, creds)
	if err != nil {
		return errorHandler.HTTPErr{
			Code: http.StatusUnauthorized,
			Type: "unauthorised",
			Err:  errors.Str("Wrong email or password"),
		}
	}

	// TODO - do I need to close tx?

	jwt, err := auth.LoginToken(usr)
	if err != nil {
		return errorHandler.HTTPErr{
			Code: http.StatusUnauthorized,
			Type: "unauthorised",
			Err:  errors.Str("Wrong email or password"),
		}
	}

	json.NewEncoder(w).Encode(jwt)

	return nil
}
