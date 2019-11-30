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

// AddRequestID middleware is used to add a unique request ID to each request
func (ah *AppHandler) AddRequestID(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// get byte Array representation of guid from xid package (12 bytes)
			id := xid.New()

			// Set the Request ID as xid.ID to
			// controller.RequestID using the constructor
			requestID := controller.NewRequestID(id)

			// Send the RequestID and http.Request to the
			// StandardResponseFields constructor to receive the
			// StandardResponseFields
			ah.StandardResponseFields = controller.NewStandardResponseFields(requestID, r)

			h.ServeHTTP(w, r) // call original
		})
}

// NewAppHandler initializes the AppHandler
func NewAppHandler(app *app.Application) *AppHandler {
	return &AppHandler{App: app}
}
