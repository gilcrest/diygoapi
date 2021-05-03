// Package handler is for all the application handlers and routing
package handler

import (
	"io"
	"net/http"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/pkg/errors"
)

// Handlers is a bundled set of all the application's HTTP handlers
type Handlers struct {
	PingHandler          PingHandler
	ReadLoggerHandler    ReadLoggerHandler
	UpdateLoggerHandler  UpdateLoggerHandler
	CreateMovieHandler   CreateMovieHandler
	FindMovieByIDHandler FindMovieByIDHandler
	FindAllMoviesHandler FindAllMoviesHandler
	UpdateMovieHandler   UpdateMovieHandler
	DeleteMovieHandler   DeleteMovieHandler
}

// ReadLoggerHandler is a Handler that reads the current state of the
// app logger
type ReadLoggerHandler http.Handler

// UpdateLoggerHandler is a Handler that reads the current state of the
// app logger
type UpdateLoggerHandler http.Handler

// CreateMovieHandler is a Handler that creates a Movie
type CreateMovieHandler http.Handler

// UpdateMovieHandler is a Handler that updates a Movie
type UpdateMovieHandler http.Handler

// DeleteMovieHandler is a Handler that deletes a Movie
type DeleteMovieHandler http.Handler

// FindMovieByIDHandler is a Handler finds a Movie by ID
type FindMovieByIDHandler http.Handler

// FindAllMoviesHandler is a Handler that returns the entire set of Movies
type FindAllMoviesHandler http.Handler

// DecoderErr is a convenience function to handle errors returned by
// json.NewDecoder(r.Body).Decode(&data) and return the appropriate
// error response
func DecoderErr(err error) error {
	switch {
	// If the request body is empty (io.EOF)
	// return an error
	case err == io.EOF:
		return errs.E(errs.InvalidRequest, errors.New("Request Body cannot be empty"))
	// If the request body has malformed JSON (io.ErrUnexpectedEOF)
	// return an error
	case err == io.ErrUnexpectedEOF:
		return errs.E(errs.InvalidRequest, errors.New("Malformed JSON"))
	// return all other errors
	case err != nil:
		return errs.E(err)
	}
	return nil
}
