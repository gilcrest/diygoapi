package handlers

import (
	"github.com/gilcrest/go-API-template/pkg/config/env"
	"github.com/gorilla/mux"
)

// PathMatch is a way of organizing routing to handlers (versioning as well)
func PathMatch(env *env.Env, rtr *mux.Router) *mux.Router {

	// match only POST requests on /api/appUser/create
	// This is the original (v1) version for the API and the response for this will never change
	//  with versioning in order to maintain a stable contract
	// func (r *Router) HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *Route
	rtr.Handle("/api/appUser/create", Handler{env, CreateUserHandler}).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// match only POST requests on /api/v1/appUser/create
	// func (r *Router) HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *Route
	//rtr.HandleFunc("/api/v1/appUser/create", createUserHandler).
	//	Methods("POST").
	//	Headers("Content-Type", "application/json")

	return rtr
}
