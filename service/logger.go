package service

import (
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"

	"github.com/gilcrest/diygoapi"
	"github.com/gilcrest/diygoapi/errs"
)

// LoggerService reads and updates the logger state
type LoggerService struct {
	Logger zerolog.Logger
}

// ReadLogger handles GET requests for the /logger endpoint
func (ls *LoggerService) Read() *diygoapi.LoggerResponse {

	var logErrorStack bool
	if zerolog.ErrorStackMarshaler != nil {
		logErrorStack = true
	}

	response := &diygoapi.LoggerResponse{
		LoggerMinimumLevel: ls.Logger.GetLevel().String(),
		GlobalLogLevel:     zerolog.GlobalLevel().String(),
		LogErrorStack:      logErrorStack,
	}

	return response
}

// Update handles PUT requests for the /logger endpoint
// and updates the logger globals
func (ls *LoggerService) Update(r *diygoapi.LoggerRequest) (*diygoapi.LoggerResponse, error) {

	if r.GlobalLogLevel != "" {
		// parse input level from request (if present) and set to that
		lvl, err := zerolog.ParseLevel(r.GlobalLogLevel)
		if err != nil {
			return nil, errs.E(errs.Validation, err)
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
			return nil, errs.E(errs.Validation, "Invalid value sent for log_error_stack")
		}
		// use input LogErrorStack boolean to determine whether to write error stack
		writeErrorStackGlobal(les)
	}

	var logErrorStack bool
	if zerolog.ErrorStackMarshaler != nil {
		logErrorStack = true
	}

	response := &diygoapi.LoggerResponse{
		LoggerMinimumLevel: ls.Logger.GetLevel().String(),
		GlobalLogLevel:     zerolog.GlobalLevel().String(),
		LogErrorStack:      logErrorStack,
	}

	return response, nil
}

// writeErrorStackGlobal is a convenience wrapper to set the zerolog
// Global variable ErrorStackMarshaler to write Error stacks for logs
func writeErrorStackGlobal(writeStack bool) {
	if !writeStack {
		zerolog.ErrorStackMarshaler = nil
		return
	}
	// set ErrorStackMarshaler to pkgerrors.MarshalStack
	// to enable error stack traces
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
}
