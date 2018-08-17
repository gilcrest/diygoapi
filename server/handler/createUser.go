package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gilcrest/go-API-template/server/response"

	"github.com/gilcrest/go-API-template/appuser"
	"github.com/gilcrest/go-API-template/db"
	"github.com/gilcrest/go-API-template/env"
	"github.com/gilcrest/go-API-template/server/errorHandler"
)

// CreateUser creates a user in the database
func CreateUser(env *env.Env, w http.ResponseWriter, req *http.Request) error {

	// Get a new logger instance
	log := env.Logger

	log.Debug().Msg("Start handler.CreateUser")
	defer log.Debug().Msg("Finish handler.CreateUser")

	// retrieve the context from the http.Request
	ctx := req.Context()

	var err error

	// Declare cur as an instance of createUserRequest
	// Decode JSON HTTP request body into a Decoder type
	//  and unmarshal that into cur
	cur := new(appuser.CreateUserRequest)
	err = json.NewDecoder(req.Body).Decode(&cur)
	if err != nil {
		return errorHandler.HTTPErr{
			Code: http.StatusBadRequest,
			Type: "invalid_request",
			Err:  err,
		}
	}
	defer req.Body.Close()

	usr, err := appuser.NewUser(ctx, env, cur)
	if err != nil {
		// initialize an errorHandler with the default Code and Type for
		// service validations (Err is set to nil as it will be set later)
		return errorHandler.HTTPErr{
			Code: http.StatusBadRequest,
			Type: "validation_error",
			Err:  err,
		}
	}

	// Call the create method of the appuser object to validate data and write to db
	err = usr.Create(ctx, log)
	if err != nil {
		return errorHandler.HTTPErr{
			Code: http.StatusBadRequest,
			Type: "validation_failed",
			Err:  err,
		}
	}

	// get a new DB Tx
	tx, err := env.DS.BeginTx(ctx, nil, db.AppDB)
	if err != nil {
		return errorHandler.HTTPErr{
			Code: http.StatusBadRequest,
			Type: "database_error",
			Err:  err,
		}
	}

	// Call the create method of the User object to write
	// to the database
	err = usr.CreateDB(ctx, log, tx)
	if err != nil {
		return errorHandler.HTTPErr{
			Code: http.StatusBadRequest,
			Type: "database_error",
			Err:  err,
		}
	}

	if !usr.UpdateTimestamp().IsZero() {
		err := tx.Commit()
		if err != nil {
			return errorHandler.HTTPErr{
				Code: http.StatusBadRequest,
				Type: "database_error",
				Err:  err,
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

	// If we successfully committed the db transaction, we can consider this
	// transaction successful and return a response with the response body
	curs, err := newCreateUserResponse(usr)
	if err != nil {
		return errorHandler.HTTPErr{
			Code: http.StatusBadRequest,
			Type: "response_error",
			Err:  errors.New("Unable to setup Response"),
		}
	}

	// Encode usr struct to JSON for the response body
	json.NewEncoder(w).Encode(*curs)
	if err != nil {
		return errorHandler.HTTPErr{
			Code: http.StatusInternalServerError,
			Type: "encode_error",
			Err:  err,
		}
	}

	return nil

}

// newCreateMemberResponse wires the member object to createMemberResponse object
// if you need to perform any manipulation from your business object to the response object
// you can do it here.  For instance, here's where I convert the CreateDate from a timestamp
// to Unix time
func newCreateUserResponse(usr *appuser.User) (*appuser.CreateUserResponse, error) {

	aud, err := response.NewInfo()
	if err != nil {
		return nil, errorHandler.HTTPErr{
			Code: http.StatusInternalServerError,
			Type: "encode_error",
			Err:  err,
		}
	}

	cur := new(appuser.CreateUserResponse)
	cur.Audit = aud
	cur.Username = usr.Username()
	cur.MobileID = usr.MobileID()
	cur.Email = usr.Email()
	cur.FirstName = usr.FirstName()
	cur.LastName = usr.LastName()
	cur.UpdateUserID = usr.UpdateUserID()
	cur.UpdateUnixTime = usr.UpdateTimestamp().Unix()

	return cur, nil
}
