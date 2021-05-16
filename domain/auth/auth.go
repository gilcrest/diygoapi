// Package auth is for authorization logic
package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/rs/zerolog"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/user"
)

// BearerTokenType is used in authorization to access a resource
const BearerTokenType string = "Bearer"

// NewAccessToken is an initializer for AccessToken
func NewAccessToken(token, tokenType string) AccessToken {
	return AccessToken{
		Token:     token,
		TokenType: tokenType,
	}
}

// AccessToken represents an access token found in an
// HTTP header, typically a Bearer token for Oauth2
type AccessToken struct {
	Token     string
	TokenType string
}

// NewGoogleOauth2Token returns a Google Oauth2 token given an AccessToken
func (at AccessToken) NewGoogleOauth2Token() *oauth2.Token {
	return &oauth2.Token{AccessToken: at.Token, TokenType: at.TokenType}
}

type contextKey string

const contextKeyAccessToken = contextKey("access-token")

// AccessTokenFromRequest gets the access token from the request
func AccessTokenFromRequest(r *http.Request) (AccessToken, error) {
	at, ok := r.Context().Value(contextKeyAccessToken).(AccessToken)
	if !ok {
		return at, errs.E(errs.Unauthenticated, errors.New("Access Token not set properly to context"))
	}
	if at.Token == "" {
		return at, errs.E(errs.Unauthenticated, errors.New("Access Token empty in context"))
	}
	return at, nil
}

// CtxWithAccessToken sets the Access Token to the given context
func CtxWithAccessToken(ctx context.Context, at AccessToken) context.Context {
	return context.WithValue(ctx, contextKeyAccessToken, at)
}

// AccessTokenConverter interface is used to convert an access token
// to a User
type AccessTokenConverter interface {
	Convert(ctx context.Context, token AccessToken) (user.User, error)
}

// Authorizer interface authorizes access to a resource given
// a user and action
type Authorizer interface {
	Authorize(lgr zerolog.Logger, sub user.User, obj string, act string) error
}

// DefaultAuthorizer struct satisfies the Authorizer interface.
// The DefaultAuthorizer.Authorize method ensures a subject (user)
// can perform a particular action on an object. e.g. gilcrest can
// read (GET) the resource at the /ping path. This is obviously
// completely bogus right now, eventually need to look into something
// like Casbin for ACL/RBAC
type DefaultAuthorizer struct{}

// Authorize authorizes a subject (user) can perform a particular
// action on an object. e.g. gilcrest can read (GET) the resource
// at the /ping path. This is obviously completely bogus right now,
// eventually need to look into something like Casbin for ACL/RBAC
func (a DefaultAuthorizer) Authorize(lgr zerolog.Logger, sub user.User, obj string, act string) error {

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
	return errs.E(errs.Unauthorized, errors.New(fmt.Sprintf("user %s does not have %s permission for %s", sub.Email, act, obj)))
}

// AccessControlList (ACL) describes permissions for a given object
type AccessControlList struct {
	Subject string
	Object  string
	Action  string
}
