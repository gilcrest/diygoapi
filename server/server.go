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
//      I chose to log with middleware from zerolog
// - removed opencensus integration
//      I may eventually add something in for tracing, but for now, removing
//      opencensus as I have not worked with it and think it has moved to
//      opentelemetry anyway
// - removed health checkers
//      I will likely add these back, but removing to simplify for now
// - removed TLS
//      I am using Google Cloud Run which handles TLS for me. To keep this as
//      simple as possible, I am removing TLS for now

// Package server provides a preconfigured HTTP server.
package server

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"github.com/gilcrest/diygoapi"
	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/server/driver"
)

const pathPrefix string = "/api"

// Services are used by the application service handlers
type Services struct {
	OrgServicer            diygoapi.OrgServicer
	AppServicer            diygoapi.AppServicer
	RegisterUserService    diygoapi.RegisterUserServicer
	PingService            diygoapi.PingServicer
	LoggerService          diygoapi.LoggerServicer
	GenesisServicer        diygoapi.GenesisServicer
	AuthenticationServicer diygoapi.AuthenticationServicer
	AuthorizationServicer  diygoapi.AuthorizationServicer
	PermissionServicer     diygoapi.PermissionServicer
	RoleServicer           diygoapi.RoleServicer
	MovieServicer          diygoapi.MovieServicer
}

// Server represents an HTTP server.
type Server struct {
	router *mux.Router
	Driver driver.Server

	// all logging is done with a zerolog.Logger
	Logger zerolog.Logger

	// Addr optionally specifies the TCP address for the server to listen on,
	// in the form "host:port". If empty, ":http" (port 80) is used.
	// The service names are defined in RFC 6335 and assigned by IANA.
	// See net.Dial for details of the address format.
	Addr string

	// Services used by the various HTTP routes and middleware.
	Services
}

// New initializes a new Server and registers
// routes to the given router
func New(rtr *mux.Router, serverDriver driver.Server, lgr zerolog.Logger) *Server {
	s := &Server{router: rtr}
	s.Logger = lgr
	s.Driver = serverDriver

	// register routes to the router
	s.registerRoutes()

	return s
}

// ListenAndServe is a wrapper to use wherever http.ListenAndServe is used.
func (s *Server) ListenAndServe() error {
	const op errs.Op = "server/Server.ListenAndServe"
	if s.Addr == "" {
		return errs.E(op, errs.Internal, "Server Addr is empty")
	}
	if s.router == nil {
		return errs.E(op, errs.Internal, "Server router is nil")
	}
	if s.Driver == nil {
		return errs.E(op, errs.Internal, "Server driver is nil")
	}
	return s.Driver.ListenAndServe(s.Addr, s.router)
}

// Shutdown gracefully shuts down the server without interrupting any active connections.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.Driver.Shutdown(ctx)
}

// Driver implements the driver.Server interface. The zero value is a valid http.Server.
type Driver struct {
	Server http.Server
}

// NewDriver creates a Driver enfolding a http.Server with default timeouts.
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
	const op errs.Op = "server/decoderErr"

	switch {
	// If the request body is empty (io.EOF)
	// return an error
	case err == io.EOF:
		return errs.E(op, errs.InvalidRequest, "request body cannot be empty")
	// If the request body has malformed JSON (io.ErrUnexpectedEOF)
	// return an error
	case err == io.ErrUnexpectedEOF:
		return errs.E(op, errs.InvalidRequest, "malformed JSON")
	// return other errors
	case err != nil:
		return errs.E(op, err)
	}
	return nil
}
