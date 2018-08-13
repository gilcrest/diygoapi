package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gilcrest/go-API-template/auth"
	"github.com/gilcrest/go-API-template/db"
	"github.com/gilcrest/go-API-template/env"
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

	// Begin a new LogDB txn
	_, err := env.DS.BeginTx(ctx, nil, db.LogDB)
	if err != nil {
		return err
	}

	// Declare creds as an instance of auth.Credentials
	// Decode JSON HTTP request body into a Decoder type
	//  and unmarshal that into creds
	creds := new(auth.Credentials)
	err = json.NewDecoder(req.Body).Decode(&creds)
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
			Err:  errors.New("Wrong email or password"),
		}
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": usr.Username(),
	})

	tokenString, error := token.SignedString([]byte("secret"))
	if error != nil {
		fmt.Println(error)
	}

	json.NewEncoder(w).Encode(auth.JwtToken{Token: tokenString})

	return nil
}
