package handlers

import "github.com/gorilla/mux"

// NewSubrouter adds any subRouters that you'd like to have as part of
// every request, i.e. I always want to be sure that every request
// has application/json as a Content-Type header
func NewSubrouter(rtr *mux.Router) *mux.Router {
	sRtr := rtr.PathPrefix("/api").Subrouter()
	return sRtr
}
