package main

import (
	"log"
	"net/http"

	"github.com/gilcrest/go-API-template/pkg/api/server/handlers"
	mwr "github.com/gilcrest/go-API-template/pkg/api/server/middleware"
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

	// API may have multiple versions and the matching may get a bit
	// lengthy, this PathMatch function helps with organizing that
	rtr = handlers.PathMatch(env, rtr)

	// handle all requests with the Gorilla router by adding
	// rtr to the DefaultServeMux
	// LogRequest middleware will log request as well
	http.Handle("/", mwr.LogRequest(env, rtr))

	if err := http.ListenAndServe("127.0.0.1:8080", nil); err != nil {
		log.Fatal(err)
	}
}
