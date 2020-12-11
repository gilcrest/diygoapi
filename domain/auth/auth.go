package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"

	"golang.org/x/oauth2"
	googleoauth "google.golang.org/api/oauth2/v2"

	"github.com/pkg/errors"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/user"
)

// AccessToken represents an Oauth2 access token found
// in an HTTP header as a Bearer token
type AccessToken struct {
	Token     string
	TokenType string
}

// NewGoogleOauth2Token returns a Google Oauth2 token given an AccessToken
func (at AccessToken) NewGoogleOauth2Token() *oauth2.Token {
	return &oauth2.Token{AccessToken: at.Token, TokenType: at.TokenType}
}

// newUser initializes the user.User struct given a Userinfo struct
// from Google
func newUser(userinfo *googleoauth.Userinfo) *user.User {
	return &user.User{
		Email:     userinfo.Email,
		LastName:  userinfo.FamilyName,
		FirstName: userinfo.GivenName,
		FullName:  userinfo.Name,
		//Gender:       userinfo.Gender,
		HostedDomain: userinfo.Hd,
		PictureURL:   userinfo.Picture,
		ProfileLink:  userinfo.Link,
	}
}

// Authorizer interface authorizes access to a resource given
// a user and action
type Authorizer interface {
	Authorize(ctx context.Context, sub *user.User, obj string, act string) error
}

// Auth struct satisfies the Authorizer interface
type Auth struct{}

// Authorize authorizes a subject (user) can perform a particular
// action on an object. e.g. gilcrest can read (GET) the resource
// at the /ping path. This is obviously completely bogus right now,
// eventually need to look into something like Casbin for ACL/RBAC
func (a Auth) Authorize(ctx context.Context, sub *user.User, obj string, act string) error {
	logger := *zerolog.Ctx(ctx)

	const (
		ping   string = "/api/v1/ping"
		movies string = "/api/v1/movies"
	)

	var authorized bool
	switch obj == ping && act == http.MethodGet {
	case true:
		switch sub.Email {
		case "gilcrest@gmail.com":
			authorized = true
		}
	}

	switch obj == movies && act == http.MethodPost || act == http.MethodPut || act == http.MethodDelete || act == http.MethodGet {
	case true:
		switch sub.Email {
		case "gilcrest@gmail.com":
			authorized = true
		}
	}

	if authorized {
		logger.Info().Str("sub", sub.Email).Str("obj", obj).Str("act", act).Msgf("Authorization Granted")
		return nil
	}

	logger.Info().Str("sub", sub.Email).Str("obj", obj).Str("act", act).Msgf("Authorization Denied")

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
