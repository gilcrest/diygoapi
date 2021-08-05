// Copyright 2018 The Go Cloud Development Kit Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// @gilcrest edits - I have made a copy of the go-cloud server code and made
// the following changes:
//
// - removed requestlog.Logger
//      I chose to log with in a middleware from zerolog
// - removed opencensus integration
//      I may eventually add something in for tracing, but for now, removing
//      opencensus as I have not worked with it and think it has moved to
//      opentelemetry anyway
// - removed health checkers
//      I will likely add these back, but removing to simplify for now
// - removed TLS
//      I am using Google Cloud Run which handles TLS for me. To keep this as
//      simple as possible, I am removing TLS for now

// Package app provides a preconfigured HTTP server.
package app

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/app/driver"
	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/user"
	"github.com/gilcrest/go-api-basic/service"
)

const pathPrefix string = "/api"

// AccessTokenConverter interface converts an access token to a User
type AccessTokenConverter interface {
	Convert(ctx context.Context, token auth.AccessToken) (user.User, error)
}

// Authorizer interface authorizes access to a resource given
// a user and action
type Authorizer interface {
	Authorize(lgr zerolog.Logger, sub user.User, obj string, act string) error
}

// CreateMovieService creates a Movie
type CreateMovieService interface {
	Create(ctx context.Context, r *service.CreateMovieRequest, u user.User) (service.MovieResponse, error)
}

// UpdateMovieService is a service for updating a Movie
type UpdateMovieService interface {
	Update(ctx context.Context, r *service.UpdateMovieRequest, u user.User) (service.MovieResponse, error)
}

// DeleteMovieService is a service for deleting a Movie
type DeleteMovieService interface {
	Delete(ctx context.Context, extlID string) (service.DeleteMovieResponse, error)
}

// FindMovieService interface reads a Movie form the database
type FindMovieService interface {
	FindMovieByID(ctx context.Context, extlID string) (service.MovieResponse, error)
	FindAllMovies(ctx context.Context) ([]service.MovieResponse, error)
}

// LoggerService reads and updates the logger state
type LoggerService interface {
	Read() service.LoggerResponse
	Update(r *service.LoggerRequest) (service.LoggerResponse, error)
}

// PingService pings the database and responds whether it is up or down
type PingService interface {
	Ping(ctx context.Context, logger zerolog.Logger) service.PingResponse
}

// Server represents an HTTP server.
type Server struct {
	router *mux.Router
	driver driver.Server

	// all logging is done with a zerolog.Logger
	logger zerolog.Logger

	// Addr optionally specifies the TCP address for the server to listen on,
	// in the form "host:port". If empty, ":http" (port 80) is used.
	// The service names are defined in RFC 6335 and assigned by IANA.
	// See net.Dial for details of the address format.
	Addr string

	// Authorization
	AccessTokenConverter AccessTokenConverter
	Authorizer           Authorizer

	// Services used by the various HTTP routes.
	PingService        PingService
	LoggerService      LoggerService
	CreateMovieService CreateMovieService
	UpdateMovieService UpdateMovieService
	DeleteMovieService DeleteMovieService
	FindMovieService   FindMovieService
}

// ServerParams is the set of configuration parameters for a Server
type ServerParams struct {
	// Logger is used for app logging
	Logger zerolog.Logger

	// Driver serves HTTP requests.
	Driver driver.Server
}

// NewServerParams is an initializer for ServerParams
func NewServerParams(lgr zerolog.Logger, d driver.Server) *ServerParams {
	options := &ServerParams{
		Logger: lgr,
		Driver: d,
	}
	return options
}

// NewServer initializes a new Server and registers
// routes to the given router
func NewServer(r *mux.Router, params *ServerParams) (*Server, error) {
	s := &Server{router: r}
	if params == nil {
		return nil, errs.E("params must not be nil")
	}
	s.logger = params.Logger
	if params.Driver == nil {
		return nil, errs.E("params.Driver must not be nil")
	}
	s.driver = params.Driver

	s.routes()

	return s, nil
}

// ListenAndServe is a wrapper to use wherever http.ListenAndServe is used.
func (s *Server) ListenAndServe() error {
	if s.Addr == "" {
		return errs.E(errs.Internal, "Server Addr is empty")
	}
	if s.router == nil {
		return errs.E(errs.Internal, "Server router is nil")
	}
	if s.driver == nil {
		return errs.E(errs.Internal, "Server driver is nil")
	}
	return s.driver.ListenAndServe(s.Addr, s.router)
}

// Shutdown gracefully shuts down the server without interrupting any active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.driver == nil {
		return nil
	}
	return s.driver.Shutdown(ctx)
}

// Driver implements the driver.Server interface. The zero value is a valid http.Server.
type Driver struct {
	Server http.Server
}

// NewDriver creates a Driver with an http.Server with default timeouts.
func NewDriver() *Driver {
	return &Driver{
		Server: http.Server{
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
	}
}

// ListenAndServe sets the address and handler on Driver's http.Server,
// then calls ListenAndServe on it.
func (d *Driver) ListenAndServe(addr string, h http.Handler) error {
	d.Server.Addr = addr
	d.Server.Handler = h
	return d.Server.ListenAndServe()
}

// Shutdown gracefully shuts down the server without interrupting any active connections,
// by calling Shutdown on Driver's http.Server
func (d *Driver) Shutdown(ctx context.Context) error {
	return d.Server.Shutdown(ctx)
}

// NewMuxRouter initializes a gorilla/mux router and
// adds the /api subroute to it
func NewMuxRouter() *mux.Router {
	// initializer gorilla/mux router
	r := mux.NewRouter()

	// send Router through PathPrefix method to validate any standard
	// subroutes you may want for your APIs. e.g. I always want to be
	// sure that every request has "/api" as part of its path prefix
	// without having to put it into every handle path in my various
	// routing functions
	s := r.PathPrefix(pathPrefix).Subrouter()

	return s
}

// decoderErr is a convenience function to handle errors returned by
// json.NewDecoder(r.Body).Decode(&data) and return the appropriate
// error response
func decoderErr(err error) error {
	switch {
	// If the request body is empty (io.EOF)
	// return an error
	case err == io.EOF:
		return errs.E(errs.InvalidRequest, "Request Body cannot be empty")
	// If the request body has malformed JSON (io.ErrUnexpectedEOF)
	// return an error
	case err == io.ErrUnexpectedEOF:
		return errs.E(errs.InvalidRequest, "Malformed JSON")
	// return other errors
	case err != nil:
		return errs.E(err)
	}
	return nil
}
