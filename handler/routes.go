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
func NewMuxRouter(lgr zerolog.Logger, mw Middleware, handlers Handlers) *mux.Router {

	// initializer gorilla/mux router
	rtr := mux.NewRouter()

	// Start a new alice handler chain
	c := alice.New()

	// add logger middleware chain; add zerolog to Context
	chainWithLogger := c.Extend(loggerChain(lgr))

	//chainWithLoggerAuth := chainWithLogger.Extend(authChain(mw))

	// send Router through PathPrefix method to validate any standard
	// subroutes you may want for your APIs. e.g. I always want to be
	// sure that every request has "/api" as part of it's path prefix
	// without having to put it into every handle path in my various
	// routing functions
	rtr = rtr.PathPrefix(pathPrefix).Subrouter()

	// Match only POST requests at /api/v1/movies
	// with Content-Type header = application/json
	rtr.Handle(moviesV1PathRoot,
		chainWithLogger.Extend(authChain(mw)).
			Append(mw.JSONContentTypeResponseMiddleware()).
			Then(handlers.CreateMovieHandler)).
		Methods(http.MethodPost).
		Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

	// Match only PUT requests having an ID at /api/v1/movies/{id}
	// with the Content-Type header = application/json
	rtr.Handle(moviesV1PathRoot+extlIDPathDir,
		chainWithLogger.Extend(authChain(mw)).
			Append(mw.JSONContentTypeResponseMiddleware()).
			Then(handlers.UpdateMovieHandler)).
		Methods(http.MethodPut).
		Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

	// Match only DELETE requests having an ID at /api/v1/movies/{id}
	rtr.Handle(moviesV1PathRoot+extlIDPathDir,
		chainWithLogger.Extend(authChain(mw)).
			Append(mw.JSONContentTypeResponseMiddleware()).
			Then(handlers.DeleteMovieHandler)).
		Methods(http.MethodDelete)

	// Match only GET requests having an ID at /api/v1/movies/{id}
	rtr.Handle(moviesV1PathRoot+extlIDPathDir,
		chainWithLogger.Extend(authChain(mw)).
			Append(mw.JSONContentTypeResponseMiddleware()).
			Then(handlers.FindMovieByIDHandler)).
		Methods(http.MethodGet)

	// Match only GET requests /api/v1/movies
	rtr.Handle(moviesV1PathRoot,
		chainWithLogger.Extend(authChain(mw)).
			Append(mw.JSONContentTypeResponseMiddleware()).
			Then(handlers.FindAllMoviesHandler)).
		Methods(http.MethodGet)

	// Match only GET requests /api/v1/logger
	rtr.Handle(loggerV1PathRoot,
		chainWithLogger.Extend(authChain(mw)).
			Append(mw.JSONContentTypeResponseMiddleware()).
			Then(handlers.ReadLoggerHandler)).
		Methods(http.MethodGet)

	// Match only PUT requests /api/v1/logger
	rtr.Handle(loggerV1PathRoot,
		chainWithLogger.Extend(authChain(mw)).
			Append(mw.JSONContentTypeResponseMiddleware()).
			Then(handlers.UpdateLoggerHandler)).
		Methods(http.MethodPut).
		Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

	// Match only GET requests at /api/v1/ping
	rtr.Handle(pingV1PathRoot,
		chainWithLogger.Append(mw.JSONContentTypeResponseMiddleware()).
			Then(handlers.PingHandler)).
		Methods(http.MethodGet)

	return rtr
}

// loggerChain returns a middleware chain (via alice.Chain)
// initialized with all the standard middleware handlers for logging. The logger
// will be added to the request context for subsequent use with pre-populated
// fields, including the request method, url, status, size, duration, remote IP,
// user agent, referer. A unique Request ID is also added to the logger, context
// and response headers.
func loggerChain(logger zerolog.Logger) alice.Chain {

	ac := alice.New(hlog.NewHandler(logger),
		hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Stringer("url", r.URL).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Msg("request logged")
		}),
		hlog.RemoteAddrHandler("remote_ip"),
		hlog.UserAgentHandler("user_agent"),
		hlog.RefererHandler("referer"),
		hlog.RequestIDHandler("request_id", "Request-Id"),
	)

	return ac
}

// authChain chains the handlers together for User authorization
func authChain(mw Middleware) alice.Chain {
	ac := alice.New(mw.AccessTokenMiddleware(),
		mw.ConvertAccessTokenMiddleware(),
		mw.AuthorizeUserMiddleware())

	return ac
}
