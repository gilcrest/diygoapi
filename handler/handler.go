package handler

import (
	"net/http"

	"github.com/gilcrest/go-api-basic/app"
	"github.com/rs/xid"
)

// AppHandler is the struct that serves the application
// and methods for handling all HTTP requests
type AppHandler struct {
	App       *app.Application
	RequestID xid.ID
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
			ah.RequestID = xid.New()

			h.ServeHTTP(w, r) // call original
		})
}

// ProvideAppHandler initializes the AppHandler
func ProvideAppHandler(app *app.Application) *AppHandler {
	return &AppHandler{App: app}
}
