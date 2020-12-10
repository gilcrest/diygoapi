package main

import (
	"context"
	"flag"

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
	// Initialize cliFlags and return a pointer to it
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

	// listen flag is used for the http.ListenAndServe addr field
	addr := flag.String("listen", ":8080", "port to listen for HTTP on")

	// Parse the command line flags from above
	flag.Parse()

	// determine logging level
	loglvl := newLogLevel(cf)

	// get environment name
	envName := newEnvName(cf)

	// get Datastore name
	dsName := newDSName(cf)

	// initialize a non-nil, empty context
	ctx := context.Background()

	// newServer function returns a pointer to a gocloud server
	// a cleanup function and an error
	srv, cleanup, err := newServer(ctx, envName, dsName, loglvl)
	if err != nil {
		log.Fatal().Err(err).Msg("Error returned from newServer")
	}
	defer cleanup()

	// Listen and serve HTTP
	log.Log().Msgf("Running, connected to the %s environment, Datastore is set to %s", envName, dsName)
	log.Fatal().Err(srv.ListenAndServe(*addr)).Msg("Fatal Server Error")
}

func newServer(ctx context.Context, envName app.EnvName, dsName datastore.Name, loglvl zerolog.Level) (*server.Server, func(), error) {
	// initialize local variables
	var (
		srv     *server.Server
		cleanup func()
		err     error
	)
	// The switch below is meant to show how you may want to setup a
	// different datastore for different environments. For instance,
	// I've decided that QA is where I'll deploy the app to GCP's
	// "Cloud Run", so I don't consider any local connections there
	// and the opposite for Local, I don't allow for GCPDatastore
	// there - only connections I could actually connect to locally.
	// The MockedDatastore is available in all environments
	switch {
	case envName == app.Local:
		switch dsName {
		case datastore.LocalDatastore: // Connect to the Local Datastore
			srv, cleanup, err = setupApp(ctx, envName, dsName, loglvl)
		case datastore.GCPCPDatastore: // Connect to the GCP Cloud Proxy Datastore
			srv, cleanup, err = setupApp(ctx, envName, dsName, loglvl)
		default:
			log.Fatal().Msgf("unknown datastore name (%s) for the %s environment", dsName, envName)
		}
	case envName == app.QA:
		switch dsName {
		case datastore.GCPDatastore:
			srv, cleanup, err = setupApp(ctx, envName, dsName, loglvl)
		default:
			log.Fatal().Msgf("unknown datastore name (%s) for the %s environment", dsName, envName)
		}
	default:
		log.Fatal().Msgf("unknown environment name = %s", envName)
	}
	return srv, cleanup, err
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

// newDatastoreName determines the datastore.Name based on
// flags passed in
func newDSName(flags *cliFlags) datastore.Name {

	switch flags.datastore {
	case "mock":
		return datastore.MockedDatastore
	case "local":
		return datastore.LocalDatastore
	case "gcpcp":
		return datastore.GCPCPDatastore
	case "gcp":
		return datastore.GCPDatastore
	default:
		return datastore.MockedDatastore
	}

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
