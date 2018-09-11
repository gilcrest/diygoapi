package app

import (
	"github.com/gilcrest/go-API-template/datastore"
	"github.com/gilcrest/httplog"
)

// routes registers handlers to the router
func (s *server) routes() {

	log := s.logger

	logdb, err := s.ds.DB(datastore.LogDB)
	if err != nil {
		// TODO - bogus...
		panic(err)
	}

	// match only POST requests on /api/v1/appuser
	s.router.Handle("/v1/appuser",
		httplog.Adapt(handleErr{H: s.handleUserCreate},
			httplog.LogAdapter(log, logdb, nil))).
		Methods("POST").
		Headers("Content-Type", "application/json")

}
