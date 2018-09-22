package app

import (
	"github.com/gilcrest/go-API-template/datastore"
	"github.com/gilcrest/httplog"
	"github.com/justinas/alice"
)

// routes registers handlers to the router
func (s *server) routes() error {

	// Get a logger instance from the server struct
	log := s.logger

	// Get logging Database to pass into httplog
	// Only need this if you plan to use the PostgreSQL
	// logging style of httplog
	logdb, err := s.ds.DB(datastore.LogDB)
	if err != nil {
		return err
	}

	s.router.HandleFunc("/v1/handlefunc/user",
		httplog.LogHandlerFunc(s.handleUserCreate(), log, logdb, nil))

	s.router.Handle("/v1/alice/user",
		alice.New(httplog.LogHandler(log, logdb, nil)).
			ThenFunc(s.handleUserCreate())).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// match only POST requests on /v1/adapter/user
	// having a Content-Type header = application/json
	s.router.Handle("/v1/adapter/user",
		httplog.Adapt(s.handleUserCreate(),
			httplog.LogAdapter(log, logdb, nil))).
		Methods("POST").
		Headers("Content-Type", "application/json")

	return nil
}
