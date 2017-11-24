package handlers

import (
	mwr "github.com/gilcrest/go-API-template/pkg/api/server/middleware"
	"github.com/gilcrest/go-API-template/pkg/env"
	"github.com/gorilla/mux"
)

// Dispatch is a way of organizing routing to handlers (versioning as well)
func Dispatch(env *env.Env, rtr *mux.Router) *mux.Router {

	// match only POST requests on /api/appUser/create
	// This is the original (v1) version for the API and the response for this
	// will never change with versioning in order to maintain a stable contract
	rtr.Handle("/appUser", mwr.Adapt(Handler{env, CreateUserHandler}, mwr.LogRequest(env))).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// match only POST requests on /api/v1/appUser/create
	rtr.Handle("/v1/appUser", mwr.Adapt(Handler{env, CreateUserHandler}, mwr.LogRequest(env))).
		Methods("POST").
		Headers("Content-Type", "application/json")

	return rtr
}
