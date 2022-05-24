package service

import (
	"strconv"

	"github.com/rs/zerolog"

	"github.com/gilcrest/diy-go-api/domain/errs"
	"github.com/gilcrest/diy-go-api/domain/logger"
)

// LoggerRequest is the request struct for the app logger
type LoggerRequest struct {
	GlobalLogLevel string `json:"global_log_level"`
	LogErrorStack  string `json:"log_error_stack"`
}

// LoggerResponse is the response struct for the current
// state of the app logger
type LoggerResponse struct {
	LoggerMinimumLevel string `json:"logger_minimum_level"`
	GlobalLogLevel     string `json:"global_log_level"`
	LogErrorStack      bool   `json:"log_error_stack"`
}

// LoggerService reads and updates the logger state
type LoggerService struct {
	Logger zerolog.Logger
}

// ReadLogger handles GET requests for the /logger endpoint
func (ls LoggerService) Read() LoggerResponse {

	var logErrorStack bool
	if zerolog.ErrorStackMarshaler != nil {
		logErrorStack = true
	}

	response := LoggerResponse{
		LoggerMinimumLevel: ls.Logger.GetLevel().String(),
		GlobalLogLevel:     zerolog.GlobalLevel().String(),
		LogErrorStack:      logErrorStack,
	}

	return response
}

// Update handles PUT requests for the /logger endpoint
// and updates the logger globals
func (ls LoggerService) Update(r *LoggerRequest) (LoggerResponse, error) {

	if r.GlobalLogLevel != "" {
		// parse input level from request (if present) and set to that
		lvl, err := zerolog.ParseLevel(r.GlobalLogLevel)
		if err != nil {
			return LoggerResponse{}, errs.E(errs.Validation, err)
		}

		clvl := zerolog.GlobalLevel()

		if lvl != clvl {
			// set the global logging level
			zerolog.SetGlobalLevel(lvl)
		}
	}

	if r.LogErrorStack != "" {
		var (
			les bool
			err error
		)
		if les, err = strconv.ParseBool(r.LogErrorStack); err != nil {
			return LoggerResponse{}, errs.E(errs.Validation, "Invalid value sent for log_error_stack")
		}
		// use input LogErrorStack boolean to set whether or not to
		// write error stack
		logger.WriteErrorStackGlobal(les)
	}

	var logErrorStack bool
	if zerolog.ErrorStackMarshaler != nil {
		logErrorStack = true
	}

	response := LoggerResponse{
		LoggerMinimumLevel: ls.Logger.GetLevel().String(),
		GlobalLogLevel:     zerolog.GlobalLevel().String(),
		LogErrorStack:      logErrorStack,
	}

	return response, nil
}
