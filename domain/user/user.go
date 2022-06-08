// Package user holds details about a person who is using the application
package user

import (
	"context"
	"net/http"

	"github.com/google/uuid"

	"github.com/gilcrest/diy-go-api/domain/errs"
	"github.com/gilcrest/diy-go-api/domain/org"
	"github.com/gilcrest/diy-go-api/domain/person"
	"github.com/gilcrest/diy-go-api/domain/secure"
)

// User holds details of a User from various providers
type User struct {
	// ID: unique identifier of the User
	ID uuid.UUID

	// ExternalID: unique external identifier of the User
	ExternalID secure.Identifier

	// username: unique (within an Org) username of the User
	Username string

	// org: Org user is associated with.
	Org org.Org

	// profile: The profile of the user
	Profile person.Profile
}

// NullUUID returns ID as uuid.NullUUID
func (u User) NullUUID() uuid.NullUUID {
	if u.ID == uuid.Nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{
		UUID:  u.ID,
		Valid: true,
	}
}

// IsValid determines whether the User has proper data to be considered valid
func (u User) IsValid() error {
	switch {
	case u.Org.ID == uuid.Nil:
		return errs.E(errs.Validation, "org ID cannot be nil")
	case u.ExternalID.String() == "":
		return errs.E(errs.Validation, "external ID cannot be empty")
	case u.Username == "":
		return errs.E(errs.Validation, "username cannot be empty")
	case u.Profile.FirstName == "":
		return errs.E(errs.Validation, "FirstName cannot be empty")
	case u.Profile.LastName == "":
		return errs.E(errs.Validation, "LastName cannot be empty")
	}
	return nil
}

type contextKey string

const contextKeyUser = contextKey("user")

// FromRequest gets the User from the request
func FromRequest(r *http.Request) (u User, err error) {
	var ok bool
	u, ok = r.Context().Value(contextKeyUser).(User)
	if !ok {
		return User{}, errs.E(errs.Internal, "User not set properly to context")
	}
	if err = u.IsValid(); err != nil {
		return User{}, err
	}
	return u, nil
}

// CtxWithUser sets the User to the given context
func CtxWithUser(ctx context.Context, u User) context.Context {
	return context.WithValue(ctx, contextKeyUser, u)
}
