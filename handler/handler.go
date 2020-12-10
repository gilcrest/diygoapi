package handler

import (
	"net/http"
	"time"

	"github.com/justinas/alice"
	"github.com/rs/zerolog/hlog"

	"github.com/gilcrest/go-api-basic/app"
)

// NewAppHandler initializes the AppHandler
func NewAppHandler(app *app.Application) *AppHandler {
	return &AppHandler{App: app}
}

// AppHandler is the struct that serves the application
// and methods for handling all HTTP requests
type AppHandler struct {
	App *app.Application
}

// ResponseHeaderHandler middleware is used to add any
// standard HTTP response headers. All of the responses for this app
// have a JSON based response body
func (ah *AppHandler) ResponseHeaderHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			h.ServeHTTP(w, r) // call original
		})
}

// AddStandardHandlerChain returns an alice.Chain initialized with all the standard
// handlers for logging, authentication and response headers and fields
func (ah *AppHandler) AddStandardHandlerChain(c alice.Chain) alice.Chain {

	// Install the logger handler with default output on the console
	c = c.Append(hlog.NewHandler(ah.App.Logger))

	// Install some provided extra handler to set some request's context fields.
	// Thanks to that handler, all our logs will come with some pre-populated fields.
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	c = c.Append(hlog.RemoteAddrHandler("ip"))
	c = c.Append(hlog.UserAgentHandler("user_agent"))
	c = c.Append(hlog.RefererHandler("referer"))
	c = c.Append(hlog.RequestIDHandler("req_id", "Request-Id"))
	c = c.Append(ah.ResponseHeaderHandler)
	//c = c.Append(ah.StandardResponseFieldsHandler)

	return c
}
