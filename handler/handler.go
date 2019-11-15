package handler

import (
	"net/http"

	"github.com/gilcrest/go-api-basic/app"
)

// AppHandler is the struct that serves the application
// and methods for handling all HTTP requests
type AppHandler struct {
	App *app.Application
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

// ProvideAppHandler initializes the AppHandler
func ProvideAppHandler(app *app.Application) *AppHandler {
	return &AppHandler{app}
}
