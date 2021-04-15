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

// BearerTokenType is used in authorization to access a
// resource
const BearerTokenType string = "Bearer"

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

// AccessTokenConverter interface is used to convert an access token
// to a User
type AccessTokenConverter interface {
	Convert(ctx context.Context, token AccessToken) (user.User, error)
}

// Authorizer interface authorizes access to a resource given
// a user and action
type Authorizer interface {
	Authorize(ctx context.Context, sub user.User, obj string, act string) error
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
func (a DefaultAuthorizer) Authorize(ctx context.Context, sub user.User, obj string, act string) error {
	lgr := *zerolog.Ctx(ctx)

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
		lgr.Info().Str("sub", sub.Email).Str("obj", obj).Str("act", act).Msgf("Authorization Granted")
		return nil
	}

	lgr.Info().Str("sub", sub.Email).Str("obj", obj).Str("act", act).Msgf("Authorization Denied")

	// "In summary, a 401 Unauthorized response should be used for missing or
	// bad authentication, and a 403 Forbidden response should be used afterwards,
	// when the user is authenticated but isn’t authorized to perform the
	// requested operation on the given resource."
	// If the user has gotten here, they have gotten through authentication
	// but do have the right access, this they are Unauthorized
	return errs.E(errs.Unauthorized, errors.New(fmt.Sprintf("user %s does not have %s permission for %s", sub.Email, act, obj)))
}

type contextKey string

const contextKeyAccessToken = contextKey("access-token")

// FromRequest gets the access token from the request
func FromRequest(r *http.Request) (AccessToken, error) {
	at := AccessToken{}
	// retrieve the context from the http.Request
	ctx := r.Context()

	at, ok := ctx.Value(contextKeyAccessToken).(AccessToken)
	if !ok {
		return at, errs.E(errs.Unauthenticated, errors.New("Access Token not set properly to context"))
	}
	if at.Token == "" {
		return at, errs.E(errs.Unauthenticated, errors.New("Access Token empty in context"))
	}
	return at, nil
}

// SetAccessToken2Context sets the Access Token to the given context
func SetAccessToken2Context(ctx context.Context, token, tokenType string) context.Context {
	at := AccessToken{
		Token:     token,
		TokenType: tokenType,
	}

	return context.WithValue(ctx, contextKeyAccessToken, at)
}

// AccessControlList (ACL) describes permissions for a given object
type AccessControlList struct {
	Subject string
	Object  string
	Action  string
}
