package errs

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"
)

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

// HTTPErrorResponse takes a writer, error and a logger, performs a
// type switch to determine if the type is an Error (which meets
// the Error interface as defined in this package), then sends the
// Error as a response to the client. If the type does not meet the
// Error interface as defined in this package, then a proper error
// is still formed and sent to the client, however, the Kind and
// Code will be Unanticipated. Logging of error is also done using
// https://github.com/rs/zerolog
func HTTPErrorResponse(w http.ResponseWriter, lgr zerolog.Logger, err error) {
	if err == nil {
		nilErrorResponse(w, lgr)
		return
	}

	var unauthenticatedErr *UnauthenticatedError
	if errors.As(err, &unauthenticatedErr) {
		unauthenticatedErrorResponse(w, lgr, unauthenticatedErr)
		return
	}

	var unauthorizedErr *UnauthorizedError
	if errors.As(err, &unauthorizedErr) {
		unauthorizedErrorResponse(w, lgr, unauthorizedErr)
		return
	}

	var typicalErr *Error
	if errors.As(err, &typicalErr) {
		typicalErrorResponse(w, lgr, typicalErr)
		return
	}

	otherErrorResponse(w, lgr, err)
}

// typicalErrorResponse replies to the request with the specified error
// message and HTTP code. It does not otherwise end the request; the
// caller should ensure no further writes are done to w.
//
// Taken from standard library and modified.
// https://golang.org/pkg/net/http/#Error
func typicalErrorResponse(w http.ResponseWriter, lgr zerolog.Logger, e *Error) {

	httpStatusCode := httpErrorStatusCode(e.Kind)

	// We can retrieve the status here and write out a specific
	// HTTP status code. If the error is empty, just send the HTTP
	// Status Code as response. Error should not be empty, but it's
	// theoretically possible, so this is just in case...
	if e.isZero() {
		lgr.Error().Stack().Int("http_statuscode", httpStatusCode).Msg("empty error")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// typical errors

	// log the error with stacktrace
	lgr.Error().Stack().Err(e.Err).
		Int("http_statuscode", httpStatusCode).
		Str("Kind", e.Kind.String()).
		Str("Parameter", string(e.Param)).
		Str("Code", string(e.Code)).
		Msg("Error Response Sent")

	// get ErrResponse
	er := newErrResponse(e)

	// Marshal errResponse struct to JSON for the response body
	errJSON, _ := json.Marshal(er)
	ejson := string(errJSON)

	// Write Content-Type headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Write HTTP Statuscode
	w.WriteHeader(httpStatusCode)

	// Write response body (json)
	fmt.Fprintln(w, ejson)
}

func newErrResponse(err *Error) ErrResponse {
	const msg string = "internal server error - please contact support"

	switch err.Kind {
	case Internal, Database:
		return ErrResponse{
			Error: ServiceError{
				Kind:    Internal.String(),
				Message: msg,
			},
		}
	default:
		return ErrResponse{
			Error: ServiceError{
				Kind:    err.Kind.String(),
				Code:    string(err.Code),
				Param:   string(err.Param),
				Message: err.Error(),
			},
		}
	}
}

// unauthenticatedErrorResponse responds with an http status code set
// to 401 (Unauthorized / Unauthenticated), an empty response body and
// a WWW-Authenticate header.
func unauthenticatedErrorResponse(w http.ResponseWriter, lgr zerolog.Logger, err *UnauthenticatedError) {
	lgr.Error().Stack().Err(err.Err).
		Int("http_statuscode", http.StatusUnauthorized).
		Str("realm", string(err.Realm())).
		Msg("Unauthenticated Request")

	w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s"`, err.Realm()))
	w.WriteHeader(http.StatusUnauthorized)
}

// unauthorizedErrorResponse responds with an http status code set to 403 (Forbidden)
// and an empty response body
func unauthorizedErrorResponse(w http.ResponseWriter, lgr zerolog.Logger, err *UnauthorizedError) {
	lgr.Error().Stack().Err(err.Err).
		Int("http_statuscode", http.StatusForbidden).
		Msg("Unauthorized Request")

	w.WriteHeader(http.StatusForbidden)
}

// nilErrorResponse responds with an http status code set to 500 (Internal Server Error)
// and an empty response body. nil error should never be sent, but in case it is...
func nilErrorResponse(w http.ResponseWriter, lgr zerolog.Logger) {
	lgr.Error().Stack().
		Int("HTTP Error StatusCode", http.StatusInternalServerError).
		Msg("nil error - no response body sent")

	w.WriteHeader(http.StatusInternalServerError)
}

// otherErrorResponse responds with an http status code set to 500 (Internal Server Error)
// and a json response body with unanticipated_error kind
func otherErrorResponse(w http.ResponseWriter, lgr zerolog.Logger, err error) {
	er := ErrResponse{
		Error: ServiceError{
			Kind:    Unanticipated.String(),
			Code:    "Unanticipated",
			Message: "Unexpected error - contact support",
		},
	}

	lgr.Error().Err(err).Msg("Unknown Error")

	// Marshal errResponse struct to JSON for the response body
	errJSON, _ := json.Marshal(er)
	ejson := string(errJSON)

	// Write Content-Type headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Write HTTP Statuscode
	w.WriteHeader(http.StatusInternalServerError)

	// Write response body (json)
	fmt.Fprintln(w, ejson)
}

// httpErrorStatusCode maps an error Kind to an HTTP Status Code
func httpErrorStatusCode(k Kind) int {
	switch k {
	case Invalid, Exist, NotExist, Private, BrokenLink, Validation, InvalidRequest:
		return http.StatusBadRequest
	// the zero value of Kind is Other, so if no Kind is present
	// in the error, Other is used. Errors should always have a
	// Kind set, otherwise, a 500 will be returned and no
	// error message will be sent to the caller
	case Other, IO, Internal, Database, Unanticipated:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// NewUnauthenticatedError is an initializer for UnauthenticatedError
func NewUnauthenticatedError(realm string, err error) *UnauthenticatedError {
	return &UnauthenticatedError{WWWAuthenticateRealm: realm, Err: err}
}

// UnauthenticatedError implements the error interface and is used
// when a request lacks valid authentication credentials.
//
// For Unauthenticated and Unauthorized errors, the response body
// should be empty. Use logger to log the error and then just send
// http.StatusUnauthorized (401).
//
// From stack overflow - https://stackoverflow.com/questions/3297048/403-forbidden-vs-401-unauthorized-http-responses
// "In summary, a 401 Unauthorized response should be used for missing or bad
// authentication, and a 403 Forbidden response should be used afterwards, when
// the user is authenticated but isn’t authorized to perform the requested
// operation on the given resource."
type UnauthenticatedError struct {
	// WWWAuthenticateRealm is a description of the protected area.
	// If no realm is specified, "DefaultRealm" will be used as realm
	WWWAuthenticateRealm string

	// The underlying error that triggered this one, if any.
	Err error
}

// Unwrap method allows for unwrapping errors using errors.As
func (e UnauthenticatedError) Unwrap() error {
	return e.Err
}

func (e UnauthenticatedError) Error() string {
	return e.Err.Error()
}

// Realm returns the WWWAuthenticateRealm of the error, if empty,
// Realm returns "DefaultRealm"
func (e UnauthenticatedError) Realm() string {
	realm := e.WWWAuthenticateRealm
	if realm == "" {
		realm = "DefaultRealm"
	}

	return realm
}

// NewUnauthorizedError is an initializer for UnauthorizedError
func NewUnauthorizedError(err error) *UnauthorizedError {
	return &UnauthorizedError{Err: err}
}

// UnauthorizedError implements the error interface and is used
// when a user is authenticated, but is not authorized to access the
// resource.
//
// For Unauthenticated and Unauthorized errors, the response body
// should be empty. Use logger to log the error and then just send
// http.StatusUnauthorized (401).
//
// From stack overflow - https://stackoverflow.com/questions/3297048/403-forbidden-vs-401-unauthorized-http-responses
// "In summary, a 401 Unauthorized response should be used for missing or bad
// authentication, and a 403 Forbidden response should be used afterwards, when
// the user is authenticated but isn’t authorized to perform the requested
// operation on the given resource."
type UnauthorizedError struct {
	// The underlying error that triggered this one, if any.
	Err error
}

// Unwrap method allows for unwrapping errors using errors.As
func (e UnauthorizedError) Unwrap() error {
	return e.Err
}

func (e UnauthorizedError) Error() string {
	return e.Err.Error()
}

// MatchUnauthenticated compares its two error arguments. It can be
// used to check for expected errors in tests. Both arguments must
// have underlying type *UnauthenticatedError or MatchUnauthenticated
// will return false. Otherwise it returns true
// if every non-zero element of the first error is equal to the
// corresponding element of the second.
// If the Err field is a *UnauthenticatedError, MatchUnauthenticated
// recurs on that field; otherwise it compares the strings returned
// by the Error methods. Elements that are in the second argument but
// not present in the first are ignored.
func MatchUnauthenticated(err1, err2 error) bool {
	e1, ok := err1.(*UnauthenticatedError)
	if !ok {
		return false
	}
	e2, ok := err2.(*UnauthenticatedError)
	if !ok {
		return false
	}
	if e1.WWWAuthenticateRealm != "" && e2.WWWAuthenticateRealm != e1.WWWAuthenticateRealm {
		return false
	}
	if e1.Err != nil {
		if _, ok := e1.Err.(*UnauthorizedError); ok {
			return MatchUnauthenticated(e1.Err, e2.Err)
		}
		if e2.Err == nil || e2.Err.Error() != e1.Err.Error() {
			return false
		}
	}
	return true
}
