// Package auth is for authorization logic
package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/user"
)

const (
	// DefaultRealm is the realm used when one is not given explicitly
	DefaultRealm WWWAuthenticateRealm = "go-api-basic"

	// BearerTokenType is used in authorization to access a resource
	BearerTokenType string = "Bearer"

	contextKeyRealm = contextKey("realm")

	contextKeyAccessToken = contextKey("access-token")
)

type contextKey string

// WWWAuthenticateRealm is a description of a protected area, used
// in the WWW-Authenticate header
type WWWAuthenticateRealm string

// RealmFromRequest gets the realm from the request, if any
func RealmFromRequest(r *http.Request) (realm WWWAuthenticateRealm, ok bool) {
	if r == nil {
		return
	}
	return RealmFromCtx(r.Context())
}

// RealmFromCtx returns the realm from the context
func RealmFromCtx(ctx context.Context) (realm WWWAuthenticateRealm, ok bool) {
	realm, ok = ctx.Value(contextKeyRealm).(WWWAuthenticateRealm)
	return
}

// CtxWithRealm sets the Realm to the given context
func CtxWithRealm(ctx context.Context, realm WWWAuthenticateRealm) context.Context {
	return context.WithValue(ctx, contextKeyRealm, realm)
}

// AccessToken represents an access token found in an
// HTTP header, typically a Bearer token for Oauth2
type AccessToken struct {
	Token     string
	TokenType string
}

// NewAccessToken is an initializer for AccessToken
func NewAccessToken(token, tokenType string) AccessToken {
	return AccessToken{
		Token:     token,
		TokenType: tokenType,
	}
}

// NewGoogleOauth2Token returns a Google Oauth2 token given an AccessToken
func (at AccessToken) NewGoogleOauth2Token() *oauth2.Token {
	return &oauth2.Token{AccessToken: at.Token, TokenType: at.TokenType}
}

// AccessTokenFromRequest returns the access token from the request, if any
func AccessTokenFromRequest(r *http.Request) (at AccessToken, ok bool) {
	if r == nil {
		return
	}
	return AccessTokenFromCtx(r.Context())
}

// AccessTokenFromCtx returns the access token from the context, if any
func AccessTokenFromCtx(ctx context.Context) (at AccessToken, ok bool) {
	at, ok = ctx.Value(contextKeyAccessToken).(AccessToken)
	return
}

// CtxWithAccessToken sets the Access Token to the given context
func CtxWithAccessToken(ctx context.Context, at AccessToken) context.Context {
	return context.WithValue(ctx, contextKeyAccessToken, at)
}

// Authorizer and the Authorize method ensures a subject (user)
// can perform a particular action on an object. e.g. gilcrest can
// read (GET) the resource at the /ping path. This is obviously
// completely bogus right now, eventually need to look into something
// for ACL/RBAC
type Authorizer struct{}

// Authorize authorizes a subject (user) can perform a particular
// action on an object. e.g. gilcrest can read (GET) the resource
// at the /ping path. This is obviously completely bogus right now,
// eventually need to look into something like Casbin for ACL/RBAC
func (a Authorizer) Authorize(lgr zerolog.Logger, sub user.User, obj string, act string) error {

	const (
		moviesPath string = "/api/v1/movies"
		loggerPath string = "/api/v1/logger"
	)

	var authorized bool
	switch (strings.HasPrefix(obj, moviesPath) || strings.HasPrefix(obj, loggerPath)) && (act == http.MethodPost || act == http.MethodPut || act == http.MethodDelete || act == http.MethodGet) {
	case true:
		switch sub.Email {
		case "otto.maddox711@gmail.com":
			authorized = true
		}
	}

	if authorized {
		lgr.Debug().Str("sub", sub.Email).Str("obj", obj).Str("act", act).Msgf("Authorized (sub: %s, obj: %s, act: %s)", sub.Email, obj, act)
		return nil
	}

	lgr.Info().Str("sub", sub.Email).Str("obj", obj).Str("act", act).Msgf("Unauthorized (sub: %s, obj: %s, act: %s)", sub.Email, obj, act)

	// "In summary, a 401 Unauthorized response should be used for missing or
	// bad authentication, and a 403 Forbidden response should be used afterwards,
	// when the user is authenticated but isn’t authorized to perform the
	// requested operation on the given resource."
	// If the user has gotten here, they have gotten through authentication
	// but do have the right access, this they are Unauthorized
	return errs.NewUnauthorizedError(errors.New(fmt.Sprintf("user %s does not have %s permission for %s", sub.Email, act, obj)))
}

// AccessControlList (ACL) describes permissions for a given object
type AccessControlList struct {
	Subject string
	Object  string
	Action  string
}
