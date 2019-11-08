package main

import (
	"database/sql"

	"github.com/gilcrest/go-api-basic/handler"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
)

func provideRouter(db *sql.DB, log zerolog.Logger) *mux.Router {
	// create a new mux (multiplex) router
	rtr := mux.NewRouter()
	// send Router through PathPrefix method to add any standard
	// subroutes you may want for your APIs. e.g. I always want to be
	// sure that every request has "/api" as part of it's path prefix
	// without having to put it into every handle path in my various
	// routing functions
	rtr = rtr.PathPrefix("/api").Subrouter()

	// Match only POST requests with Content-Type header = application/json
	rtr.Handle("/v1/movie",
		alice.New(
			handler.StdResponseHeader).
			ThenFunc(handler.AddMovie(db, log))).
		Methods("POST").
		Headers("Content-Type", "application/json")

	return rtr
}
