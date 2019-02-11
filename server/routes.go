package server

import (
	"github.com/gilcrest/alice"
	"github.com/gilcrest/errors"
	"github.com/gilcrest/httplog"
	"github.com/gilcrest/servertoken"
	"github.com/gilcrest/srvr/datastore"
)

// routes registers handlers to the router
func (s *Server) routes() error {
	const op errors.Op = "server/routes"

	// Get pointer to logging database to pass into httplog
	// Only need this if you plan to use the PostgreSQL
	// logging style of httplog
	logdb, err := s.DS.DB(datastore.LogDB)
	if err != nil {
		return errors.E(op, err)
	}

	// Get App Database for token authentication
	appdb, err := s.DS.DB(datastore.AppDB)
	if err != nil {
		return errors.E(op, err)
	}

	// httplog.NewOpts gets a new httplog.Opts struct
	// (with all flags set to false)
	opts := new(httplog.Opts)
	opts.Option(httplog.Log2Database(true, true, true, true, true))

	// function (`LogHandler`) that takes a handler and returns a handler (aka Constructor)
	// (`func (http.Handler) http.Handler`)	- used with alice
	// Also, match only POST requests with Content-Type header = application/json
	s.Router.Handle("/v1/alice/movie",
		alice.New(
			httplog.LogHandler(s.Logger, logdb, opts),
			s.handleRespHeader,
			servertoken.Handler(s.Logger, appdb)).
			ThenFunc(s.handlePost())).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// HandlerFunc middleware example
	// function takes an http.HandlerFunc and returns an http.HandlerFunc
	// Also, match only POST requests with Content-Type header = application/json
	s.Router.HandleFunc("/v1/handlefunc/movie",
		httplog.LogHandlerFunc(s.handlePost(), s.Logger, logdb, opts)).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// Adapter Type middleware example
	// Also, match only POST requests with Content-Type header = application/json
	s.Router.Handle("/v1/adapter/movie",
		httplog.Adapt(s.handlePost(),
			httplog.LogAdapter(s.Logger, logdb, opts))).
		Methods("POST").
		Headers("Content-Type", "application/json")

	return nil
}
