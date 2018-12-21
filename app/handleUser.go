package app

import (
	"encoding/json"
	"net/http"

	"github.com/gilcrest/errors"
	"github.com/gilcrest/go-API-template/lib/usr"
	"github.com/gilcrest/httplog"
	"github.com/gilcrest/srvr/datastore"
)

// handleUserCreate creates a user in the database
func (s *Server) handleUserCreate() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		// request is the expected service request fields
		type request struct {
			Username       string `json:"username"`
			Password       string `json:"password"`
			MobileID       string `json:"mobile_id"`
			Email          string `json:"email"`
			FirstName      string `json:"first_name"`
			LastName       string `json:"last_name"`
			CreateUsername string `json:"create_username"`
		}

		// response is the expected service response fields
		type response struct {
			Username       string         `json:"username"`
			MobileID       string         `json:"mobile_id"`
			Email          string         `json:"email"`
			FirstName      string         `json:"first_name"`
			LastName       string         `json:"last_name"`
			UpdateUsername string         `json:"update_username"`
			UpdateUnixTime int64          `json:"created"`
			Audit          *httplog.Audit `json:"audit"`
		}

		// retrieve the context from the http.Request
		ctx := req.Context()

		var err error

		// Declare rqst as an instance of request
		// Decode JSON HTTP request body into a Decoder type
		//  and unmarshal that into rqst
		rqst := new(request)
		err = json.NewDecoder(req.Body).Decode(&rqst)
		defer req.Body.Close()
		if err != nil {
			err = errors.HTTPErr{
				Code: http.StatusBadRequest,
				Kind: errors.Invalid,
				Err:  err,
			}
			errors.HTTPError(w, err)
			return
		}

		// declare a new instance of usr.User
		usr := new(usr.User)
		usr.Username = rqst.Username
		usr.Password = rqst.Password
		usr.MobileID = rqst.MobileID
		usr.Email = rqst.Email
		usr.FirstName = rqst.FirstName
		usr.LastName = rqst.LastName
		usr.CreateClientID = "TODO"
		usr.CreateUsername = rqst.CreateUsername
		usr.UpdateClientID = "TODO"
		usr.UpdateUsername = rqst.CreateUsername

		// Call the create method of the User object to validate data and write to db
		err = usr.Create(ctx, s.Logger)
		if err != nil {
			err = errors.HTTPErr{
				Code: http.StatusBadRequest,
				Kind: errors.Validation,
				Err:  err,
			}
			errors.HTTPError(w, err)
			return
		}

		// get a new DB Tx
		tx, err := s.DS.BeginTx(ctx, nil, datastore.AppDB)
		if err != nil {
			err = errors.HTTPErr{
				Code: http.StatusBadRequest,
				Kind: errors.Database,
				Err:  err,
			}
			errors.HTTPError(w, err)
			return
		}

		// Call the create method of the User object to write
		// to the database
		err = usr.CreateDB(ctx, s.Logger, tx)
		if err != nil {
			err = errors.HTTPErr{
				Code: http.StatusBadRequest,
				Kind: errors.Database,
				Err:  err,
			}
			errors.HTTPError(w, err)
			return
		}

		if !usr.UpdateTimestamp.IsZero() {
			err := tx.Commit()
			if err != nil {
				err = errors.HTTPErr{
					Code: http.StatusBadRequest,
					Kind: errors.Database,
					Err:  err,
				}
				errors.HTTPError(w, err)
				return
			}
		} else {
			err = tx.Rollback()
			if err != nil {
				err = errors.HTTPErr{
					Code: http.StatusBadRequest,
					Kind: errors.Database,
					Err:  err,
				}
				errors.HTTPError(w, err)
				return
			}
		}

		// If we successfully committed the db transaction, we can consider this
		// transaction successful and return a response with the response body

		// create new AuditOpts struct and set options to true that you
		// want to see in the response body (Request ID is always present)
		aopt := new(httplog.AuditOpts)
		aopt.Host = true
		aopt.Port = true
		aopt.Path = true
		aopt.Query = true

		// get a new httplog.Audit struct from NewAudit using the
		// above set options and request context
		aud, err := httplog.NewAudit(ctx, aopt)
		if err != nil {
			err = errors.HTTPErr{
				Code: http.StatusInternalServerError,
				Kind: errors.Other,
				Err:  err,
			}
			errors.HTTPError(w, err)
			return
		}

		// create a new response struct and set Audit and other
		// relevant elements
		resp := new(response)
		resp.Audit = aud
		resp.Username = usr.Username
		resp.MobileID = usr.MobileID
		resp.Email = usr.Email
		resp.FirstName = usr.FirstName
		resp.LastName = usr.LastName
		resp.UpdateUsername = usr.UpdateUsername
		resp.UpdateUnixTime = usr.UpdateTimestamp.Unix()

		// Encode response struct to JSON for the response body
		json.NewEncoder(w).Encode(*resp)
		if err != nil {
			err = errors.HTTPErr{
				Code: http.StatusInternalServerError,
				Kind: errors.Other,
				Err:  err,
			}
			errors.HTTPError(w, err)
			return
		}
	}
}
