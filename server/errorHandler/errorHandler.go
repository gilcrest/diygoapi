package errorHandler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gilcrest/go-API-template/env"
)

// Error represents a handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type Error interface {
	error
	Status() int
	ErrType() string
}

// HTTPErr represents an error with an associated HTTP status code.
type HTTPErr struct {
	Code int
	Type string
	Err  error
}

// Allows HTTPErr to satisfy the error interface.
func (hse HTTPErr) Error() string {
	return hse.Err.Error()
}

// SetErr creates an error type and adds it to the struct
func (hse *HTTPErr) SetErr(s string) {
	hse.Err = errors.New(s)
}

// ErrType returns a string error type/code
func (hse HTTPErr) ErrType() string {
	return hse.Type
}

// Status Returns an HTTP status code.
func (hse HTTPErr) Status() int {
	return hse.Code
}

type errResponse struct {
	Error svcError `json:"error"`
}

type svcError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// The ErrHandler struct that takes a configured Env and a function matching
// our useful signature.
type ErrHandler struct {
	Env *env.Env
	H   func(e *env.Env, w http.ResponseWriter, r *http.Request) error
}

// ServeHTTP allows Handler type to satisfy the http.Handler interface
func (h ErrHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Get a new logger instance
	log := h.Env.Logger

	log.Debug().Msg("Start Handler.ServeHTTP")
	defer log.Debug().Msg("Finish Handler.ServeHTTP")

	err := h.H(h.Env, w, r)

	if err != nil {
		// We perform a "type switch" https://tour.golang.org/methods/16
		// to determine the interface value type
		switch e := err.(type) {
		// If the interface value is of type Error (not a typical error, but
		// the Error interface defined above), then
		case Error:
			// We can retrieve the status here and write out a specific
			// HTTP status code.
			log.Printf("HTTP %d - %s", e.Status(), e)

			er := errResponse{
				Error: svcError{
					Type:    e.ErrType(),
					Message: e.Error(),
				},
			}

			// Marshal errResponse struct to JSON for the response body
			errJSON, _ := json.MarshalIndent(er, "", "    ")

			http.Error(w, string(errJSON), e.Status())

		default:
			// Any error types we don't specifically look out for default
			// to serving a HTTP 500
			cd := http.StatusInternalServerError
			er := errResponse{
				Error: svcError{
					Type:    "unknown_error",
					Message: "Unexpected error - contact support",
				},
			}

			log.Error().Msgf("Unknown Error - HTTP %d - %s", cd, err.Error())

			// Marshal errResponse struct to JSON for the response body
			errJSON, _ := json.MarshalIndent(er, "", "    ")

			http.Error(w, string(errJSON), cd)
		}
	}

}
