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

	// Get pointer to logging database to pass into httplog
	// Only need this if you plan to use the PostgreSQL
	// logging style of httplog
	logdb, err := s.ds.DB(datastore.LogDB)
	if err != nil {
		return err
	}

	// httplog.NewOpts gets a new httplog.Opts struct
	// (with all flags set to false)
	opts := new(httplog.Opts)

	// For the examples below, I chose to turn on db logging only
	// Log the request headers only (body has password on this api!)
	// Log both the response headers and body
	opts.Log2DB.Enable = true
	opts.Log2DB.Request.Header = true
	opts.Log2DB.Response.Header = true
	opts.Log2DB.Response.Body = true

	// HandlerFunc middleware example
	// function takes an http.HandlerFunc and returns an http.HandlerFunc
	// Also, match only POST requests with Content-Type header = application/json
	s.router.HandleFunc("/v1/handlefunc/user",
		httplog.LogHandlerFunc(s.handleUserCreate(), log, logdb, opts)).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// function (`LogHandler`) that takes a handler and returns a handler (aka Constructor)
	// (`func (http.Handler) http.Handler`)	- used with alice
	// Also, match only POST requests with Content-Type header = application/json
	s.router.Handle("/v1/alice/user",
		alice.New(httplog.LogHandler(log, logdb, opts), s.handleStdHeader, s.handleAuth).
			ThenFunc(s.handleUserCreate())).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// Adapter Type middleware example
	// Also, match only POST requests with Content-Type header = application/json
	s.router.Handle("/v1/adapter/user",
		httplog.Adapt(s.handleUserCreate(),
			httplog.LogAdapter(log, logdb, opts))).
		Methods("POST").
		Headers("Content-Type", "application/json")

	return nil
}
