package app

import (
	"net/http"

	"github.com/gilcrest/errors"
)

// handleStdHeader middleware is used to add standard HTTP response headers
func (s *server) handleStdHeader(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			h.ServeHTTP(w, r) // call original
		})
}

func (s *server) handleAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			err := errors.Str("fake")

			if err != nil {
				err = errors.HTTPErr{
					Code: http.StatusUnauthorized,
					Kind: errors.Permission,
					Err:  err,
				}
				errors.HTTPError(w, err)
				return
			}
			h.ServeHTTP(w, r) // call original
		})
}
