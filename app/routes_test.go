package app

import (
	"net/http"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gorilla/mux"
)

func TestNewMuxRouter(t *testing.T) {
	t.Run("typical", func(t *testing.T) {

		// initialize quickest checker
		c := qt.New(t)

		rtr := NewMuxRouter()

		s := Server{
			router: rtr,
		}

		s.routes()

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
