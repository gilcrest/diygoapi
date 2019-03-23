package server

import (
	"net/http"

	"github.com/gilcrest/alice"
	"github.com/gilcrest/env/datastore"
	"github.com/gilcrest/errors"
	"github.com/gilcrest/servertoken"
)

// routes registers handlers to the router
func (s *Server) routes() error {
	const op errors.Op = "server/Server.routes"

	// Get App Database for token authentication
	appdb, err := s.DS.DB(datastore.AppDB)
	if err != nil {
		return errors.E(op, err)
	}

	// Match only POST requests with Content-Type header = application/json
	s.Router.Handle("/v1/movie",
		alice.New(
			s.handleStdResponseHeader,
			servertoken.Handler(s.Logger, appdb)).
			ThenFunc(s.handlePost())).
		Methods("POST").
		Headers("Content-Type", "application/json")

	return nil
}

// handleStdResponseHeader middleware is used to add standard HTTP response headers
func (s *Server) handleStdResponseHeader(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			h.ServeHTTP(w, r) // call original
		})
}
