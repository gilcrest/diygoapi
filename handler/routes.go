package handler

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/hlog"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
)

const (
	contentTypeHeaderKey        string = "Content-Type"
	appJSONContentTypeHeaderVal string = "application/json"
	pathPrefix                  string = "/api"
	moviesV1PathRoot            string = "/v1/movies"
	extlIDPathDir               string = "/{extlID}"
	loggerV1PathRoot            string = "/v1/logger"
	pingV1PathRoot              string = "/v1/ping"
)

// NewMuxRouter initializes a gorilla/mux router and registers routes
// and corresponding middleware/handlers to it
func NewMuxRouter(logger zerolog.Logger, mw Middleware, handlers Handlers) *mux.Router {

	// initializer gorilla/mux router
	rtr := mux.NewRouter()

	// Start a new alice handler chain
	c := alice.New()

	// add loggerHandlerChain handler chain and zerolog logger to Context
	c = loggerHandlerChain(logger, c)

	// send Router through PathPrefix method to validate any standard
	// subroutes you may want for your APIs. e.g. I always want to be
	// sure that every request has "/api" as part of it's path prefix
	// without having to put it into every handle path in my various
	// routing functions
	rtr = rtr.PathPrefix(pathPrefix).Subrouter()

	// Match only POST requests at /api/v1/movies
	// with Content-Type header = application/json
	rtr.Handle(moviesV1PathRoot,
		c.Extend(authChain(mw, c)).
			Append(mw.JSONContentTypeResponseMiddleware()).
			Then(handlers.CreateMovieHandler)).
		Methods(http.MethodPost).
		Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

	// Match only PUT requests having an ID at /api/v1/movies/{id}
	// with the Content-Type header = application/json
	rtr.Handle(moviesV1PathRoot+extlIDPathDir,
		c.Extend(authChain(mw, c)).
			Append(mw.JSONContentTypeResponseMiddleware()).
			Then(handlers.UpdateMovieHandler)).
		Methods(http.MethodPut).
		Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

	// Match only DELETE requests having an ID at /api/v1/movies/{id}
	rtr.Handle(moviesV1PathRoot+extlIDPathDir,
		c.Extend(authChain(mw, c)).
			Append(mw.JSONContentTypeResponseMiddleware()).
			Then(handlers.DeleteMovieHandler)).
		Methods(http.MethodDelete)

	// Match only GET requests having an ID at /api/v1/movies/{id}
	rtr.Handle(moviesV1PathRoot+extlIDPathDir,
		c.Extend(authChain(mw, c)).
			Append(mw.JSONContentTypeResponseMiddleware()).
			Then(handlers.FindMovieByIDHandler)).
		Methods(http.MethodGet)

	// Match only GET requests /api/v1/movies
	rtr.Handle(moviesV1PathRoot,
		c.Extend(authChain(mw, c)).
			Append(mw.JSONContentTypeResponseMiddleware()).
			Then(handlers.FindAllMoviesHandler)).
		Methods(http.MethodGet)

	// Match only GET requests /api/v1/logger
	rtr.Handle(loggerV1PathRoot,
		c.Extend(authChain(mw, c)).
			Append(mw.JSONContentTypeResponseMiddleware()).
			Then(handlers.ReadLoggerHandler)).
		Methods(http.MethodGet)

	// Match only PUT requests /api/v1/logger
	rtr.Handle(loggerV1PathRoot,
		c.Extend(authChain(mw, c)).
			Append(mw.JSONContentTypeResponseMiddleware()).
			Then(handlers.UpdateLoggerHandler)).
		Methods(http.MethodPut).
		Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

	// Match only GET requests at /api/v1/ping
	rtr.Handle(pingV1PathRoot,
		c.Append(mw.JSONContentTypeResponseMiddleware()).
			Then(handlers.PingHandler)).
		Methods(http.MethodGet)

	return rtr
}

// loggerHandlerChain returns a handler chain (via alice.Chain)
// initialized with all the standard handlers for logging. The logger
// will be added to the request context for subsequent use with pre-populated
// fields, including the request method, url, status, size, duration, remote IP,
// user agent, referer. A unique Request ID is also added to the logger, context
// and response headers.
func loggerHandlerChain(logger zerolog.Logger, c alice.Chain) alice.Chain {

	// Install the logger handler with default output on the console
	c = c.Append(hlog.NewHandler(logger))

	// Install extra handler to set request's context fields.
	// Thanks to that handler, all our logs will come with some pre-populated fields.
	c = c.Append(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Stringer("url", r.URL).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	})).
		Append(hlog.RemoteAddrHandler("remote_ip")).
		Append(hlog.UserAgentHandler("user_agent")).
		Append(hlog.RefererHandler("referer")).
		Append(hlog.RequestIDHandler("request_id", "Request-Id"))

	return c
}

// authChain chains the handlers together for User authorization
func authChain(mw Middleware, c alice.Chain) alice.Chain {
	c = c.Append(mw.AccessTokenMiddleware()).
		Append(mw.ConvertAccessTokenMiddleware()).
		Append(mw.AuthorizeUserMiddleware())

	return c
}
