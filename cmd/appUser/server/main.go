package main

import (
	"log"
	"net/http"

	"github.com/gilcrest/go-API-template/pkg/config/db"
	"github.com/gilcrest/go-API-template/pkg/config/env"
	"github.com/gilcrest/go-API-template/pkg/handlers"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {

	// Initializes "environment" type to be passed around to functions
	// func envInit() *env.Env
	env := envInit()

	// create a new mux (multiplex) router
	// func NewRouter() *Router
	r := mux.NewRouter()

	// API may have multiple versions and the matching may get a bit
	// lengthy, this routeMatcher function helps with organizing that
	// func routeMatcher(rtr *mux.Router) *mux.Router
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

// Initializes "environment" object to be passed around to functions
func envInit() *env.Env {

	logger, _ := zap.NewProduction()

	// returns an open database handle of 0 or more underlying connections
	// func NewDB() (*sql.DB, error)
	sqldb, err := db.NewDB()

	if err != nil {
		log.Fatal(err)
	}

	environment := &env.Env{Db: sqldb, Logger: logger}

	return environment

}
