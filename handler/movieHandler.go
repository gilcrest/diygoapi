package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gilcrest/go-api-basic/controller/moviecontroller"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/rs/zerolog/hlog"
)

// CreateMovie handles POST requests for the /movies endpoint
// and creates a movie in the database
func (ah *AppHandler) CreateMovie(w http.ResponseWriter, r *http.Request) {
	logger := *hlog.FromRequest(r)

	// Initialize the MovieController
	mc := moviecontroller.NewMovieController(ah.App)

	// Send the request context and request struct to the controller
	// Receive a response or error in return
	response, err := mc.CreateMovie(r)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(*response)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
		return
	}
}

//// UpdateMovie handles PUT requests for the /movies/{id} endpoint
//// and updates the given movie
//func (ah *AppHandler) UpdateMovie(w http.ResponseWriter, r *http.Request) {
//	logger := *hlog.FromRequest(r)
//
//	// retrieve the user from the request Context
//	token, err := accessToken(r)
//	if err != nil {
//		errs.HTTPErrorResponse(w, logger, err)
//		return
//	}
//
//	// gorilla mux Vars function returns the route variables for the
//	// current request, if any. id is the external id given for the
//	// movie
//	vars := mux.Vars(r)
//	id := vars["id"]
//
//	// Declare requestData as an instance of moviecontroller.RequestData
//	requestData := new(moviecontroller.RequestData)
//
//	// Decode JSON HTTP request body into a Decoder type
//	// and unmarshal that into requestData
//	err = json.NewDecoder(r.Body).Decode(&requestData)
//	defer r.Body.Close()
//	// Call DecoderErr to determine if body is nil, json is malformed
//	// or any other error
//	err = DecoderErr(err)
//	if err != nil {
//		errs.HTTPErrorResponse(w, logger, err)
//		return
//	}
//
//	// Initialize the MovieController
//	mc := moviecontroller.NewMovieController(ah.App, ah.StandardResponseFields)
//
//	// Send the request context and request struct to the controller
//	// Receive a response or error in return
//	resp, err := mc.Update(r.Context(), id, requestData, token)
//	if err != nil {
//		errs.HTTPErrorResponse(w, logger, err)
//		return
//	}
//
//	// Encode response struct to JSON for the response body
//	err = json.NewEncoder(w).Encode(resp)
//	if err != nil {
//		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
//		return
//	}
//}
//
//// DeleteMovie handles DELETE requests for the /movies/{id} endpoint
//// and updates the given movie
//func (ah *AppHandler) DeleteMovie(w http.ResponseWriter, r *http.Request) {
//	logger := *hlog.FromRequest(r)
//
//	// retrieve the user from the request Context
//	token, err := accessToken(r)
//	if err != nil {
//		errs.HTTPErrorResponse(w, logger, err)
//		return
//	}
//
//	// gorilla mux Vars function returns the route variables for the
//	// current request, if any. id is the external id given for the
//	// movie
//	vars := mux.Vars(r)
//	id := vars["id"]
//
//	// Initialize the MovieController
//	mc := moviecontroller.NewMovieController(ah.App, ah.StandardResponseFields)
//
//	// Send the request context and request struct to the controller
//	// Receive a response or error in return
//	resp, err := mc.Delete(r.Context(), id, token)
//	if err != nil {
//		errs.HTTPErrorResponse(w, logger, err)
//		return
//	}
//
//	// Encode response struct to JSON for the response body
//	err = json.NewEncoder(w).Encode(resp)
//	if err != nil {
//		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
//		return
//	}
//}
//
//// FindByID handles GET requests for the /movies/{id} endpoint
//// and finds a movie by it's ID
//func (ah *AppHandler) FindByID(w http.ResponseWriter, r *http.Request) {
//	logger := *hlog.FromRequest(r)
//
//	// retrieve the user from the request Context
//	token, err := accessToken(r)
//	if err != nil {
//		errs.HTTPErrorResponse(w, logger, err)
//		return
//	}
//
//	vars := mux.Vars(r)
//	id := vars["id"]
//
//	// Initialize the MovieController
//	mc := moviecontroller.NewMovieController(ah.App, ah.StandardResponseFields)
//
//	// Send the request context and request struct to the controller
//	// Receive a response or error in return
//	resp, err := mc.FindByID(r.Context(), id, token)
//	if err != nil {
//		errs.HTTPErrorResponse(w, logger, err)
//		return
//	}
//
//	// Encode response struct to JSON for the response body
//	err = json.NewEncoder(w).Encode(resp)
//	if err != nil {
//		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
//		return
//	}
//}
//
//// FindAll handles GET requests for the /movies endpoint
//// and finds all movies
//func (ah *AppHandler) FindAll(w http.ResponseWriter, r *http.Request) {
//	logger := *hlog.FromRequest(r)
//
//	// retrieve the user from the request Context
//	token, err := accessToken(r)
//	if err != nil {
//		errs.HTTPErrorResponse(w, logger, err)
//		return
//	}
//
//	// Initialize the MovieController
//	mc := moviecontroller.NewMovieController(ah.App, ah.StandardResponseFields)
//
//	// Send the request context and request struct to the controller
//	// Receive a response or error in return
//	resp, err := mc.FindAll(r.Context(), token)
//	if err != nil {
//		errs.HTTPErrorResponse(w, logger, err)
//		return
//	}
//
//	// Encode response struct to JSON for the response body
//	err = json.NewEncoder(w).Encode(resp)
//	if err != nil {
//		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
//		return
//	}
//}
