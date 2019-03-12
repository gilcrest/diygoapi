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

		const dateFormat string = "Jan 02 2006"

		// retrieve the context from the http.Request
		ctx := req.Context()

		// get a new httplog.Audit struct from NewAudit
		aud, err := httplog.NewAudit(ctx)
		if err != nil {
			// log error
			s.Logger.Error().Err(err).Str("RequestID", aud.RequestID).Msg("")
			// response error
			err = errors.RE(http.StatusInternalServerError, errors.Other, err)
			errors.HTTPError(w, err)
			return
		}

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

		// get a new DB Tx from the PostgreSQL datastore within the server struct
		tx, err := s.DS.BeginTx(ctx, nil, datastore.AppDB)
		if err != nil {
			err = errors.RE(http.StatusInternalServerError, errors.Database)
			errors.HTTPError(w, err)
			return
		}

		// Call the create method of the Movie object to validate and insert the data
		err = movie.Create(ctx, s.Logger, tx)
		if err != nil {
			// log error
			s.Logger.Error().Err(err).Str("RequestID", aud.RequestID).Msg("")
			// All errors should be an errors.Error type
			// Use Kind, Code and Error from lower level errors to populate
			// RE (Response Error)
			if e, ok := err.(*errors.Error); ok {
				err := errors.RE(http.StatusBadRequest, e.Kind, e.Param, e.Code, err)
				errors.HTTPError(w, err)
				return
			}

			// if falls through type assertion, then serve an unanticipated error
			err := errors.RE(http.StatusInternalServerError, errors.Unanticipated)
			errors.HTTPError(w, err)
			return
		}

		// If we successfully created/committed the db transaction, we can consider this
		// transaction successful and return a response with the response body

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
