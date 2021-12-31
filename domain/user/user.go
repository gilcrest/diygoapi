// Package user holds details about a person who is using the application
package user

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/org"
	"github.com/gilcrest/go-api-basic/domain/person"
)

// User holds details of a User from Google
type User struct {
	// ID: unique identifier of the User
	ID uuid.UUID

	// username: unique (within an Org) username of the User
	Username string

	// org: Org user is associated with.
	Org org.Org

	// profile: The profile of the user
	Profile person.Profile
}

// IsValid determines whether the User has proper data to be considered valid
func (u User) IsValid() bool {
	switch {
	case u.Username == "":
		return false
	case u.Profile.FirstName == "":
		return false
	case u.Profile.LastName == "":
		return false
	}
	return true
}

type contextKey string

const contextKeyUser = contextKey("user")

// FromRequest gets the User from the request
func FromRequest(r *http.Request) (User, error) {
	u, ok := r.Context().Value(contextKeyUser).(User)
	if !ok {
		return u, errs.E(errs.Internal, "User not set properly to context")
	}
	if !u.IsValid() {
		return u, errs.E(errs.Internal, "User empty in context")
	}
	return u, nil
}

// CtxWithUser sets the User to the given context
func CtxWithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, contextKeyUser, u)
}
