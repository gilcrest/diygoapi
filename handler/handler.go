package handler

import (
	"github.com/gilcrest/errs"
	"io"
	"net/http"

	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/controller"
	"github.com/rs/xid"
)

// AppHandler is the struct that serves the application
// and methods for handling all HTTP requests
type AppHandler struct {
	App *app.Application
	controller.StandardResponseFields
}

// AddStandardResponseHeaders middleware is used to add any
// standard HTTP response headers
func (ah *AppHandler) AddStandardResponseHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			h.ServeHTTP(w, r) // call original
		})
}

// SetStandardResponseFields middleware is used to set the fields
// that are part of every response
func (ah *AppHandler) SetStandardResponseFields(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// If the app is being mocked, then we want a static TraceID for the response
			if ah.App.Mock {
				ah.StandardResponseFields = controller.NewStandardResponseFields(controller.NewMockTraceID(), r)
			} else {
				// get byte Array representation of guid from xid package (12 bytes)
				id := xid.New()

				// Send a new TraceID and the http.Request to the
				// StandardResponseFields constructor to set the
				// StandardResponseFields of the AppHandler
				ah.StandardResponseFields = controller.NewStandardResponseFields(controller.NewTraceID(id), r)
			}

			h.ServeHTTP(w, r) // call original
		})
}

// NewAppHandler initializes the AppHandler
func NewAppHandler(app *app.Application) *AppHandler {
	return &AppHandler{App: app}
}

// When an error is returned by json.NewDecoder(r.Body).Decode(&data)
// this function will determine the appropriate error response
func DecoderErr(err error) error {
	const op errs.Op = "handler/DecoderErr"

	switch {
	// If the request body is empty (io.EOF)
	// return an error
	case err == io.EOF:
		return errs.E(op, errs.InvalidRequest, "Request Body cannot be empty")
	// If the request body has malformed JSON (io.ErrUnexpectedEOF)
	// return an error
	case err == io.ErrUnexpectedEOF:
		return errs.E(op, errs.InvalidRequest, "Malformed JSON")
	// return all other errors
	case err != nil:
		return errs.E(op, err)
	}
	return nil
}
