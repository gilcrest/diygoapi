package handler

import (
	"net/http"

	"github.com/gorilla/mux"
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

// NewMuxRouterWithSubroutes initializes a gorilla/mux router and
// adds the /api subroute to it
func NewMuxRouterWithSubroutes() *mux.Router {
	// initializer gorilla/mux router
	r := mux.NewRouter()

	// send Router through PathPrefix method to validate any standard
	// subroutes you may want for your APIs. e.g. I always want to be
	// sure that every request has "/api" as part of it's path prefix
	// without having to put it into every handle path in my various
	// routing functions
	s := r.PathPrefix(pathPrefix).Subrouter()

	return s
}

// Routes registers routes corresponding middleware/handlers to the
// given gorilla/mux router
func Routes(rtr *mux.Router, mw Middleware, handlers Handlers) {

	// Match only POST requests at /api/v1/movies
	// with Content-Type header = application/json
	rtr.Handle(moviesV1PathRoot,
		mw.LoggerChain().Extend(mw.CtxWithUserChain()).
			Append(mw.AuthorizeUserHandler).
			Append(mw.JSONContentTypeResponseHandler).
			Then(handlers.CreateMovieHandler)).
		Methods(http.MethodPost).
		Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

	// Match only PUT requests having an ID at /api/v1/movies/{extlID}
	// with the Content-Type header = application/json
	rtr.Handle(moviesV1PathRoot+extlIDPathDir,
		mw.LoggerChain().Extend(mw.CtxWithUserChain()).
			Append(mw.AuthorizeUserHandler).
			Append(mw.JSONContentTypeResponseHandler).
			Then(handlers.UpdateMovieHandler)).
		Methods(http.MethodPut).
		Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

	// Match only DELETE requests having an ID at /api/v1/movies/{extlID}
	rtr.Handle(moviesV1PathRoot+extlIDPathDir,
		mw.LoggerChain().Extend(mw.CtxWithUserChain()).
			Append(mw.AuthorizeUserHandler).
			Append(mw.JSONContentTypeResponseHandler).
			Then(handlers.DeleteMovieHandler)).
		Methods(http.MethodDelete)

	// Match only GET requests having an ID at /api/v1/movies/{extlID}
	rtr.Handle(moviesV1PathRoot+extlIDPathDir,
		mw.LoggerChain().Extend(mw.CtxWithUserChain()).
			Append(mw.AuthorizeUserHandler).
			Append(mw.JSONContentTypeResponseHandler).
			Then(handlers.FindMovieByIDHandler)).
		Methods(http.MethodGet)

	// Match only GET requests /api/v1/movies
	rtr.Handle(moviesV1PathRoot,
		mw.LoggerChain().Extend(mw.CtxWithUserChain()).
			Append(mw.AuthorizeUserHandler).
			Append(mw.JSONContentTypeResponseHandler).
			Then(handlers.FindAllMoviesHandler)).
		Methods(http.MethodGet)

	// Match only GET requests /api/v1/logger
	rtr.Handle(loggerV1PathRoot,
		mw.LoggerChain().Extend(mw.CtxWithUserChain()).
			Append(mw.AuthorizeUserHandler).
			Append(mw.JSONContentTypeResponseHandler).
			Then(handlers.ReadLoggerHandler)).
		Methods(http.MethodGet)

	// Match only PUT requests /api/v1/logger
	rtr.Handle(loggerV1PathRoot,
		mw.LoggerChain().Extend(mw.CtxWithUserChain()).
			Append(mw.AuthorizeUserHandler).
			Append(mw.JSONContentTypeResponseHandler).
			Then(handlers.UpdateLoggerHandler)).
		Methods(http.MethodPut).
		Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

	// Match only GET requests at /api/v1/ping
	rtr.Handle(pingV1PathRoot,
		mw.LoggerChain().
			Append(mw.JSONContentTypeResponseHandler).
			Then(handlers.PingHandler)).
		Methods(http.MethodGet)

}
