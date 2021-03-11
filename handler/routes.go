package handler

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
)

const (
	pathPrefix       string = "/api"
	moviesV1PathRoot string = "/v1/movies"
)

// NewMuxRouter sets up the mux.Router and registers routes to URL paths
// using the available handlers
func NewMuxRouter(logger zerolog.Logger, handlers Handlers) *mux.Router {
	// create a new gorilla/mux router
	rtr := mux.NewRouter()

	// Start a new alice handler chain
	c := alice.New()

	// add LoggerHandlerChain handler chain and zerolog logger to Context
	c = LoggerHandlerChain(logger, c)

	// send Router through PathPrefix method to validate any standard
	// subroutes you may want for your APIs. e.g. I always want to be
	// sure that every request has "/api" as part of it's path prefix
	// without having to put it into every handle path in my various
	// routing functions
	rtr = rtr.PathPrefix(pathPrefix).Subrouter()

	// Match only POST requests at /api/v1/movies
	// with Content-Type header = application/json
	rtr.Handle(moviesV1PathRoot,
		c.Append(AccessTokenHandler).
			Append(JSONContentTypeHandler).
			Then(handlers.CreateMovieHandler)).
		Methods(http.MethodPost).
		Headers("Content-Type", "application/json")

	// Match only PUT requests having an ID at /api/v1/movies/{id}
	// with the Content-Type header = application/json
	rtr.Handle(moviesV1PathRoot+"/{extlID}",
		c.Append(AccessTokenHandler).
			Append(JSONContentTypeHandler).
			Then(handlers.UpdateMovieHandler)).
		Methods(http.MethodPut).
		Headers("Content-Type", "application/json")

	// Match only DELETE requests having an ID at /api/v1/movies/{id}
	rtr.Handle(moviesV1PathRoot+"/{extlID}",
		c.Append(AccessTokenHandler).
			Append(JSONContentTypeHandler).
			Then(handlers.DeleteMovieHandler)).
		Methods(http.MethodDelete)

	// Match only GET requests having an ID at /api/v1/movies/{id}
	rtr.Handle(moviesV1PathRoot+"/{extlID}",
		c.Append(AccessTokenHandler).
			Append(JSONContentTypeHandler).
			Then(handlers.FindMovieByIDHandler)).
		Methods(http.MethodGet)

	// Match only GET requests /api/v1/movies
	rtr.Handle(moviesV1PathRoot,
		c.Append(AccessTokenHandler).
			Append(JSONContentTypeHandler).
			Then(handlers.FindAllMoviesHandler)).
		Methods(http.MethodGet)

	// Match only GET requests at /api/v1/ping
	rtr.Handle("/v1/ping",
		c.Append(JSONContentTypeHandler).
			Then(handlers.PingHandler)).
		Methods(http.MethodGet)

	return rtr
}
