package main

import (
	"context"
	"flag"
	"os"

	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/handler"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gocloud.dev/server"
)

// cliFlags are the command line flags parsed at startup
type cliFlags struct {
	logLevel string
	env      string
	mock     bool
}

func main() {
	cf := new(cliFlags)

	// loglvl flag allows for setting logging level, e.g. to run the server
	// with level set to debug, it'd be: ./server loglvl=debug
	// If not set, defaults to error
	flag.StringVar(&cf.logLevel, "loglvl", "error", "sets log level (debug, info, warn, fatal, panic, disabled)")

	// env flag allows for setting environment, e.g. Production, QA, etc.
	// example: -env=dev, -env=qa, -env=stg, -env=prod
	// If not set, defaults to dev
	flag.StringVar(&cf.env, "env", "local", "sets app environment (local, qa, stg, prod)")

	// mock flag will set the app to "mock mode" and no database
	// calls will be submitted and a mock (aka "stubbed") response
	// will be returned. If not set, defaults to false (not in "mock mode")
	flag.BoolVar(&cf.mock, "mock", false, "API will not submit anything to the database and return a mocked response")

	addr := flag.String("listen", ":8080", "port to listen for HTTP on")

	flag.Parse()

	log.Log().Msgf("Server Mock Mode set to %t\n", cf.mock)

	ctx := context.Background()
	var srv *server.Server
	var cleanup func()
	var err error
	switch cf.env {
	case "local":
		srv, cleanup, err = setupLocal(ctx, cf)
	default:
		log.Fatal().Msgf("unknown -env=%s", cf.env)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("Error returned from main switch")
	}
	defer cleanup()

	// Listen and serve HTTP.
	log.Printf("Running, connected to %q cloud", cf.env)
	log.Fatal().Err(srv.ListenAndServe(*addr))

	// rtr, err := setupRouter(cf)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("")
	// }

	// // handle all requests with the Gorilla router
	// http.Handle("/", rtr)

	// // ListenAndServe on port 8080, not specifying a particular IP address
	// // for this particular implementation
	// if err := http.ListenAndServe(":8080", nil); err != nil {
	// 	log.Fatal().Err(err).Msg("")
	// }
}

// newEnvName sets up the environment name (e.g. Production, Staging, QA, etc.)
// It takes a pointer to a string as that is how a parsed command line flag news
// and the intention is for the name to be set at run time
func newEnvName(flags *cliFlags) app.EnvName {

	var name app.EnvName

	switch flags.env {
	case "local":
		name = app.Local
	case "qa":
		name = app.QA
	case "stg":
		name = app.Staging
	case "prod":
		name = app.Production
	default:
		name = app.Local
	}

	log.Log().Msgf("Environment set to %s", name)

	return name
}

// NewLogger sets up the zerolog.Logger
func newLogger(lvl zerolog.Level) zerolog.Logger {
	// empty string for TimeFieldFormat will write logs with UNIX time
	zerolog.TimeFieldFormat = ""
	// set logging level based on input
	zerolog.SetGlobalLevel(lvl)
	// start a new logger with Stdout as the target
	lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return lgr
}

// NewLogLevel sets up the logging level (e.g. Debug, Info, Error, etc.)
// It takes a pointer to a string as that is how a parsed command line flag news
// and the intention is for the name to be set at run time
func newLogLevel(flags *cliFlags) zerolog.Level {

	var lvl zerolog.Level

	switch flags.logLevel {
	case "debug":
		lvl = zerolog.DebugLevel
	case "info":
		lvl = zerolog.InfoLevel
	case "warn":
		lvl = zerolog.WarnLevel
	case "fatal":
		lvl = zerolog.FatalLevel
	case "panic":
		lvl = zerolog.PanicLevel
	case "disabled":
		lvl = zerolog.Disabled
	default:
		lvl = zerolog.ErrorLevel
	}

	log.Log().Msgf("Logging Level set to %s", lvl)

	return lvl
}

func newDSName(flags *cliFlags) datastore.DSName {

	if flags.mock {
		return datastore.MockDatastore
	}

	return datastore.AppDatastore
}

func newRouter(hdl *handler.AppHandler) *mux.Router {
	// create a new mux (multiplex) router
	rtr := mux.NewRouter()

	// send Router through PathPrefix method to validate any standard
	// subroutes you may want for your APIs. e.g. I always want to be
	// sure that every request has "/api" as part of it's path prefix
	// without having to put it into every handle path in my various
	// routing functions
	rtr = rtr.PathPrefix("/api").Subrouter()

	// Match only POST requests at /api/v1/movies
	// with Content-Type header = application/json
	rtr.Handle("/v1/movies",
		alice.New(
			hdl.AddStandardResponseHeaders,
			hdl.AddRequestID).
			ThenFunc(hdl.AddMovie())).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// Match only GET requests having an ID at /api/v1/movies/{id}
	// with the Content-Type header = application/json
	rtr.Handle("/v1/movies/{id}",
		alice.New(
			hdl.AddStandardResponseHeaders,
			hdl.AddRequestID).
			ThenFunc(hdl.FindByID())).
		Methods("GET")

	// Match only GET requests /api/v1/movies
	// with the Content-Type header = application/json
	rtr.Handle("/v1/movies",
		alice.New(
			hdl.AddStandardResponseHeaders,
			hdl.AddRequestID).
			ThenFunc(hdl.FindAll())).
		Methods("GET")

	// Match only PUT requests having an ID at /api/v1/movies/{id}
	// with the Content-Type header = application/json
	rtr.Handle("/v1/movies/{id}",
		alice.New(
			hdl.AddStandardResponseHeaders,
			hdl.AddRequestID).
			ThenFunc(hdl.Update())).
		Methods("PUT").
		Headers("Content-Type", "application/json")

	// Match only DELETE requests having an ID at /api/v1/movies/{id}
	rtr.Handle("/v1/movies/{id}",
		alice.New(
			hdl.AddStandardResponseHeaders,
			hdl.AddRequestID).
			ThenFunc(hdl.Delete())).
		Methods("DELETE")

	return rtr
}
