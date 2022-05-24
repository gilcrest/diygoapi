package service_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"github.com/gilcrest/diy-go-api/datastore/datastoretest"
	"github.com/gilcrest/diy-go-api/domain/logger"
	"github.com/gilcrest/diy-go-api/service"
)

func TestDBAuthorizer_Authorize(t *testing.T) {
	t.Run("valid user", func(t *testing.T) {
		c := qt.New(t)

		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
		ds, cleanup := datastoretest.NewDatastore(t)
		c.Cleanup(cleanup)

		adt := findTestAudit(context.Background(), t, ds)

		dba := service.DBAuthorizer{Datastorer: ds}

		// Authorize must be tested inside a handler as it uses mux.CurrentRoute
		testAuthorizeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := dba.Authorize(lgr, r, adt)
			c.Assert(err, qt.IsNil)
		})

		rr := httptest.NewRecorder()

		rtr := mux.NewRouter()
		rtr.Handle("/api/v1/ping", testAuthorizeHandler).Methods(http.MethodGet)
		rtr.ServeHTTP(rr, req)

		// If there is any issues with the Access Token, the body
		// should be empty and the status code should be 401
		c.Assert(rr.Code, qt.Equals, http.StatusOK)

	})
	t.Run("valid user with path vars", func(t *testing.T) {
		c := qt.New(t)

		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/orgs/123", nil)
		ds, cleanup := datastoretest.NewDatastore(t)
		t.Cleanup(cleanup)

		adt := findTestAudit(context.Background(), t, ds)

		dba := service.DBAuthorizer{Datastorer: ds}

		// Authorize must be tested inside a handler as it uses mux.CurrentRoute
		testAuthorizeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err := dba.Authorize(lgr, r, adt)
			c.Assert(err, qt.IsNil)
		})

		rr := httptest.NewRecorder()

		rtr := mux.NewRouter()
		rtr.Handle("/api/v1/orgs/{extlID}", testAuthorizeHandler).Methods(http.MethodGet)
		rtr.ServeHTTP(rr, req)

		// If there is any issues with the Access Token, the body
		// should be empty and the status code should be 401
		c.Assert(rr.Code, qt.Equals, http.StatusOK)

	})
}
