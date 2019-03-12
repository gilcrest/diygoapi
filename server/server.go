package server

import (
	"net/http"

	"github.com/gilcrest/errors"
	"github.com/gilcrest/srvr"
	"github.com/gilcrest/srvr/datastore"
	"github.com/rs/zerolog"
)

// Server struct is a pointer to srvr.Server
// allows additional local methods to be added
type Server struct {
	*srvr.Server
}

// handleRespHeader middleware is used to add standard HTTP response headers
func (s *Server) handleRespHeader(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			h.ServeHTTP(w, r) // call original
		})
}

// NewServer is a constructor for the Server struct
// Sets up the struct and registers routes
func NewServer(lvl zerolog.Level) (*Server, error) {
	const op errors.Op = "server/NewServer"

	// call constructor for Server struct from srvr module
	srvr, err := srvr.NewServer(lvl)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// Use type embedding to make srvr.Server struct part of
	// local Server struct
	server := &Server{srvr}

	// Use Datastore(DS) Option method to initialize logging
	// database
	err = server.DS.Option(datastore.InitLogDB())
	if err != nil {
		return nil, errors.E(op, err)
	}

	// routes registers handlers to the Server router
	err = server.routes()
	if err != nil {
		return nil, errors.E(op, err)
	}

	return server, nil
}
