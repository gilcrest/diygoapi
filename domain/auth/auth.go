// Package auth is for user and application authorization logic
package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/casbin/casbin/v2"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"

	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/errs"
)

const (
	// BearerTokenType is used in authorization to access a resource
	BearerTokenType       string = "Bearer"
	contextKeyAccessToken        = contextKey("access-token")
)

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

type contextKey string

// Oauth2TokenFromRequest returns the Oauth2 token from the request context, if any
func Oauth2TokenFromRequest(r *http.Request) (ot oauth2.Token, ok bool) {
	if r == nil {
		return
	}
	return OAuth2TokenFromCtx(r.Context())
}

// OAuth2TokenFromCtx returns the Oauth2 token from the context, if any
func OAuth2TokenFromCtx(ctx context.Context) (ot oauth2.Token, ok bool) {
	ot, ok = ctx.Value(contextKeyAccessToken).(oauth2.Token)
	return
}

// CtxWithOauth2Token sets the Oauth2 Token to the given context
func CtxWithOauth2Token(ctx context.Context, ot oauth2.Token) context.Context {
	return context.WithValue(ctx, contextKeyAccessToken, ot)
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
