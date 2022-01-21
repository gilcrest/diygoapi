// Package authtest provides testing helper functions for the
// auth package
package authtest

import (
	"context"
	"testing"

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
