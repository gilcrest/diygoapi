package app

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gilcrest/go-API-template/datastore"
	"github.com/gilcrest/go-API-template/lib/usr"
	"github.com/gilcrest/httplog"
)

// CreateUser creates a user in the database
func (s *server) handleUserCreate(w http.ResponseWriter, req *http.Request) error {

	// request is the expected service request fields
	type request struct {
		Username     string `json:"username"`
		Password     string `json:"password"`
		MobileID     string `json:"mobile_id"`
		Email        string `json:"email"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		UpdateUserID string `json:"udpate_user_id"`
	}

	// response is the expected service response fields
	type response struct {
		Username       string         `json:"username"`
		MobileID       string         `json:"mobile_id"`
		Email          string         `json:"email"`
		FirstName      string         `json:"first_name"`
		LastName       string         `json:"last_name"`
		UpdateUserID   string         `json:"update_user_id"`
		UpdateUnixTime int64          `json:"created"`
		Audit          *httplog.Audit `json:"audit"`
	}

	// Get logger instance
	log := s.logger

	// retrieve the context from the http.Request
	ctx := req.Context()

	var err error

	// Declare cur as an instance of createUserRequest
	// Decode JSON HTTP request body into a Decoder type
	//  and unmarshal that into cur
	rqst := new(request)
	err = json.NewDecoder(req.Body).Decode(&rqst)
	if err != nil {
		return HTTPErr{
			Code: http.StatusBadRequest,
			Type: "invalid_request",
			Err:  err,
		}
	}
	defer req.Body.Close()

	// declare a new instance of usr.User
	usr := new(usr.User)

	err = usr.SetUsername(rqst.Username)
	if err != nil {
		return HTTPErr{
			Code: http.StatusBadRequest,
			Type: "validation_error",
			Err:  err,
		}
	}
	err = usr.SetPassword(ctx, rqst.Password)
	if err != nil {
		return HTTPErr{
			Code: http.StatusBadRequest,
			Type: "validation_error",
			Err:  err,
		}
	}
	err = usr.SetMobileID(rqst.MobileID)
	if err != nil {
		return HTTPErr{
			Code: http.StatusBadRequest,
			Type: "validation_error",
			Err:  err,
		}
	}
	err = usr.SetEmail(rqst.Email)
	if err != nil {
		return HTTPErr{
			Code: http.StatusBadRequest,
			Type: "validation_error",
			Err:  err,
		}
	}
	err = usr.SetFirstName(rqst.FirstName)
	if err != nil {
		return HTTPErr{
			Code: http.StatusBadRequest,
			Type: "validation_error",
			Err:  err,
		}
	}
	err = usr.SetLastName(rqst.LastName)
	if err != nil {
		return HTTPErr{
			Code: http.StatusBadRequest,
			Type: "validation_error",
			Err:  err,
		}
	}
	err = usr.SetUpdateClientID("TBD")
	if err != nil {
		return HTTPErr{
			Code: http.StatusBadRequest,
			Type: "validation_error",
			Err:  err,
		}
	}
	err = usr.SetUpdateUserID(rqst.UpdateUserID)
	if err != nil {
		return HTTPErr{
			Code: http.StatusBadRequest,
			Type: "validation_error",
			Err:  err,
		}
	}

	// Call the create method of the User object to validate data and write to db
	err = usr.Create(ctx, log)
	if err != nil {
		return HTTPErr{
			Code: http.StatusBadRequest,
			Type: "validation_failed",
			Err:  err,
		}
	}

	// get a new DB Tx
	tx, err := s.ds.BeginTx(ctx, nil, datastore.AppDB)
	if err != nil {
		return HTTPErr{
			Code: http.StatusBadRequest,
			Type: "database_error",
			Err:  err,
		}
	}

	// Call the create method of the User object to write
	// to the database
	err = usr.CreateDB(ctx, log, tx)
	if err != nil {
		return HTTPErr{
			Code: http.StatusBadRequest,
			Type: "database_error",
			Err:  err,
		}
	}

	if !usr.UpdateTimestamp().IsZero() {
		err := tx.Commit()
		if err != nil {
			return HTTPErr{
				Code: http.StatusBadRequest,
				Type: "database_error",
				Err:  errors.New("Database error, contact support"),
			}
		}
	} else {
		err = tx.Rollback()
		if err != nil {
			return HTTPErr{
				Code: http.StatusBadRequest,
				Type: "database_error",
				Err:  errors.New("Database error, contact support"),
			}
		}
	}

	resp := new(response)

	// If we successfully committed the db transaction, we can consider this
	// transaction successful and return a response with the response body
	aud, err := httplog.NewAudit(ctx, nil)
	if err != nil {
		return HTTPErr{
			Code: http.StatusInternalServerError,
			Type: "encode_error",
			Err:  err,
		}
	}

	resp.Audit = aud
	resp.Username = usr.Username()
	resp.MobileID = usr.MobileID()
	resp.Email = usr.Email()
	resp.FirstName = usr.FirstName()
	resp.LastName = usr.LastName()
	resp.UpdateUserID = usr.UpdateUserID()
	resp.UpdateUnixTime = usr.UpdateTimestamp().Unix()

	// Encode usr struct to JSON for the response body
	json.NewEncoder(w).Encode(*resp)
	if err != nil {
		return HTTPErr{
			Code: http.StatusInternalServerError,
			Type: "encode_error",
			Err:  err,
		}
	}

	return nil

}
