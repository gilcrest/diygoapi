package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/gilcrest/go-api-basic/app"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// cliFlags are the command line flags parsed at startup
type cliFlags struct {
	logLevel string
	port     int
}

func main() {
	// Initialize cliFlags and return a pointer to it
	cf := new(cliFlags)

	// loglvl flag allows for setting logging level, e.g. to run the server
	// with level set to debug, it'd be: ./server loglvl=debug
	// If not set, defaults to error
	flag.StringVar(&cf.logLevel, "loglvl", "info", "sets log level (debug, warn, error, fatal, panic, disabled)")

	// port flag is what http.ListenAndServe will listen on. default is 8080 if not set
	flag.IntVar(&cf.port, "port", 8080, "network port to listen on")

	// Parse the command line flags from above
	flag.Parse()

	// determine logging level
	loglvl := newLogLevel(cf)

	// setup logger with appropriate defaults
	logger := app.NewLogger(loglvl)

	// validate port in acceptable range
	if cf.port < 0 || cf.port > 65535 {
		logger.Fatal().Msgf("port %d is not within valid port range (0 to 65535", cf.port)
	}

	// initialize a non-nil, empty context
	ctx := context.Background()

	// newServer function returns a pointer to a gocloud server, a
	// cleanup function and an error
	srv, cleanup, err := newServer(ctx, logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Error returned from newServer")
	}
	defer cleanup()

	// Listen and serve HTTP
	log.Fatal().Err(srv.ListenAndServe(fmt.Sprintf(":%d", cf.port))).Msg("Fatal Server Error")
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
	case "error":
		lvl = zerolog.ErrorLevel
	case "fatal":
		lvl = zerolog.FatalLevel
	case "panic":
		lvl = zerolog.PanicLevel
	case "disabled":
		lvl = zerolog.Disabled
	default:
		lvl = zerolog.InfoLevel
	}

	return lvl
}
