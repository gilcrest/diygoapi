// Package auth is for user and application authorization logic
package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/errs"
)

// BearerTokenType is used in authorization to access a resource
const BearerTokenType string = "Bearer"

// Provider defines the provider of authorization (Google, Apple, auth0, etc.)
type Provider uint8

// Provider of authorization
//
// The app uses Oauth2 to authorize users with one of the following Providers
const (
	Invalid Provider = iota
	Google           // Google
	Apple            // Apple
)

func (p Provider) String() string {
	switch p {
	case Google:
		return "google"
	case Apple:
		return "apple"
	}
	return "invalid_provider"
}

// NewProvider initializes a Provider given a case-insensitive string
func NewProvider(s string) Provider {
	switch strings.ToLower(s) {
	case "google":
		return Google
	case "apple":
		return Apple
	}
	return Invalid
}

// CasbinAuthorizer holds the casbin.Enforcer struct
type CasbinAuthorizer struct {
	Enforcer *casbin.Enforcer
}

// Authorize ensures that a subject (user.User) can perform a
// particular action on an object. e.g. subject otto.maddox711@gmail.com
// can read (GET) the object (resource) at the /api/v1/movies path.
// Casbin is set up to use an RBAC (Role-Based Access Control) model
// Users with the admin role can *write* (GET, PUT, POST, DELETE).
// Users with the user role can only *read* (GET)
func (a CasbinAuthorizer) Authorize(lgr zerolog.Logger, r *http.Request, adt audit.Audit) error {
	// subject: Username
	sub := adt.User.Username

	// object: current route path
	route := mux.CurrentRoute(r)

	// CurrentRoute can return a nil if route not setup properly or
	// is being called outside the handler of the matched route
	if route == nil {
		return errs.E(errs.Unauthorized, "nil route returned from mux.CurrentRoute")
	}

	obj, err := route.GetPathTemplate()
	if err != nil {
		return errs.E(errs.Unauthorized, err)
	}

	// action: based on http method
	var act string
	switch r.Method {
	case http.MethodGet:
		act = "read"
	case http.MethodDelete:
		act = "delete"
	default:
		act = "write"
	}

	authorized, err := a.Enforcer.Enforce(sub, obj, act)
	if err != nil {
		return errs.E(errs.Unauthorized, err)
	}
	if !authorized {
		lgr.Info().Str("sub", sub).Str("obj", obj).Str("act", act).Msgf("Unauthorized (sub: %s, obj: %s, act: %s)", sub, obj, act)

		// "In summary, a 401 Unauthorized response should be used for missing or
		// bad authentication, and a 403 Forbidden response should be used afterwards,
		// when the user is authenticated but isn’t authorized to perform the
		// requested operation on the given resource."
		// If the user has gotten here, they have gotten through authentication
		// but do have the right access, this they are Unauthorized
		return errs.E(errs.Unauthorized, fmt.Sprintf("user %s does not have %s permission for %s", sub, act, obj))
	}

	lgr.Debug().Str("sub", sub).Str("obj", obj).Str("act", act).Msgf("Authorized (sub: %s, obj: %s, act: %s)", sub, obj, act)
	return nil
}
