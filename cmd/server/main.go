package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/gilcrest/go-API-template/app"
	"github.com/rs/zerolog"
)

func main() {

	loglvlFlag := flag.String("loglvl", "error", "sets log level")

	flag.Parse()

	loglevel := logLevel(loglvlFlag)

	srv, err := app.NewServer(loglevel)
	if err != nil {
		panic(err)
	}

	// handle all requests with the Gorilla router
	http.Handle("/", srv.Router())

	// ListenAndServe on port 8080, not specifying a particular IP address
	// for this particular implementation
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
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
