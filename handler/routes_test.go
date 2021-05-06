package handler

import (
	"net/http"
	"os"
	"testing"

	"github.com/gilcrest/go-api-basic/domain/auth"

	qt "github.com/frankban/quicktest"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/datastore/datastoretest"
	"github.com/gilcrest/go-api-basic/datastore/pingstore"
	"github.com/gilcrest/go-api-basic/domain/auth/authtest"
	"github.com/gilcrest/go-api-basic/domain/logger"
	"github.com/gilcrest/go-api-basic/domain/random"
)

func TestNewMuxRouter(t *testing.T) {
	t.Run("typical", func(t *testing.T) {

		// initialize quickest checker
		c := qt.New(t)

		// initialize a zerolog Logger
		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)

		defaultDatastore, cleanup := datastoretest.NewDefaultDatastore(t, lgr)
		t.Cleanup(cleanup)

		// initialize MockTransactor for the moviestore
		mockTransactor := newMockTransactor(t)

		// initialize MockSelector for the moviestore
		mockSelector := newMockSelector(t)

		// initialize DefaultStringGenerator
		randomStringGenerator := random.DefaultStringGenerator{}

		// initialize DefaultMovieHandlers
		defaultMovieHandlers := DefaultMovieHandlers{
			RandomStringGenerator: randomStringGenerator,
			Transactor:            mockTransactor,
			Selector:              mockSelector,
		}

		// setup handlers
		createMovieHandler := NewCreateMovieHandler(defaultMovieHandlers)
		findMovieByIDHandler := NewFindMovieByIDHandler(defaultMovieHandlers)
		findAllMoviesHandler := NewFindAllMoviesHandler(defaultMovieHandlers)
		updateMovieHandler := NewUpdateMovieHandler(defaultMovieHandlers)
		deleteMovieHandler := NewDeleteMovieHandler(defaultMovieHandlers)
		readLoggerHandler := NewReadLoggerHandler()
		updateLoggerHandler := NewUpdateLoggerHandler()
		defaultPinger := pingstore.NewDefaultPinger(defaultDatastore)
		defaultPingHandler := DefaultPingHandler{
			Pinger: defaultPinger,
		}
		pingHandler := NewPingHandler(defaultPingHandler)
		handlers := Handlers{
			CreateMovieHandler:   createMovieHandler,
			FindMovieByIDHandler: findMovieByIDHandler,
			FindAllMoviesHandler: findAllMoviesHandler,
			UpdateMovieHandler:   updateMovieHandler,
			DeleteMovieHandler:   deleteMovieHandler,
			ReadLoggerHandler:    readLoggerHandler,
			UpdateLoggerHandler:  updateLoggerHandler,
			PingHandler:          pingHandler,
		}

		mw := Middleware{
			Logger:               lgr,
			AccessTokenConverter: authtest.NewMockAccessTokenConverter(t),
			Authorizer:           auth.DefaultAuthorizer{},
		}

		rtr := NewMuxRouterWithSubroutes()

		Routes(rtr, mw, handlers)

		// r holds the path and http method to be tested
		type r struct {
			PathTemplate string
			HTTPMethods  []string
		}

		// use a slice literal to create the routes in order of how
		// they are registered in NewMuxRouter
		wantRoutes := []r{
			{pathPrefix + moviesV1PathRoot, []string{http.MethodPost}},
			{pathPrefix + moviesV1PathRoot + extlIDPathDir, []string{http.MethodPut}},
			{pathPrefix + moviesV1PathRoot + extlIDPathDir, []string{http.MethodDelete}},
			{pathPrefix + moviesV1PathRoot + extlIDPathDir, []string{http.MethodGet}},
			{pathPrefix + moviesV1PathRoot, []string{http.MethodGet}},
			{pathPrefix + loggerV1PathRoot, []string{http.MethodGet}},
			{pathPrefix + loggerV1PathRoot, []string{http.MethodPut}},
			{pathPrefix + pingV1PathRoot, []string{http.MethodGet}},
		}

		// make a slice of r for use in the Walk function
		gotRoutes := make([]r, 0)

		// use gorilla/mux Walk function to walk the registered routes
		// routes will be added in order registered
		err := rtr.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			pathTemplate, err := route.GetPathTemplate()
			// check for errors from GetPathTemplate()
			c.Assert(err, qt.IsNil)

			methods, err := route.GetMethods()
			// check for errors from GetMethods()
			c.Assert(err, qt.IsNil)

			gotRoutes = append(gotRoutes, r{PathTemplate: pathTemplate, HTTPMethods: methods})

			return nil
		})

		// check for errors from Walk
		c.Assert(err, qt.IsNil)

		// assert that the routes from NewMuxRouter is equal to the
		// routes we want
		c.Assert(gotRoutes, qt.DeepEquals, wantRoutes)

	})
}
