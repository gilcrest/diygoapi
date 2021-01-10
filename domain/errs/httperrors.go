package errs

import (
	"encoding/json"
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
func HTTPErrorResponse(w http.ResponseWriter, logger zerolog.Logger, err error) {

	var httpStatusCode int

	if err != nil {
		// perform a "type switch" https://tour.golang.org/methods/16
		// to determine the interface value type
		switch e := err.(type) {
		// If the interface value is of type Error (not a typical error, but
		// the Error interface defined above), then
		case *Error:
			httpStatusCode = httpErrorStatusCode(e.Kind)
			// We can retrieve the status here and write out a specific
			// HTTP status code. If the error is empty, just
			// send the HTTP Status Code as response
			if e.isZero() {
				logger.Error().Stack().Int("http_statuscode", httpStatusCode).Msg("empty error")
				sendError(w, "", httpStatusCode)
			} else if e.Kind == Unauthenticated {
				// For Unauthenticated and Unauthorized errors,
				// the response body should be empty. Use logger
				// to log the error and then just send
				// http.StatusUnauthorized (401) or http.StatusForbidden (403)
				// depending on the circumstances. "In summary, a
				// 401 Unauthorized response should be used for missing or bad authentication,
				// and a 403 Forbidden response should be used afterwards, when the user is
				// authenticated but isn’t authorized to perform the requested operation on
				// the given resource."
				logger.Error().Stack().Err(e.Err).
					Int("http_statuscode", http.StatusUnauthorized).
					Msg("Unauthenticated Request")
				sendError(w, "", httpStatusCode)
			} else if e.Kind == Unauthorized {
				logger.Error().Stack().Err(e.Err).
					Int("http_statuscode", http.StatusForbidden).
					Msg("Unauthorized Request")
				sendError(w, "", httpStatusCode)
			} else {
				// log the error with stacktrace
				logger.Error().Stack().Err(e.Err).
					Int("http_statuscode", httpStatusCode).
					Str("Kind", e.Kind.String()).
					Str("Parameter", string(e.Param)).
					Str("Code", string(e.Code)).
					Msg("Response Error Sent")

				// setup ErrResponse
				er := ErrResponse{
					Error: ServiceError{
						Kind:    e.Kind.String(),
						Code:    string(e.Code),
						Param:   string(e.Param),
						Message: e.Error(),
					},
				}

				// Marshal errResponse struct to JSON for the response body
				errJSON, _ := json.Marshal(er)

				sendError(w, string(errJSON), httpStatusCode)
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

			logger.Error().Msgf("Unknown Error - HTTP %d - %s", cd, err.Error())

			// Marshal errResponse struct to JSON for the response body
			errJSON, _ := json.Marshal(er)

			sendError(w, string(errJSON), cd)
		}
	} else {
		httpStatusCode = httpErrorStatusCode(Other)
		// if a nil error is passed, do not write a response body,
		// just send the HTTP Status Code
		logger.Error().Int("HTTP Error StatusCode", httpStatusCode).Msg("nil error - no response body sent")
		sendError(w, "", httpStatusCode)
	}
}

// Taken from standard library, but changed to send application/json as header
// Error replies to the request with the specified error message and HTTP code.
// It does not otherwise end the request; the caller should ensure no further
// writes are done to w.
// The error message should be json.
func sendError(w http.ResponseWriter, errStr string, httpStatusCode int) {
	if errStr != "" {
		w.Header().Set("Content-Type", "application/json")
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	// TODO - refactor this package to allow for WWW-Authenticate header on 401/403
	//if httpStatusCode == 401 || httpStatusCode == 403 {
	//	br := fmt.Sprintf(`Bearer realm="%s"`, realm)
	//	w.Header().Set("WWW-Authenticate", br)
	//}
	w.WriteHeader(httpStatusCode)
	// Only write response body if there is an error string populated
	if errStr != "" {
		_, _ = fmt.Fprintln(w, errStr)
	}
}

// httpErrorStatusCode maps an error Kind to an HTTP Status Code
func httpErrorStatusCode(k Kind) int {
	switch k {
	case Unauthenticated:
		return http.StatusUnauthorized
	case Unauthorized, Permission:
		return http.StatusForbidden
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
