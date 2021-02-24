// Package authtest provides testing helper functions for the
// auth package
package authtest

import (
	"context"
	"testing"

	"github.com/gilcrest/go-api-basic/domain/user/usertest"

	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/user"
)

func NewMockAuthorizer(t *testing.T) MockAuthorizer {
	return MockAuthorizer{t: t}
}

type MockAuthorizer struct {
	t *testing.T
}

func (ma MockAuthorizer) Authorize(ctx context.Context, sub user.User, obj string, act string) error {
	ma.t.Helper()

	return nil
}

func NewAccessToken(t *testing.T) auth.AccessToken {
	t.Helper()

	return auth.AccessToken{Token: "abc123def1", TokenType: auth.BearerTokenType}
}

func NewMockAccessTokenConverter(t *testing.T) MockAccessTokenConverter {
	return MockAccessTokenConverter{t: t}
}

type MockAccessTokenConverter struct {
	t *testing.T
}

func (m MockAccessTokenConverter) Convert(ctx context.Context, token auth.AccessToken) (user.User, error) {
	m.t.Helper()

	return usertest.NewUser(m.t), nil
}
