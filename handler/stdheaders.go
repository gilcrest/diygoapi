package handler

import "net/http"

// StdResponseHeader middleware is used to add
// standard HTTP response headers
func StdResponseHeader(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			h.ServeHTTP(w, r) // call original
		})
}
