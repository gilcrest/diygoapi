package handlers

import (
	"log"
	"net/http"

	"github.com/gilcrest/go-API-template/pkg/config/env"
)

// Error represents a handler error. It provides methods for a HTTP status
// code and embeds the built-in error interface.
type Error interface {
	error
	Status() int
}

// HTTPStatusError represents an error with an associated HTTP status code.
type HTTPStatusError struct {
	Code int
	Err  error
}

// Allows HTTPStatusError to satisfy the error interface.
func (hse HTTPStatusError) Error() string {
	return hse.Err.Error()
}

// Status Returns an HTTP status code.
func (hse HTTPStatusError) Status() int {
	return hse.Code
}

// The Handler struct that takes a configured Env and a function matching
// our useful signature.
type Handler struct {
	Env *env.Env
	H   func(e *env.Env, w http.ResponseWriter, r *http.Request) error
}

// ServeHTTP allows Handler type to satisfy the http.Handler interface
func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := h.H(h.Env, w, r)
	if err != nil {
		switch e := err.(type) {
		case Error:
			// We can retrieve the status here and write out a specific
			// HTTP status code.
			log.Printf("HTTP %d - %s", e.Status(), e)
			http.Error(w, e.Error(), e.Status())
		default:
			// Any error types we don't specifically look out for default
			// to serving a HTTP 500
			http.Error(w, http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
		}
	}
}
