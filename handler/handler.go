package handler

import (
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
			// get byte Array representation of guid from xid package (12 bytes)
			id := xid.New()

			// Send a new TraceID and the http.Request to the
			// StandardResponseFields constructor to set the
			// StandardResponseFields of the AppHandler
			ah.StandardResponseFields = controller.NewStandardResponseFields(controller.NewTraceID(id), r)

			h.ServeHTTP(w, r) // call original
		})
}

// NewAppHandler initializes the AppHandler
func NewAppHandler(app *app.Application) *AppHandler {
	return &AppHandler{App: app}
}

// NewMockAppHandler initializes the AppHandler
func NewMockAppHandler(app *app.Application, r *http.Request) *AppHandler {

	appHandler := &AppHandler{App: app}
	// Send a mocked TraceID and the http.Request to the
	// StandardResponseFields constructor to set the
	// StandardResponseFields of the AppHandler
	appHandler.StandardResponseFields = controller.NewStandardResponseFields(controller.NewMockTraceID(), r)

	return appHandler
}
