package handler

import (
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
)

// NewMuxRouter sets up the mux.Router and registers routes to URL paths
// using the available handlers
func NewMuxRouter(logger zerolog.Logger, handlers Handlers) *mux.Router {
	// I should take this as a dependency, but need to do some work with wire
	rtr := mux.NewRouter()

	// I should take this as a dependency, but need to do some work with wire
	c := alice.New()

	// add Standard Handler chain and zerolog logger to Context
	c = AddStandardHandlerChain(logger, c)

	// send Router through PathPrefix method to validate any standard
	// subroutes you may want for your APIs. e.g. I always want to be
	// sure that every request has "/api" as part of it's path prefix
	// without having to put it into every handle path in my various
	// routing functions
	rtr = rtr.PathPrefix("/api").Subrouter()

	// Match only POST requests at /api/v1/movies
	// with Content-Type header = application/json
	rtr.Handle("/v1/movies",
		c.Append(AccessTokenHandler).
			Append(JSONContentTypeHandler).
			Then(handlers.CreateMovieHandler)).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// Match only PUT requests having an ID at /api/v1/movies/{id}
	// with the Content-Type header = application/json
	rtr.Handle("/v1/movies/{id}",
		c.Append(AccessTokenHandler).
			Append(JSONContentTypeHandler).
			Then(handlers.UpdateMovieHandler)).
		Methods("PUT").
		Headers("Content-Type", "application/json")

	// Match only DELETE requests having an ID at /api/v1/movies/{id}
	rtr.Handle("/v1/movies/{id}",
		c.Append(AccessTokenHandler).
			Append(JSONContentTypeHandler).
			Then(handlers.DeleteMovieHandler)).
		Methods("DELETE")

	// Match only GET requests having an ID at /api/v1/movies/{id}
	rtr.Handle("/v1/movies/{id}",
		c.Append(AccessTokenHandler).
			Append(JSONContentTypeHandler).
			Then(handlers.FindMovieByIDHandler)).
		Methods("GET")

	// Match only GET requests /api/v1/movies
	rtr.Handle("/v1/movies",
		c.Append(AccessTokenHandler).
			Append(JSONContentTypeHandler).
			Then(handlers.FindAllMoviesHandler)).
		Methods("GET")

	// Match only GET requests at /api/v1/ping
	rtr.Handle("/v1/ping",
		c.Append(JSONContentTypeHandler).
			Then(handlers.PingHandler)).
		Methods("GET")

	return rtr
}
