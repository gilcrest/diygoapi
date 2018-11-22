package app

import "net/http"

// handleStdHeader middleware is used to add standard HTTP response headers
func (s *server) handleStdHeader(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			h.ServeHTTP(w, r) // call original
		})
}
