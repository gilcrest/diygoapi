package server

import (
	"github.com/gilcrest/env"
	"github.com/gilcrest/errors"
	"github.com/rs/zerolog"
)

// Server struct contains the environment and additional methods
// for running our HTTP server
type Server struct {
	*env.Env
}

// NewServer is a constructor for the Server struct
// Sets up the struct and registers routes
func NewServer(name env.Name, lvl zerolog.Level) (*Server, error) {
	const op errors.Op = "server/NewServer"

	// call constructor for Env struct from env module
	env, err := env.NewEnv(name, lvl)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// Use type embedding to make env.Env struct part of Server struct
	server := &Server{env}

	// routes registers handlers to the Server router
	err = server.routes()
	if err != nil {
		return nil, errors.E(op, err)
	}

	return server, nil
}
