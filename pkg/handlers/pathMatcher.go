package handlers

import (
	"github.com/gorilla/mux"
)

func PathMatcher(h *UserHandler, rtr *mux.Router) *mux.Router {

	// match only POST requests on /api/appUser/create
	// This is the original (v1) version for the API and the response for this will never change
	//  with versioning in order to maintain a stable contract
	// func (r *Router) HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *Route
	rtr.HandleFunc("/api/appUser/create", h.CreateUserHandler).
		Methods("POST").
		Headers("Content-Type", "application/json")

	// match only POST requests on /api/v1/appUser/create
	// func (r *Router) HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *Route
	//rtr.HandleFunc("/api/v1/appUser/create", createUserHandler).
	//	Methods("POST").
	//	Headers("Content-Type", "application/json")

	return rtr
}
