package main

import (
	"flag"
	"net/http"

	"github.com/rs/zerolog/log"
)

// cliFlags are the command line flags parsed at startup
type cliFlags struct {
	logLevel string
	envName  string
}

func main() {
	cf := new(cliFlags)

	// loglvl flag allows for setting logging level, e.g. to run the server
	// with level set to debug, it'd be: ./server loglvl=debug
	// If not set, defaults to error
	flag.StringVar(&cf.logLevel, "loglvl", "error", "sets log level (debug, info, warn, fatal, panic, disabled)")

	// env flag allows for setting environment, e.g. Production, QA, etc.
	// example: env=dev, env=qa, env=stg, env=prod
	// If not set, defaults to dev
	flag.StringVar(&cf.envName, "env", "dev", "sets app environment (dev, qa, stg, prod)")

	flag.Parse()

	app, err := setupApplication(cf)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	// handle all requests with the Gorilla router
	http.Handle("/", app.router)

	// ListenAndServe on port 8080, not specifying a particular IP address
	// for this particular implementation
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal().Err(err).Msg("")
	}
}
