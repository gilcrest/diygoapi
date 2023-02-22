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

	var e *Error
	if errors.As(err, &e) {
		switch e.Kind {
		case Unauthenticated:
			unauthenticatedErrorResponse(w, lgr, e)
			return
		case Unauthorized:
			unauthorizedErrorResponse(w, lgr, e)
			return
		default:
			typicalErrorResponse(w, lgr, e)
			return
		}
	}

	unknownErrorResponse(w, lgr, err)
}

// typicalErrorResponse replies to the request with the specified error
// message and HTTP code. It does not otherwise end the request; the
// caller should ensure no further writes are done to w.
//
// Taken from standard library and modified.
// https://golang.org/pkg/net/http/#Error
func typicalErrorResponse(w http.ResponseWriter, lgr zerolog.Logger, e *Error) {
	const op Op = "errs/typicalErrorResponse"

	httpStatusCode := httpErrorStatusCode(e.Kind)

	// We can retrieve the status here and write out a specific
	// HTTP status code. If the error is empty, just send the HTTP
	// Status Code as response. Error should not be empty, but it's
	// theoretically possible, so this is just in case...
	if e.isZero() {
		lgr.Error().Stack().Msgf("error sent to %s, but empty - very strange, investigate", op)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// typical errors
	const errMsg = "error response sent to client"

	if zerolog.ErrorStackMarshaler != nil {
		err := TopError(e)

		// log the error with stacktrace from "github.com/pkg/errors"
		// do not bother to log with op stack
		lgr.Error().Stack().Err(err).
			Int("http_statuscode", httpStatusCode).
			Str("Kind", e.Kind.String()).
			Str("Parameter", string(e.Param)).
			Str("Code", string(e.Code)).
			Msg(errMsg)
	} else {
		ops := OpStack(e)
		if len(ops) > 0 {
			j, _ := json.Marshal(ops)
			// log the error with the op stack
			lgr.Error().RawJSON("stack", j).Err(e.Err).
				Int("http_statuscode", httpStatusCode).
				Str("Kind", e.Kind.String()).
				Str("Parameter", string(e.Param)).
				Str("Code", string(e.Code)).
				Msg(errMsg)
		} else {
			// no op stack present, log the error without that field
			lgr.Error().Err(e.Err).
				Int("http_statuscode", httpStatusCode).
				Str("Kind", e.Kind.String()).
				Str("Parameter", string(e.Param)).
				Str("Code", string(e.Code)).
				Msg(errMsg)
		}
	}

	// get ErrResponse
	er := newErrResponse(e)

	// Marshal errResponse struct to JSON for the response body
	errJSON, _ := json.Marshal(er)
	ej := string(errJSON)

	// Write Content-Type headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Write HTTP Statuscode
	w.WriteHeader(httpStatusCode)

	// Write response body (json)
	fmt.Fprintln(w, ej)
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

// unauthenticatedErrorResponse responds with http status code 401
// (Unauthorized / Unauthenticated), an empty response body and a
// WWW-Authenticate header.
func unauthenticatedErrorResponse(w http.ResponseWriter, lgr zerolog.Logger, e *Error) {
	if e.Realm == "" {
		e.Realm = "default"
	}

	if zerolog.ErrorStackMarshaler != nil {
		err := TopError(e)

		// log the error with stacktrace from "github.com/pkg/errors"
		// do not bother to log with op stack
		lgr.Error().Stack().Err(err).
			Int("http_statuscode", http.StatusUnauthorized).
			Str("realm", string(e.Realm)).
			Msg("Unauthenticated Request")
	} else {
		ops := OpStack(e)
		if len(ops) > 0 {
			j, _ := json.Marshal(ops)
			// log the error with the op stack
			lgr.Error().RawJSON("stack", j).Err(e.Err).
				Int("http_statuscode", http.StatusUnauthorized).
				Str("realm", string(e.Realm)).
				Msg("Unauthenticated Request")
		} else {
			// no op stack present, log the error without that field
			lgr.Error().Err(e.Err).
				Int("http_statuscode", http.StatusUnauthorized).
				Str("realm", string(e.Realm)).
				Msg("Unauthenticated Request")
		}
	}

	w.Header().Set("WWW-Authenticate", fmt.Sprintf(`Bearer realm="%s"`, e.Realm))
	w.WriteHeader(http.StatusUnauthorized)
}

// unauthorizedErrorResponse responds with http status code 403 (Forbidden)
// and an empty response body.
func unauthorizedErrorResponse(w http.ResponseWriter, lgr zerolog.Logger, e *Error) {

	if zerolog.ErrorStackMarshaler != nil {
		err := TopError(e)

		// log the error with stacktrace from "github.com/pkg/errors"
		// do not bother to log with op stack
		lgr.Error().Stack().Err(err).
			Int("http_statuscode", http.StatusForbidden).
			Msg("Unauthorized Request")
	} else {
		ops := OpStack(e)
		if len(ops) > 0 {
			j, _ := json.Marshal(ops)
			// log the error with the op stack
			lgr.Error().RawJSON("stack", j).Err(e.Err).
				Int("http_statuscode", http.StatusForbidden).
				Msg("Unauthorized Request")
		} else {
			// no op stack present, log the error without that field
			lgr.Error().Err(e.Err).
				Int("http_statuscode", http.StatusForbidden).
				Msg("Unauthorized Request")
		}
	}

	w.WriteHeader(http.StatusForbidden)
}

// nilErrorResponse responds with http status code 500 (Internal Server Error)
// and an empty response body. nil error should never be sent, but in case it is...
func nilErrorResponse(w http.ResponseWriter, lgr zerolog.Logger) {
	lgr.Error().Stack().
		Int("HTTP Error StatusCode", http.StatusInternalServerError).
		Msg("nil error - no response body sent")

	w.WriteHeader(http.StatusInternalServerError)
}

// unknownErrorResponse responds with http status code 500 (Internal Server Error)
// and a json response body with unanticipated_error kind
func unknownErrorResponse(w http.ResponseWriter, lgr zerolog.Logger, err error) {
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
	ej := string(errJSON)

	// Write Content-Type headers
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// Write HTTP Statuscode
	w.WriteHeader(http.StatusInternalServerError)

	// Write response body (json)
	fmt.Fprintln(w, ej)
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
