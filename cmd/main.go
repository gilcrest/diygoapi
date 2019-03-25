package main

import (
	"flag"
	"net/http"

	"github.com/gilcrest/envy"
	"github.com/gilcrest/go-api-basic/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {

	// loglvl flag allows for setting logging level, e.g. to run the server
	// with level set to debug, it'd be: ./server loglvl=debug
	loglvlFlag := flag.String("loglvl", "error", "sets log level")

	// env flag allows for setting environment, e.g. Production, QA, etc.
	// example: env=dev, env=qa, env=stg, env=prod
	envFlag := flag.String("env", "dev", "sets log level")

	flag.Parse()

	// get appropriate zerolog.Level based on flag
	loglevel := logLevel(loglvlFlag)
	log.Log().Msgf("Logging Level set to %s", loglevel)

	// get appropriate envy.Name based on flag
	eName := envName(envFlag)
	log.Log().Msgf("Environment set to %s", eName)

	// call constructor for Server struct
	server, err := server.NewServer(eName, loglevel)
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	// handle all requests with the Gorilla router
	http.Handle("/", server.Router)

	// ListenAndServe on port 8080, not specifying a particular IP address
	// for this particular implementation
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal().Err(err).Msg("")
	}
}

func logLevel(s *string) zerolog.Level {

	var lvl zerolog.Level

	// dereference the string pointer to get flag value
	ds := *s

	switch ds {
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

func envName(s *string) envy.Name {

	var name envy.Name

	// dereference the string pointer to get flag value
	ds := *s

	switch ds {
	case "dev":
		name = envy.Dev
	case "qa":
		name = envy.QA
	case "stg":
		name = envy.Staging
	case "prod":
		name = envy.Production
	default:
		name = envy.Dev
	}
	return name
}
