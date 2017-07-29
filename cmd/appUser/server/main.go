package main

import (
	"log"
	"net/http"

	"github.com/gilcrest/go-API-template/pkg/config/env"
	"github.com/gilcrest/go-API-template/pkg/handlers"

	"github.com/gorilla/mux"
)

func main() {

	// Initializes "environment" type to be passed around to functions
	// func Init() (*Env, error)
	env, err := env.Init()

	if err != nil {
		log.Fatal(err)
	}

	// create a new mux (multiplex) router
	// func NewRouter() *Router
	r := mux.NewRouter()

	// API may have multiple versions and the matching may get a bit
	// lengthy, this PathMatch function helps with organizing that
	// func PathMatch(env *env.Env, rtr *mux.Router) *mux.Router
	r = handlers.PathMatch(env, r)

	// handle all requests with the Gorilla router by adding
	// r to the DefaultServeMux
	// func Handle(pattern string, handler Handler)
	http.Handle("/", r)

	// func ListenAndServe(addr string, handler Handler) error
	if err := http.ListenAndServe("127.0.0.1:8080", nil); err != nil {
		log.Fatal(err)
	}
}
