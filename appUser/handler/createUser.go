package handler

import (
	"context"
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

	log.Debug().Msg("Start handler.CreateUserHandler")
	defer log.Debug().Msg("Finish handler.CreateUserHandler")

	// retrieve the context from the http.Request
	ctx := req.Context()

	var err error

	// Declare cur as an instance of createUserRequest
	// Decode JSON HTTP request body into a Decoder type
	//  and unmarshal that into cur
	cur := new(createUserRequest)
	err = json.NewDecoder(req.Body).Decode(&cur)
	if err != nil {
		return errorHandler.HTTPErr{
			Code: http.StatusBadRequest,
			Type: "invalid_request",
			Err:  err,
		}
	}
	defer req.Body.Close()

	usr, err := newUser(ctx, env, cur)
	if err != nil {
		// newUser returns a proper HTTPErr, so just return it
		return err
	}

	// Call the create method of the appUser object to validate data and write to db
	tx, err := usr.Create(ctx, env)
	if err != nil {
		return errorHandler.HTTPErr{
			Code: http.StatusBadRequest,
			Type: "validation_failed",
			Err:  err,
		}
	}

	if !usr.CreateDate.IsZero() {
		// If we have successfully written rows to the db, we commit the transaction
		// CreateDate should only be populated if the db transaction was successful
		// and everything is in order
		err = tx.Commit()
		if err != nil {
			// We return an HTTPErr here, which wraps the error
			// returned from our DB queries. You could argue that you may not
			// want to send db related info back to the caller...
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

type createUserRequest struct {
	Username     string `json:"username"`
	MobileID     string `json:"mobile_id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	CreateUserID string `json:"create_user_id"`
}

type createUserResponse struct {
	Username       string `json:"username"`
	MobileID       string `json:"mobile_id"`
	Email          string `json:"email"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	CreateUserID   string `json:"create_user_id"`
	CreateUnixTime int64  `json:"created"`
}

// newUser performs basic service validations and wires request data
// into User business object
func newUser(ctx context.Context, env *env.Env, cur *createUserRequest) (*appUser.User, error) {

	// Get a new logger instance
	log := env.Logger

	log.Debug().Msg("Start handler.newUser")
	defer log.Debug().Msg("Finish handler.newUser")

	// declare a new instance of appUser.User
	usr := new(appUser.User)

	// initialize an errorHandler with the default Code and Type for
	// service validations (Err is set to nil as it will be set later)
	e := errorHandler.HTTPErr{
		Code: http.StatusBadRequest,
		Type: "validation_error",
		Err:  nil,
	}

	// for each field you can go through whatever validations you wish
	// and use the SetErr method of the HTTPErr struct to add the proper
	// error text
	switch {
	// Username is required
	case cur.Username == "":
		e.SetErr("Username is a required field")
		return nil, e
	// Username cannot be blah...
	case cur.Username == "blah":
		e.SetErr("Username cannot be blah")
		return nil, e
	default:
		usr.Username = cur.Username
	}

	// for brevity for this template, I won't perform other validations at
	// this point and just wire the rest of the input to the business object
	usr.MobileID = cur.MobileID
	usr.Email = cur.Email
	usr.FirstName = cur.FirstName
	usr.LastName = cur.LastName
	usr.CreateUserID = cur.CreateUserID

	return usr, nil

}

// newCreateMemberResponse wires the member object to createMemberResponse object
// if you need to perform any manipulation from your business object to the response object
// you can do it here.  For instance, here's where I convert the CreateDate from a timestamp
// to Unix time
func newCreateUserResponse(usr *appUser.User) (*createUserResponse, error) {
	cur := new(createUserResponse)
	cur.Username = usr.Username
	cur.MobileID = usr.MobileID
	cur.Email = usr.Email
	cur.FirstName = usr.FirstName
	cur.LastName = usr.LastName
	cur.CreateUserID = usr.CreateUserID
	cur.CreateUnixTime = usr.CreateDate.Unix()

	return cur, nil
}
