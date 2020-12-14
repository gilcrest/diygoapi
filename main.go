package main

import (
	"context"
	"flag"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	// listen flag is used for the http.ListenAndServe addr field
	addr := flag.String("listen", ":8080", "port to listen for HTTP on")

	// Parse the command line flags from above
	flag.Parse()

	// determine logging level
	loglvl := newLogLevel(cf)

	// initialize a non-nil, empty context
	ctx := context.Background()

	// newServer function returns a pointer to a gocloud server, a
	// cleanup function and an error
	srv, cleanup, err := newServer(ctx, loglvl)
	if err != nil {
		log.Fatal().Err(err).Msg("Error returned from newServer")
	}
	defer cleanup()

	// Listen and serve HTTP
	log.Fatal().Err(srv.ListenAndServe(*addr)).Msg("Fatal Server Error")
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
