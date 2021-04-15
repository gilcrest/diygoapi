package handler

import (
	"net/http"
	"os"
	"testing"

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

		// initialize mockAccessTokenConverter
		mockAccessTokenConverter := authtest.NewMockAccessTokenConverter(t)

		// initialize DefaultStringGenerator
		randomStringGenerator := random.DefaultStringGenerator{}

		// initialize DefaultMovieHandlers
		defaultMovieHandlers := DefaultMovieHandlers{
			RandomStringGenerator: randomStringGenerator,
			AccessTokenConverter:  mockAccessTokenConverter,
			Authorizer:            authtest.NewMockAuthorizer(t),
			Transactor:            mockTransactor,
			Selector:              mockSelector,
		}

		// setup handlers
		createMovieHandler := ProvideCreateMovieHandler(defaultMovieHandlers)
		findMovieByIDHandler := ProvideFindMovieByIDHandler(defaultMovieHandlers)
		findAllMoviesHandler := ProvideFindAllMoviesHandler(defaultMovieHandlers)
		updateMovieHandler := ProvideUpdateMovieHandler(defaultMovieHandlers)
		deleteMovieHandler := ProvideDeleteMovieHandler(defaultMovieHandlers)
		defaultLoggerHandlers := DefaultLoggerHandlers{
			AccessTokenConverter: mockAccessTokenConverter,
			Authorizer:           authtest.NewMockAuthorizer(t),
		}
		readLoggerHandler := NewReadLoggerHandler(defaultLoggerHandlers)
		updateLoggerHandler := NewUpdateLoggerHandler(defaultLoggerHandlers)
		defaultPinger := pingstore.NewDefaultPinger(defaultDatastore)
		defaultPingHandler := DefaultPingHandler{
			Pinger: defaultPinger,
		}
		pingHandler := ProvidePingHandler(defaultPingHandler)
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

		// get a new router
		router := NewMuxRouter(lgr, handlers)

		// r holds the path and http method to be tested
		type r struct {
			PathTemplate string
			HTTPMethods  []string
		}

		// use a slice literal to create the routes in order of how
		// they are registered in NewMuxRouter
		wantRoutes := []r{
			{pathPrefix + moviesV1PathRoot, []string{http.MethodPost}},
			{pathPrefix + moviesV1PathRoot + "/{extlID}", []string{http.MethodPut}},
			{pathPrefix + moviesV1PathRoot + "/{extlID}", []string{http.MethodDelete}},
			{pathPrefix + moviesV1PathRoot + "/{extlID}", []string{http.MethodGet}},
			{pathPrefix + moviesV1PathRoot, []string{http.MethodGet}},
			{pathPrefix + loggerV1PathRoot, []string{http.MethodGet}},
			{pathPrefix + loggerV1PathRoot, []string{http.MethodPut}},
			{pathPrefix + "/v1/ping", []string{http.MethodGet}},
		}

		// make a slice of r for use in the Walk function
		gotRoutes := make([]r, 0)

		// use gorilla/mux Walk function to walk the registered routes
		// routes will be added in order registered
		err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
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
