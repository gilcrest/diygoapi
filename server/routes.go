package server

import (
	"github.com/gilcrest/alice"
	"github.com/gilcrest/errors"
	"github.com/gilcrest/servertoken"
	"github.com/gilcrest/srvr/datastore"
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
			s.handleRespHeader,
			servertoken.Handler(s.Logger, appdb)).
			ThenFunc(s.handlePost())).
		Methods("POST").
		Headers("Content-Type", "application/json")

	return nil
}
