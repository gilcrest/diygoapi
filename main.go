package main

import (
	"context"
	"flag"
	"os"

	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gocloud.dev/server"
)

// cliFlags are the command line flags parsed at startup
type cliFlags struct {
	logLevel  string
	env       string
	datastore string
}

func main() {
	cf := new(cliFlags)

	// loglvl flag allows for setting logging level, e.g. to run the server
	// with level set to debug, it'd be: ./server loglvl=debug
	// If not set, defaults to error
	flag.StringVar(&cf.logLevel, "loglvl", "error", "sets log level (debug, info, warn, fatal, panic, disabled)")

	// env flag allows for setting environment, e.g. Production, QA, etc.
	// you can of course, change these names to whatever you want,
	// these are just examples
	// example: -env=dev, -env=qa, -env=stg, -env=prod
	// If not set, defaults to dev
	flag.StringVar(&cf.env, "env", "local", "sets app environment (local, mock, qa, stg, prod)")

	// datastore flag will set which datastore is to be used. If
	// datastore=mock, the app is set to "mock mode" and no database
	// calls will be submitted and a mock (aka "stubbed") response
	// will be returned. If not set, defaults to false (not in "mock mode")
	flag.StringVar(&cf.datastore, "datastore", "local", "sets the app datastore")

	addr := flag.String("listen", ":8080", "port to listen for HTTP on")

	flag.Parse()

	// determine logging level
	loglvl := newLogLevel(cf)

	// get environment name
	envName := newEnvName(cf)

	// get Datastore name
	dsName := newDSName(cf)

	// initialize local variables
	var (
		srv     *server.Server
		cleanup func()
		err     error
	)

	ctx := context.Background()
	switch {
	case envName == app.Local:
		switch dsName {
		case datastore.LocalDatastore:
			srv, cleanup, err = setupApp(ctx, envName, dsName, loglvl)
		case datastore.GCPCPDatastore:
			srv, cleanup, err = setupApp(ctx, envName, dsName, loglvl)
		case datastore.GCPDatastore:
			srv, cleanup, err = setupApp(ctx, envName, dsName, loglvl)
		case datastore.MockDatastore:
			srv, cleanup, err = setupAppwMock(ctx, envName, dsName, loglvl)
		default:
			log.Fatal().Msgf("unknown datastore name = %s", dsName)
		}
	case envName == app.QA:
		switch dsName {
		case datastore.LocalDatastore:
			srv, cleanup, err = setupApp(ctx, envName, dsName, loglvl)
		case datastore.GCPCPDatastore:
			srv, cleanup, err = setupApp(ctx, envName, dsName, loglvl)
		case datastore.GCPDatastore:
			srv, cleanup, err = setupApp(ctx, envName, dsName, loglvl)
		case datastore.MockDatastore:
			srv, cleanup, err = setupAppwMock(ctx, envName, dsName, loglvl)
		default:
			log.Fatal().Msgf("unknown datastore name = %s", dsName)
		}
	default:
		log.Fatal().Msgf("unknown environment name = %s", envName)
	}
	if err != nil {
		log.Fatal().Err(err).Msg("Error returned from main switch")
	}
	defer cleanup()

	// Listen and serve HTTP
	log.Log().Msgf("Running, connected to the %s environment, datastore is set to %s", envName, dsName)
	log.Fatal().Err(srv.ListenAndServe(*addr))
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

	return name
}

func newDSName(flags *cliFlags) datastore.DSName {

	switch flags.datastore {
	case "mock":
		return datastore.MockDatastore
	case "local":
		return datastore.LocalDatastore
	case "gcpcp":
		return datastore.GCPCPDatastore
	case "gcp":
		return datastore.GCPDatastore
	default:
		return datastore.MockDatastore
	}

}

// NewLogger sets up the zerolog.Logger
func newLogger(lvl zerolog.Level) zerolog.Logger {
	// empty string for TimeFieldFormat will write logs with UNIX time
	zerolog.TimeFieldFormat = ""
	// set logging level based on input
	zerolog.SetGlobalLevel(lvl)
	// start a new logger with Stdout as the target
	lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()

	lgr.Log().Msgf("Logging Level set to %s", lvl)

	return lgr
}

// newLogLevel sets up the logging level (e.g. Debug, Info, Error, etc.)
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

	return lvl
}
