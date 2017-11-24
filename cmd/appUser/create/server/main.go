package main

import (
	"log"
	"net/http"

	"github.com/gilcrest/go-API-template/pkg/api/server/handlers"
	"github.com/gilcrest/go-API-template/pkg/env"

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
	r := handlers.NewSubrouter(rtr)

	// API may have multiple versions and the matching may get a bit
	// lengthy, this RouteMatch function helps with organizing that
	r = handlers.Dispatch(env, r)

	// handle all requests with the Gorilla router by adding
	// rtr to the DefaultServeMux
	// LogRequest middleware will log request as well
	http.Handle("/", r)

	if err := http.ListenAndServe("127.0.0.1:8080", nil); err != nil {
		log.Fatal(err)
	}
}
