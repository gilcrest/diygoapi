package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gilcrest/go-api-basic/controller/moviectl"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gorilla/mux"
)

// AddMovie handles POST requests for the /movies endpoint
// and creates a movie in the database
func (ah *AppHandler) AddMovie() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op errs.Op = "handler/AppHandler.AddMovie"

		// Declare rqst as an instance of moviectl.AddMovieRequest
		rqst := new(moviectl.MovieRequest)
		// Decode JSON HTTP request body into a Decoder type
		// and unmarshal that into rqst
		err := json.NewDecoder(r.Body).Decode(&rqst)
		defer r.Body.Close()
		if err != nil {
			err = errs.RE(http.StatusBadRequest, errs.InvalidRequest, errs.E(op, err))
			errs.HTTPError(w, err)
			return
		}

		// retrieve the context from the http.Request
		ctx := r.Context()

		// Initialize the MovieController
		mc := moviectl.ProvideMovieController(ah.App, ah.StandardResponseFields)

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

// FindByID handles GET requests for the /movies/{id} endpoint
// and finds a movie by it's ID
func (ah *AppHandler) FindByID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op errs.Op = "handler/AppHandler.FindByID"

		vars := mux.Vars(r)
		id := vars["id"]

		// retrieve the context from the http.Request
		ctx := r.Context()

		// Initialize the MovieController
		mc := moviectl.ProvideMovieController(ah.App, ah.StandardResponseFields)

		// Send the request context and request struct to the controller
		// Receive a response or error in return
		resp, err := mc.FindByID(ctx, id)
		if err != nil {
			err = errs.RE(http.StatusBadRequest, errs.InvalidRequest, err)
			errs.HTTPError(w, err)
			return
		}

		// Encode response struct to JSON for the response body
		json.NewEncoder(w).Encode(resp)
		if err != nil {
			err = errs.RE(http.StatusInternalServerError, errs.Internal)
			errs.HTTPError(w, err)
			return
		}
	}
}

// FindAll handles GET requests for the /movies endpoint
// and finds all movies
func (ah *AppHandler) FindAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op errs.Op = "handler/AppHandler.FindByID"

		// retrieve the context from the http.Request
		ctx := r.Context()

		// Initialize the MovieController
		mc := moviectl.ProvideMovieController(ah.App, ah.StandardResponseFields)

		// Send the request context and request struct to the controller
		// Receive a response or error in return
		resp, err := mc.FindAll(ctx, r)
		if err != nil {
			err = errs.RE(http.StatusBadRequest, errs.InvalidRequest, err)
			errs.HTTPError(w, err)
			return
		}

		// Encode response struct to JSON for the response body
		json.NewEncoder(w).Encode(resp)
		if err != nil {
			err = errs.RE(http.StatusInternalServerError, errs.Internal)
			errs.HTTPError(w, err)
			return
		}
	}
}
