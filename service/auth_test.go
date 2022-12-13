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

	"github.com/gilcrest/diygoapi/logger"
	"github.com/gilcrest/diygoapi/service"
	"github.com/gilcrest/diygoapi/sqldb/sqldbtest"
)

func TestDBAuthorizer_Authorize(t *testing.T) {
	t.Run("valid user", func(t *testing.T) {
		c := qt.New(t)

		const pingPath string = "/api/v1/ping"

		lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)
		req := httptest.NewRequest(http.MethodGet, pingPath, nil)
		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		// start db txn using pgxpool
		ctx := context.Background()
		tx, err := db.BeginTx(ctx)
		if err != nil {
			c.Fatalf("BeginTx() error = %v", err)
		}
		c.Cleanup(func() { _ = db.RollbackTx(ctx, tx, err) })

		adt := findTestAudit(context.Background(), c, tx)

		dba := service.DBAuthorizationService{Datastorer: db}

		// Authorize must be tested inside a handler as it uses mux.CurrentRoute
		testAuthorizeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err = dba.Authorize(r, lgr, adt)
			c.Assert(err, qt.IsNil)
		})

		rr := httptest.NewRecorder()

		rtr := mux.NewRouter()
		rtr.Handle(pingPath, testAuthorizeHandler).Methods(http.MethodGet)
		rtr.ServeHTTP(rr, req)

		// If there is any issues with the Access Token, the body
		// should be empty and the status code should be 401
		c.Assert(rr.Code, qt.Equals, http.StatusOK)

	})
	t.Run("valid user with path vars", func(t *testing.T) {
		c := qt.New(t)

		lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/orgs/123", nil)
		db, cleanup := sqldbtest.NewDB(t)
		t.Cleanup(cleanup)

		// start db txn using pgxpool
		ctx := context.Background()
		tx, err := db.BeginTx(ctx)
		if err != nil {
			c.Fatalf("BeginTx() error = %v", err)
		}
		c.Cleanup(func() { _ = db.RollbackTx(ctx, tx, err) })

		adt := findTestAudit(context.Background(), c, tx)

		dba := service.DBAuthorizationService{Datastorer: db}

		// Authorize must be tested inside a handler as it uses mux.CurrentRoute
		testAuthorizeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err = dba.Authorize(r, lgr, adt)
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
