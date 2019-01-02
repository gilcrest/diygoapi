package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gilcrest/errors"
	"github.com/gilcrest/httplog"
	"github.com/gilcrest/movie"
	"github.com/gilcrest/srvr/datastore"
)

// handlePost handles POST requests for the /movie endpoint
// and creates a movie in the database
func (s *Server) handlePost() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {

		// request is the expected service request fields
		type request struct {
			Title    string `json:"title"`
			Year     int    `json:"year"`
			Rated    string `json:"rated"`
			Released string `json:"releaseDate"`
			RunTime  int    `json:"runTime"`
			Director string `json:"director"`
			Writer   string `json:"writer"`
		}

		// response is the expected service response fields
		type response struct {
			Title    string         `json:"title"`
			Year     int            `json:"year"`
			Rated    string         `json:"rated"`
			Released string         `json:"releaseDate"`
			RunTime  int            `json:"runTime"`
			Director string         `json:"director"`
			Writer   string         `json:"writer"`
			Audit    *httplog.Audit `json:"audit"`
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
		movie := new(movie.Movie)
		movie.Title = rqst.Title
		movie.Year = rqst.Year
		movie.Rated = rqst.Rated
		movie.Released = time.Date(2018, time.September, 23, 0, 0, 0, 0, time.Local) //rqst.Released
		movie.RunTime = rqst.RunTime
		movie.Director = rqst.Director
		movie.Writer = rqst.Writer

		// Call the create method of the User object to validate data and write to db
		err = movie.Validate()
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
		tx, err = movie.Create(ctx, s.Logger, tx)
		if err != nil {
			err = errors.HTTPErr{
				Code: http.StatusBadRequest,
				Kind: errors.Database,
				Err:  err,
			}
			errors.HTTPError(w, err)
			return
		}

		if !movie.CreateTimestamp.IsZero() {
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
		resp.Title = movie.Title
		resp.Year = movie.Year
		resp.Rated = movie.Rated
		resp.Released = movie.Released.String()
		resp.RunTime = movie.RunTime
		resp.Director = movie.Director
		resp.Writer = movie.Writer

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
