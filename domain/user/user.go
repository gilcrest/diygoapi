// Package user holds details about a person who is using the application
package user

import (
	"context"
	"net/http"

	"github.com/gilcrest/go-api-basic/domain/errs"
)

// User holds details of a User from Google
type User struct {
	// Email: The user's email address.
	Email string

	// LastName: The user's last name.
	LastName string

	// FirstName: The user's first name.
	FirstName string

	// FullName: The user's full name.
	FullName string

	// HostedDomain: The hosted domain e.g. example.com if the user
	// is Google apps user.
	HostedDomain string

	// PictureURL: URL of the user's picture image.
	PictureURL string

	// ProfileLink: URL of the profile page.
	ProfileLink string
}

// IsValid determines whether or not the User has proper
// data to be considered valid
func (u User) IsValid() bool {
	switch {
	case u.Email == "":
		return false
	case u.FirstName == "":
		return false
	case u.LastName == "":
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
