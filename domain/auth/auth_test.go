package auth_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/casbin/casbin/v2"
	qt "github.com/frankban/quicktest"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/domain/app"
	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/logger"
	"github.com/gilcrest/go-api-basic/domain/org"
	"github.com/gilcrest/go-api-basic/domain/person"
	"github.com/gilcrest/go-api-basic/domain/user"
)

func TestNewProvider(t *testing.T) {
	t.Run("google", func(t *testing.T) {
		c := qt.New(t)
		p := auth.NewProvider("GoOgLe")
		c.Assert(p, qt.Equals, auth.Google)
	})
	t.Run("apple", func(t *testing.T) {
		c := qt.New(t)
		p := auth.NewProvider("ApPlE")
		c.Assert(p, qt.Equals, auth.Apple)
	})
	t.Run("invalid", func(t *testing.T) {
		c := qt.New(t)
		p := auth.NewProvider("anything else!")
		c.Assert(p, qt.Equals, auth.Invalid)
	})
}

func TestProvider_String(t *testing.T) {
	t.Run("google", func(t *testing.T) {
		c := qt.New(t)
		p := auth.NewProvider("GoOgLe")
		provider := p.String()
		c.Assert(provider, qt.Equals, "google")
	})
	t.Run("apple", func(t *testing.T) {
		c := qt.New(t)
		p := auth.NewProvider("APPLe")
		provider := p.String()
		c.Assert(provider, qt.Equals, "apple")
	})
	t.Run("invalid", func(t *testing.T) {
		c := qt.New(t)
		p := auth.NewProvider("anything else")
		provider := p.String()
		c.Assert(provider, qt.Equals, "invalid_provider")
	})
}

func TestCasbinAuthorizer_Authorize(t *testing.T) {
	t.Run("valid user", func(t *testing.T) {
		c := qt.New(t)

		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)
		a := app.App{}
		u := user.User{
			ID:       uuid.Nil,
			Username: "dan@dangillis.dev",
			Org:      org.Org{},
			Profile:  person.Profile{},
		}
		adt := audit.Audit{
			App:    a,
			User:   u,
			Moment: time.Now(),
		}
		// initialize casbin enforcer (using config files for now, will migrate to db)
		casbinEnforcer, err := casbin.NewEnforcer("../../config/rbac_model.conf", "../../config/rbac_policy.csv")
		if err != nil {
			c.Fatal("casbin.NewEnforcer error")
		}
		ca := auth.CasbinAuthorizer{Enforcer: casbinEnforcer}

		// Authorize must be tested inside a handler as it uses mux.CurrentRoute
		testAuthorizeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err = ca.Authorize(lgr, r, adt)
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
}
