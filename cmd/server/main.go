package main

import (
	"log"
	"net/http"

	"github.com/gilcrest/go-API-template/env"
	"github.com/gilcrest/go-API-template/server/dispatch"
	"github.com/gorilla/mux"
)

func main() {

	// Initializes "environment" struct type
	env, err := env.NewEnv()

	if err != nil {
		log.Fatal(err)
	}

	// create a new mux (multiplex) router
	rtr := mux.NewRouter()

	// send Router through subRouter function to add any standard
	// Subroutes you may want for your APIs
	r := dispatch.NewSubrouter(rtr)

	// API may have multiple versions and the matching may get a bit
	// lengthy, this Dispatch function helps with organizing that
	r = dispatch.Dispatch(env, r)

	// handle all requests with the Gorilla router by adding
	// rtr to the DefaultServeMux
	// LogRequest middleware will log request as well
	http.Handle("/", r)

	// ListenAndServe on port 8080, not specifying a particular IP address
	// for this particular implementation
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
