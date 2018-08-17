package handler

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gilcrest/go-API-template/auth"
	"github.com/gilcrest/go-API-template/db"
	"github.com/gilcrest/go-API-template/env"
	"github.com/gilcrest/go-API-template/server/errorHandler"
)

// CreateClientHandler is used to create a new client (aka app)
// and generate clientID, clientSecret, etc.
func CreateClientHandler(env *env.Env, w http.ResponseWriter, req *http.Request) error {

	// Get a new logger instance
	log := env.Logger

	log.Debug().Msg("Start handler.CreateClientHandler")
	defer log.Debug().Msg("Finish handler.CreateClientHandler")

	// retrieve the context from the http.Request
	ctx := req.Context()

	// Declare creds as an instance of auth.Credentials
	// Decode JSON HTTP request body into a Decoder type
	//  and unmarshal that into creds
	clientRequest := new(auth.CreateClientRequest)
	err := json.NewDecoder(req.Body).Decode(&clientRequest)
	if err != nil {
		return errorHandler.HTTPErr{
			Code: http.StatusBadRequest,
			Type: "invalid_request",
			Err:  err,
		}
	}
	defer req.Body.Close()

	client, err := auth.NewClient(ctx, env, clientRequest)
	if err != nil {
		// initialize an errorHandler with the default Code and Type for
		// service validations (Err is set to nil as it will be set later)
		return errorHandler.HTTPErr{
			Code: http.StatusBadRequest,
			Type: "validation_error",
			Err:  err,
		}
	}

	// Begin the AppDB txn
	tx, err := env.DS.BeginTx(ctx, nil, db.AppDB)
	if err != nil {
		return err
	}

	// Call the CreateClientDB method of the Client object
	// to write to the db
	tx, err = client.CreateClientDB(ctx, tx)
	if err != nil {
		return errorHandler.HTTPErr{
			Code: http.StatusBadRequest,
			Type: "database_error",
			Err:  errors.New("Database error, contact support"),
		}
	}

	if !client.DMLTime.IsZero() {
		err := tx.Commit()
		if err != nil {
			return errorHandler.HTTPErr{
				Code: http.StatusBadRequest,
				Type: "database_error",
				Err:  errors.New("Database error, contact support"),
			}
		}
	} else {
		err = tx.Rollback()
		if err != nil {
			return errorHandler.HTTPErr{
				Code: http.StatusBadRequest,
				Type: "database_error",
				Err:  errors.New("Database error, contact support"),
			}
		}
	}

	json.NewEncoder(w).Encode(&client)

	return nil
}

func commitOrRollback(*sql.Tx, *auth.Client) error {
	return nil
}
