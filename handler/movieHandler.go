package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gilcrest/errs"
	"github.com/gilcrest/go-api-basic/controller/movieController"
	"github.com/gorilla/mux"
)

// AddMovie handles POST requests for the /movies endpoint
// and creates a movie in the database
func (ah *AppHandler) AddMovie(w http.ResponseWriter, r *http.Request) {
	const op errs.Op = "handler/AppHandler.AddMovie"

	// Declare requestData as an instance of movieController.RequestData
	requestData := new(movieController.RequestData)

	// Decode JSON HTTP request body into a Decoder type
	// and unmarshal that into the MovieRequest struct in the
	// AddMovieHandler
	err := json.NewDecoder(r.Body).Decode(&requestData)
	defer r.Body.Close()
	// Call DecoderErr to determine if body is nil, json is malformed
	// or any other error
	err = DecoderErr(err)
	if err != nil {
		err = errs.RE(http.StatusBadRequest, errs.InvalidRequest, errs.E(op, err))
		errs.HTTPError(w, err)
		return
	}

	// retrieve the context from the http.Request
	ctx := r.Context()

	// Initialize the MovieController
	mc := movieController.NewMovieController(ah.App, ah.StandardResponseFields)

	// Send the request context and request struct to the controller
	// Receive a response or error in return
	response, err := mc.Add(ctx, requestData)
	if err != nil {
		err = errs.RE(http.StatusBadRequest, errs.E(op, err))
		errs.HTTPError(w, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(*response)
	if err != nil {
		err = errs.RE(http.StatusInternalServerError, errs.E(op, errs.Internal))
		errs.HTTPError(w, err)
		return
	}
}

// FindByID handles GET requests for the /movies/{id} endpoint
// and finds a movie by it's ID
func (ah *AppHandler) FindByID(w http.ResponseWriter, r *http.Request) {
	const op errs.Op = "handler/AppHandler.FindByID"

	vars := mux.Vars(r)
	id := vars["id"]

	// retrieve the context from the http.Request
	ctx := r.Context()

	// Initialize the MovieController
	mc := movieController.NewMovieController(ah.App, ah.StandardResponseFields)

	// Send the request context and request struct to the controller
	// Receive a response or error in return
	resp, err := mc.FindByID(ctx, id)
	if err != nil {
		err = errs.RE(http.StatusBadRequest, errs.InvalidRequest, errs.E(op, err))
		errs.HTTPError(w, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		err = errs.RE(http.StatusInternalServerError, errs.E(op, errs.Internal))
		errs.HTTPError(w, err)
		return
	}
}

// FindAll handles GET requests for the /movies endpoint
// and finds all movies
func (ah *AppHandler) FindAll(w http.ResponseWriter, r *http.Request) {
	const op errs.Op = "handler/AppHandler.FindAll"

	// retrieve the context from the http.Request
	ctx := r.Context()

	// Initialize the MovieController
	mc := movieController.NewMovieController(ah.App, ah.StandardResponseFields)

	// Send the request context and request struct to the controller
	// Receive a response or error in return
	resp, err := mc.FindAll(ctx)
	if err != nil {
		err = errs.RE(http.StatusBadRequest, errs.InvalidRequest, errs.E(op, err))
		errs.HTTPError(w, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		err = errs.RE(http.StatusInternalServerError, errs.E(op, errs.Internal))
		errs.HTTPError(w, err)
		return
	}
}

// Update handles PUT requests for the /movies/{id} endpoint
// and updates the given movie
func (ah *AppHandler) Update(w http.ResponseWriter, r *http.Request) {
	const op errs.Op = "handler/AppHandler.Update"

	vars := mux.Vars(r)
	id := vars["id"]

	// Declare requestData as an instance of movieController.RequestData
	requestData := new(movieController.RequestData)

	// Decode JSON HTTP request body into a Decoder type
	// and unmarshal that into requestData
	err := json.NewDecoder(r.Body).Decode(&requestData)
	defer r.Body.Close()
	// Call DecoderErr to determine if body is nil, json is malformed
	// or any other error
	err = DecoderErr(err)
	if err != nil {
		err = errs.RE(http.StatusBadRequest, errs.InvalidRequest, errs.E(op, err))
		errs.HTTPError(w, err)
		return
	}

	// retrieve the context from the http.Request
	ctx := r.Context()

	// Initialize the MovieController
	mc := movieController.NewMovieController(ah.App, ah.StandardResponseFields)

	// Send the request context and request struct to the controller
	// Receive a response or error in return
	resp, err := mc.Update(ctx, id, requestData)
	if err != nil {
		err = errs.RE(http.StatusBadRequest, errs.InvalidRequest, errs.E(op, err))
		errs.HTTPError(w, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		err = errs.RE(http.StatusInternalServerError, errs.E(op, errs.Internal))
		errs.HTTPError(w, err)
		return
	}
}

// Delete handles DELETE requests for the /movies/{id} endpoint
// and updates the given movie
func (ah *AppHandler) Delete(w http.ResponseWriter, r *http.Request) {
	const op errs.Op = "handler/AppHandler.Delete"

	vars := mux.Vars(r)
	id := vars["id"]

	// retrieve the context from the http.Request
	ctx := r.Context()

	// Initialize the MovieController
	mc := movieController.NewMovieController(ah.App, ah.StandardResponseFields)

	// Send the request context and request struct to the controller
	// Receive a response or error in return
	resp, err := mc.Delete(ctx, id)
	if err != nil {
		err = errs.RE(http.StatusBadRequest, errs.InvalidRequest, errs.E(op, err))
		errs.HTTPError(w, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(resp)
	if err != nil {
		err = errs.RE(http.StatusInternalServerError, errs.E(op, errs.Internal))
		errs.HTTPError(w, err)
		return
	}
}
