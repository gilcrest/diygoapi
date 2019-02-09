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
			Title    string `json:"Title"`
			Year     int    `json:"Year"`
			Rated    string `json:"Rated"`
			Released string `json:"ReleaseDate"`
			RunTime  int    `json:"RunTime"`
			Director string `json:"Director"`
			Writer   string `json:"Writer"`
		}

		// response is the expected service response fields
		type response struct {
			Title           string         `json:"Title"`
			Year            int            `json:"Year"`
			Rated           string         `json:"Rated"`
			Released        string         `json:"ReleaseDate"`
			RunTime         int            `json:"RunTime"`
			Director        string         `json:"Director"`
			Writer          string         `json:"Writer"`
			CreateTimestamp string         `json:"CreateTimestamp"`
			Audit           *httplog.Audit `json:"audit"`
		}

		// retrieve the context from the http.Request
		ctx := req.Context()

		var err error
		const dateFormat string = "Jan 02 2006"

		// Declare rqst as an instance of request
		// Decode JSON HTTP request body into a Decoder type
		//  and unmarshal that into rqst
		rqst := new(request)
		err = json.NewDecoder(req.Body).Decode(&rqst)
		defer req.Body.Close()
		if err != nil {
			err = errors.RE(http.StatusBadRequest, errors.InvalidRequest, err)
			errors.HTTPError(w, err)
			return
		}

		// declare a new instance of usr.User
		movie := new(movie.Movie)
		movie.Title = rqst.Title
		movie.Year = rqst.Year
		movie.Rated = rqst.Rated
		t, err := time.Parse(dateFormat, rqst.Released)
		if err != nil {
			err = errors.RE(http.StatusBadRequest,
				errors.Validation,
				errors.Code("invalid_date_format"),
				errors.Parameter("ReleaseDate"),
				err)
			errors.HTTPError(w, err)
			return
		}
		movie.Released = t
		movie.RunTime = rqst.RunTime
		movie.Director = rqst.Director
		movie.Writer = rqst.Writer

		// Call the Validate method of the movie object
		// to validate request input data
		err = movie.Validate()
		if err != nil {
			// Do a type assertion on the Error coming back, should be
			// an errors.Error - use Kind, Code and Parameter coming
			// up from the error
			if e, ok := err.(*errors.Error); ok {
				err := errors.RE(http.StatusBadRequest, e.Kind, e.Code, e.Param, err)
				errors.HTTPError(w, err)
				return
			}
			// if falls through type assertion, then serve an unanticipated error
			err := errors.RE(http.StatusInternalServerError, errors.Unanticipated)
			errors.HTTPError(w, err)
			return
		}

		// get a new DB Tx
		tx, err := s.DS.BeginTx(ctx, nil, datastore.AppDB)
		if err != nil {
			err = errors.RE(http.StatusInternalServerError, errors.Database)
			errors.HTTPError(w, err)
			return
		}

		// Call the create method of the User object to write
		// to the database
		err = movie.CreateDB(ctx, s.Logger, tx)
		if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				err = errors.RE(http.StatusInternalServerError, errors.Database)
				errors.HTTPError(w, err)
				return
			}
			// Do a type assertion on the Error coming back, if an errors.Error
			// then use code coming up from the error
			if e, ok := err.(*errors.Error); ok {
				err := errors.RE(http.StatusBadRequest, errors.Validation, e.Code, err)
				errors.HTTPError(w, err)
				return
			}
			err = errors.RE(http.StatusInternalServerError, errors.Database)
			errors.HTTPError(w, err)
			return
		}

		if err := tx.Commit(); err != nil {
			err = errors.RE(http.StatusInternalServerError, errors.Database)
			errors.HTTPError(w, err)
			return
		}

		// If we successfully committed the db transaction, we can consider this
		// transaction successful and return a response with the response body

		// get a new httplog.Audit struct from NewAudit using the
		// above set options and request context
		aud, err := httplog.NewAudit(ctx)
		if err != nil {
			err = errors.RE(http.StatusInternalServerError, errors.Unanticipated)
			errors.HTTPError(w, err)
			return
		}

		// create a new response struct and set Audit and other
		// relevant elements
		resp := new(response)
		resp.Title = movie.Title
		resp.Year = movie.Year
		resp.Rated = movie.Rated
		resp.Released = movie.Released.Format(dateFormat)
		resp.RunTime = movie.RunTime
		resp.Director = movie.Director
		resp.Writer = movie.Writer
		resp.CreateTimestamp = movie.CreateTimestamp.Format(time.RFC3339)
		resp.Audit = aud

		// Encode response struct to JSON for the response body
		json.NewEncoder(w).Encode(*resp)
		if err != nil {
			err = errors.RE(http.StatusInternalServerError, errors.Internal)
			errors.HTTPError(w, err)
			return
		}
	}
}
