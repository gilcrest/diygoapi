package errs

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/rs/zerolog/log"
)

// hError represents an HTTP handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type hError interface {
	error
	Status() int
	ErrKind() string
	ErrParam() string
	ErrCode() string
	StatusOnly() bool
}

// HTTPErr represents an error with an associated HTTP status code.
type HTTPErr struct {
	HTTPStatusCode int
	Kind           Kind
	Param          Parameter
	Code           Code
	Err            error
}

// Status Returns an HTTP Status Code.
func (hse HTTPErr) Status() int {
	return hse.HTTPStatusCode
}

// ErrKind returns a string denoting the "kind" of error
func (hse HTTPErr) ErrKind() string {
	if hse.Kind == 0 {
		return ""
	}
	return hse.Kind.String()
}

// ErrParam returns a string denoting the "kind" of error
func (hse HTTPErr) ErrParam() string {
	return string(hse.Param)
}

// ErrCode returns a string denoting the "kind" of error
func (hse HTTPErr) ErrCode() string {
	return string(hse.Code)
}

// StatusOnly determines if the only field populated is the HTTP Status Code
// If so, the error response body should not be populated
func (hse *HTTPErr) StatusOnly() bool {
	return hse.HTTPStatusCode != 0 && hse.Kind == 0 && hse.Param == "" && hse.Code == "" && hse.Err == nil
}

// SetErr creates an error type and adds it to the struct
func (hse *HTTPErr) SetErr(s string) {
	hse.Err = errors.New(s)
}

// Allows HTTPErr to satisfy the error interface.
func (hse HTTPErr) Error() string {
	// In case user forgets to add an error type to HTTPErr
	if hse.Err == nil {
		return ""
	}
	return hse.Err.Error()
}

// ErrResponse is used as the Response Body
type ErrResponse struct {
	Error ServiceError `json:"error"`
}

// ServiceError has fields for Service errors. All fields with no data will
// be omitted
type ServiceError struct {
	Kind    string `json:"kind,omitempty"`
	Code    string `json:"code,omitempty"`
	Param   string `json:"param,omitempty"`
	Message string `json:"message,omitempty"`
}

// HTTPError takes a writer and an error, performs a type switch to
// determine if the type is an HTTPError (which meets the Error interface
// as defined in this package), then sends the Error as a response to the
// client. If the type does not meet the Error interface as defined in this
// package, then a proper error is still formed and sent to the client,
// however, the Kind and Code will be Unanticipated.
func HTTPError(w http.ResponseWriter, err error) {
	const op Op = "errors.httpError"

	if err != nil {
		// We perform a "type switch" https://tour.golang.org/methods/16
		// to determine the interface value type
		switch e := err.(type) {
		// If the interface value is of type Error (not a typical error, but
		// the Error interface defined above), then
		case hError:
			// We can retrieve the status here and write out a specific
			// HTTP status code.
			if e.StatusOnly() {
				log.Error().Int("HTTP Error StatusCode", e.Status()).Msg("")
			} else {
				log.Error().Msgf("HTTP %d - %s", e.Status(), e)
			}
			if e.StatusOnly() {
				sendError(w, "", e.Status())
			} else {
				er := ErrResponse{
					Error: ServiceError{
						Kind:    e.ErrKind(),
						Code:    e.ErrCode(),
						Param:   e.ErrParam(),
						Message: e.Error(),
					},
				}

				// Marshal errResponse struct to JSON for the response body
				errJSON, _ := json.MarshalIndent(er, "", "    ")

				sendError(w, string(errJSON), e.Status())
			}

		default:
			// Any error types we don't specifically look out for default
			// to serving a HTTP 500
			cd := http.StatusInternalServerError
			er := ErrResponse{
				Error: ServiceError{
					Kind:    Unanticipated.String(),
					Code:    "Unanticipated",
					Message: "Unexpected error - contact support",
				},
			}

			log.Error().Msgf("Unknown Error - HTTP %d - %s", cd, err.Error())

			// Marshal errResponse struct to JSON for the response body
			errJSON, _ := json.MarshalIndent(er, "", "    ")

			sendError(w, string(errJSON), cd)
		}
	}
}

// Taken from standard library, but changed to send application/json as header
// Error replies to the request with the specified error message and HTTP code.
// It does not otherwise end the request; the caller should ensure no further
// writes are done to w.
// The error message should be json.
func sendError(w http.ResponseWriter, error string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(statusCode)
	// Only write response body if there is an error string populated
	if error != "" {
		fmt.Fprintln(w, error)
	}
}

// RE builds an HTTP Response error value from its arguments.
// There must be at least one argument or RE panics.
// The type of each argument determines its meaning.
// If more than one argument of a given type is presented,
// only the last one is recorded.
//
// The types are:
func RE(args ...interface{}) error {
	if len(args) == 0 {
		panic("call to errors.RE with no arguments")
	}

	fullErr := new(Error)

	e := &HTTPErr{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case int:
			e.HTTPStatusCode = arg
		case Kind:
			e.Kind = arg
		case string:
			e.Code = Code(arg)
		case Code:
			e.Code = arg
		case Parameter:
			e.Param = arg
		case *Error:
			// capture original error for logging before stripping
			fullErr = arg

			// For API response errors, don't show full recursion details,
			// just the error message
			e.Err = StripStack(arg)
		case error:
			e.Err = arg
		default:
			_, file, line, _ := runtime.Caller(1)
			log.Error().Msgf("errors.E: bad call from %s:%d: %v", file, line, args)
			return fmt.Errorf("unknown type %T, value %v in error call", arg, arg)
		}
	}

	log.Error().Err(fullErr).Msg("Error Response")

	return e
}

// StripStack takes an Error type (Error defined in this module) and
// removes the leading stack information
func StripStack(e error) error {
	err, ok := e.(*Error)
	if ok {
		// get error string
		errStr := err.Error()
		// get position where |: character lands in string
		idx := strings.Index(errStr, "|:")
		// substring from after the |: character
		substring := errStr[idx+3:]
		// put substring back into error
		return errors.New(substring)
	}
	// If it's not an Error type, don't strip anything
	return e
}
