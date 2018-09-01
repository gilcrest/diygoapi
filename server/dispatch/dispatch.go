package dispatch

import (
	"github.com/gilcrest/go-API-template/db"
	"github.com/gilcrest/go-API-template/env"
	eh "github.com/gilcrest/go-API-template/server/errorHandler"
	"github.com/gilcrest/go-API-template/server/handler"
	"github.com/gilcrest/go-API-template/server/middleware"
	"github.com/gilcrest/httplog"
	"github.com/gorilla/mux"
)

// Dispatch is a way of organizing routing to handlers (versioning as well)
func Dispatch(env *env.Env, rtr *mux.Router) *mux.Router {

	log := env.Logger

	logdb, err := env.DS.DB(db.LogDB)
	if err != nil {
		// TODO - bogus...
		return nil
	}

	// initialize new instance of APIAudit
	audit := new(middleware.APIAudit)

	// match only POST requests on /api/appuser
	// This is the original (v1) version for the API and the response for this
	// will never change with versioning in order to maintain a stable contract
	rtr.Handle("/appuser",
		middleware.Adapt(eh.ErrHandler{Env: env, H: handler.CreateUser},
			middleware.LogRequest(env, audit),
			middleware.LogResponse(env, audit))).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// match only POST requests on /api/v1/appuser
	rtr.Handle("/v1/appuser",
		middleware.Adapt(eh.ErrHandler{Env: env, H: handler.CreateUser},
			middleware.LogRequest(env, audit),
			middleware.LogResponse(env, audit))).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// match only POST requests on /api/login
	rtr.Handle("/login",
		middleware.Adapt(eh.ErrHandler{Env: env, H: handler.LoginHandler},
			middleware.LogRequest(env, audit),
			middleware.LogResponse(env, audit))).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// match only POST requests on /api/login
	rtr.Handle("/client",
		httplog.Adapt(eh.ErrHandler{Env: env, H: handler.CreateClientHandler},
			httplog.HTTPLog(log, logdb, nil))).
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
