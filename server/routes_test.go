package server

import (
	"net/http"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gorilla/mux"
)

func TestNewMuxRouter(t *testing.T) {
	t.Run("all routes", func(t *testing.T) {

		// initialize quickest checker
		c := qt.New(t)

		rtr := NewMuxRouter()

		s := Server{
			router: rtr,
		}

		s.registerRoutes()

		// r holds the path and http method to be tested
		type r struct {
			PathTemplate string
			HTTPMethods  []string
		}

		// use a slice literal to create the routes in order of how
		// they are registered in NewMuxRouter
		wantRoutes := []r{
			{PathTemplate: pathPrefix + moviesV1PathRoot, HTTPMethods: []string{http.MethodPost}},
			{PathTemplate: pathPrefix + moviesV1PathRoot + extlIDPathDir, HTTPMethods: []string{http.MethodPut}},
			{PathTemplate: pathPrefix + moviesV1PathRoot + extlIDPathDir, HTTPMethods: []string{http.MethodDelete}},
			{PathTemplate: pathPrefix + moviesV1PathRoot + extlIDPathDir, HTTPMethods: []string{http.MethodGet}},
			{PathTemplate: pathPrefix + moviesV1PathRoot, HTTPMethods: []string{http.MethodGet}},
			{PathTemplate: pathPrefix + orgsV1PathRoot, HTTPMethods: []string{http.MethodPost}},
			{PathTemplate: pathPrefix + orgsV1PathRoot + extlIDPathDir, HTTPMethods: []string{http.MethodPut}},
			{PathTemplate: pathPrefix + orgsV1PathRoot + extlIDPathDir, HTTPMethods: []string{http.MethodDelete}},
			{PathTemplate: pathPrefix + orgsV1PathRoot, HTTPMethods: []string{http.MethodGet}},
			{PathTemplate: pathPrefix + orgsV1PathRoot + extlIDPathDir, HTTPMethods: []string{http.MethodGet}},
			{PathTemplate: pathPrefix + appsV1PathRoot, HTTPMethods: []string{http.MethodPost}},
			{PathTemplate: pathPrefix + usersV1PathRoot, HTTPMethods: []string{http.MethodPost}},
			{PathTemplate: pathPrefix + loggerV1PathRoot, HTTPMethods: []string{http.MethodGet}},
			{PathTemplate: pathPrefix + loggerV1PathRoot, HTTPMethods: []string{http.MethodPut}},
			{PathTemplate: pathPrefix + pingV1PathRoot, HTTPMethods: []string{http.MethodGet}},
			{PathTemplate: pathPrefix + permissionV1PathRoot, HTTPMethods: []string{http.MethodPost}},
			{PathTemplate: pathPrefix + permissionV1PathRoot, HTTPMethods: []string{http.MethodGet}},
			{PathTemplate: pathPrefix + permissionV1PathRoot + extlIDPathDir, HTTPMethods: []string{http.MethodDelete}},
			{PathTemplate: pathPrefix + genesisV1PathRoot, HTTPMethods: []string{http.MethodPost}},
			{PathTemplate: pathPrefix + genesisV1PathRoot, HTTPMethods: []string{http.MethodGet}},
		}

		// make a slice of r for use in the Walk function
		gotRoutes := make([]r, 0)

		// use gorilla/mux Walk function to walk the registered routes.
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
