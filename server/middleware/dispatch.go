package middleware

import (
	hdl "github.com/gilcrest/go-API-template/appUser/handler"
	"github.com/gilcrest/go-API-template/env"
	"github.com/gorilla/mux"
)

// Dispatch is a way of organizing routing to handlers (versioning as well)
func Dispatch(env *env.Env, rtr *mux.Router) *mux.Router {

	// initialize new instance of APIAudit
	audit := new(APIAudit)

	// match only POST requests on /api/appUser/create
	// This is the original (v1) version for the API and the response for this
	// will never change with versioning in order to maintain a stable contract
	rtr.Handle("/appUser", Adapt(hdl.ErrHandler{Env: env, H: hdl.CreateUser}, LogRequest(env, audit), LogResponse(env, audit))).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// match only POST requests on /api/v1/appUser/create
	rtr.Handle("/v1/appUser", Adapt(hdl.ErrHandler{Env: env, H: hdl.CreateUser}, LogRequest(env, audit), LogResponse(env, audit))).
		Methods("POST").
		Headers("Content-Type", "application/json")

	return rtr
}

// NewSubrouter adds any subRouters that you'd like to have as part of
// every request, i.e. I always want to be sure that every request has
// "/api" as part of it's path prefix without having to put it into
// every handle path in my various routing functions
func NewSubrouter(rtr *mux.Router) *mux.Router {
	sRtr := rtr.PathPrefix("/api").Subrouter()
	return sRtr
}
