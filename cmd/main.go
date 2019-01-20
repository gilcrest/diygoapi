package main

import (
	"flag"
	"net/http"

	"github.com/gilcrest/go-API-template/server"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {

	// loglvl flag allows for setting logging level, e.g. to run the server
	// with level set to debug, it'd be: ./server loglvl=debug
	loglvlFlag := flag.String("loglvl", "error", "sets log level")

	flag.Parse()

	// get appropriate zerolog.Level based on flag
	loglevel := logLevel(loglvlFlag)

	// call constructor for Server struct
	server, err := server.NewServer(loglevel)
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
