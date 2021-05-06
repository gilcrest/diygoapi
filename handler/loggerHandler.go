package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/logger"
)

// NewReadLoggerHandler is a provider for ReadLoggerHandler
func NewReadLoggerHandler() ReadLoggerHandler {
	return http.HandlerFunc(ReadLogger)
}

// NewUpdateLoggerHandler is a provider for UpdateLoggerHandler
func NewUpdateLoggerHandler() UpdateLoggerHandler {
	return http.HandlerFunc(UpdateLogger)
}

// ReadLogger handles GET requests for the /logger endpoint
func ReadLogger(w http.ResponseWriter, r *http.Request) {
	// readLoggerResponse is the response struct for the current
	// state of the app logger
	type readLoggerResponse struct {
		LoggerMinimumLevel string `json:"logger_minimum_level"`
		GlobalLogLevel     string `json:"global_log_level"`
		LogErrorStack      bool   `json:"log_error_stack"`
	}

	lgr := *hlog.FromRequest(r)

	var logErrorStack bool
	if zerolog.ErrorStackMarshaler != nil {
		logErrorStack = true
	}

	response := readLoggerResponse{
		LoggerMinimumLevel: lgr.GetLevel().String(),
		GlobalLogLevel:     zerolog.GlobalLevel().String(),
		LogErrorStack:      logErrorStack,
	}

	// Encode response struct to JSON for the response body
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, errs.E(errs.Internal, err))
		return
	}
}

// UpdateLogger handles PUT requests for the /logger endpoint
// and updates the given movie
func UpdateLogger(w http.ResponseWriter, r *http.Request) {
	// updateLoggerRequest is the request struct for the app logger
	type updateLoggerRequest struct {
		GlobalLogLevel string `json:"global_log_level"`
		LogErrorStack  string `json:"log_error_stack"`
	}

	// updateLoggerResponse is the response struct for the current
	// state of the app logger
	type updateLoggerResponse struct {
		LoggerMinimumLevel string `json:"logger_minimum_level"`
		GlobalLogLevel     string `json:"global_log_level"`
		LogErrorStack      bool   `json:"log_error_stack"`
	}

	lgr := *hlog.FromRequest(r)

	// Declare rb as an instance of updateLoggerRequest
	rb := new(updateLoggerRequest)

	// Decode JSON HTTP request body into a json.Decoder type
	// and unmarshal that into rb
	err := json.NewDecoder(r.Body).Decode(&rb)
	defer r.Body.Close()
	// Call DecoderErr to determine if body is nil, json is malformed
	// or any other error
	err = DecoderErr(err)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	if rb.GlobalLogLevel != "" {
		// parse input level from request (if present) and set to that
		lvl, err := zerolog.ParseLevel(rb.GlobalLogLevel)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}
		// set the global logging level
		zerolog.SetGlobalLevel(lvl)
	}

	if rb.LogErrorStack != "" {
		var les bool
		if les, err = strconv.ParseBool(rb.LogErrorStack); err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}
		// use input LogErrorStack boolean to set whether or not to
		// write error stack
		logger.WriteErrorStackGlobal(les)
	}

	var logErrorStack bool
	if zerolog.ErrorStackMarshaler != nil {
		logErrorStack = true
	}

	response := updateLoggerResponse{
		LoggerMinimumLevel: lgr.GetLevel().String(),
		GlobalLogLevel:     zerolog.GlobalLevel().String(),
		LogErrorStack:      logErrorStack,
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, errs.E(errs.Internal, err))
		return
	}
}
