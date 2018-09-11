package main

import (
	"log"
	"net/http"

	"github.com/gilcrest/go-API-template/app"
	"github.com/rs/zerolog"
)

func main() {

	srv, err := app.NewServer(zerolog.DebugLevel)
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
