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

// NewMockAuthorizer is an initializer for MockAuthorizer
func NewMockAuthorizer(t *testing.T) MockAuthorizer {
	return MockAuthorizer{t: t}
}

// MockAuthorizer mocks authorizing access for
// a user to a given object to perform a given action.
type MockAuthorizer struct {
	t *testing.T
}

// Authorize mocks authorizing access for
// a user to a given object to perform a given action. Authorize
// never returns an error, thus everything sent in is always authorized
func (ma MockAuthorizer) Authorize(ctx context.Context, sub user.User, obj string, act string) error {
	ma.t.Helper()

	return nil
}

// NewAccessToken returns a mock auth.AccessToken
func NewAccessToken(t *testing.T) auth.AccessToken {
	t.Helper()

	return auth.AccessToken{Token: "abc123def1", TokenType: auth.BearerTokenType}
}

// NewMockAccessTokenConverter is an initializer for a MockAccessTokenConverter
func NewMockAccessTokenConverter(t *testing.T) MockAccessTokenConverter {
	return MockAccessTokenConverter{t: t}
}

// MockAccessTokenConverter mocks converting an auth.AccessToken to a user.User
type MockAccessTokenConverter struct {
	t *testing.T
}

// Convert returns a static test user.User
func (m MockAccessTokenConverter) Convert(ctx context.Context, token auth.AccessToken) (user.User, error) {
	m.t.Helper()

	return usertest.NewUser(m.t), nil
}
