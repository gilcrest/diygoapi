package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gilcrest/go-api-basic/controller/moviectl"
	"github.com/gilcrest/go-api-basic/domain/errs"
)

// AddMovie handles POST requests for the /movie endpoint
// and creates a movie in the database
func (ah *AppHandler) AddMovie() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const op errs.Op = "app/Application.AddMovie"

		// Declare rqst as an instance of moviectl.AddMovieRequest
		rqst := new(moviectl.AddMovieRequest)
		// Decode JSON HTTP request body into a Decoder type
		// and unmarshal that into rqst
		err := json.NewDecoder(req.Body).Decode(&rqst)
		defer req.Body.Close()
		if err != nil {
			err = errs.RE(http.StatusBadRequest, errs.InvalidRequest, errs.E(op, err))
			errs.HTTPError(w, err)
			return
		}

		// retrieve the context from the http.Request
		ctx := req.Context()

		// Initialize the MovieController
		mc := moviectl.ProvideMovieController(ah.App)

		// Send the request context and request struct to the controller
		// Receive a response or error in return
		resp, err := mc.Add(ctx, rqst)
		if err != nil {
			err = errs.RE(http.StatusBadRequest, errs.InvalidRequest, err)
			errs.HTTPError(w, err)
			return
		}

		// Encode response struct to JSON for the response body
		json.NewEncoder(w).Encode(*resp)
		if err != nil {
			err = errs.RE(http.StatusInternalServerError, errs.Internal)
			errs.HTTPError(w, err)
			return
		}
	}
}
