package handlers

import "github.com/gorilla/mux"

// NewSubrouter adds any subRouters that you'd like to have as part of
// every request, i.e. I always want to be sure that every request has
// "/api" as part of it's path prefix without having to put it into
// every handle path in my various routing functions
func NewSubrouter(rtr *mux.Router) *mux.Router {
	sRtr := rtr.PathPrefix("/api").Subrouter()
	return sRtr
}
