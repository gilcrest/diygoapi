package server

import (
	"github.com/gilcrest/envy"
	"github.com/gilcrest/errors"
	"github.com/rs/zerolog"
)

// Server struct contains the environment (envy.Env) and additional methods
// for running our HTTP server
type Server struct {
	*envy.Env
}

// NewServer is a constructor for the Server struct
// Sets up the struct and registers routes
func NewServer(name envy.Name, lvl zerolog.Level) (*Server, error) {
	const op errors.Op = "server/NewServer"

	// call constructor for Env struct from env module
	env, err := envy.NewEnv(name, lvl)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// Use type embedding to make envy.Env struct part of Server struct
	server := &Server{env}

	// routes registers handlers to the Server router
	err = server.routes()
	if err != nil {
		return nil, errors.E(op, err)
	}

	return server, nil
}
