package app

import (
	"encoding/json"
	"net/http"

	"github.com/gilcrest/go-API-template/lib/errors"
	"github.com/rs/zerolog/log"
)

// Error represents a handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type Error interface {
	error
	Status() int
	ErrKind() string
}

// HTTPErr represents an error with an associated HTTP status code.
type HTTPErr struct {
	Code int
	Kind errors.Kind
	Err  error
}

// Allows HTTPErr to satisfy the error interface.
func (hse HTTPErr) Error() string {
	return hse.Err.Error()
}

// SetErr creates an error type and adds it to the struct
func (hse *HTTPErr) SetErr(s string) {
	hse.Err = errors.Str(s)
}

// ErrKind returns a string denoting the "kind" of error
func (hse HTTPErr) ErrKind() string {
	return hse.Kind.String()
}

// Status Returns an HTTP status code.
func (hse HTTPErr) Status() int {
	return hse.Code
}

type errResponse struct {
	Error svcError `json:"error"`
}

type svcError struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
}

// httpError takes a writer and an error, performs a type switch to
// determine if the type is an HTTPError (which meets the Error interface
// as defined in this package), then sends the Error as a response to the
// client. If the type is not meet the Error interface as defined in this
// package, then a proper error is still formed and sent to the client,
// however, the Kind will be Unanticipated.
func httpError(w http.ResponseWriter, err error) {
	const op errors.Op = "app.httpError"

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
					Kind:    e.ErrKind(),
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
					Kind:    errors.Unanticipated.String(),
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
