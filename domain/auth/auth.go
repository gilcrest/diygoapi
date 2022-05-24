// Package auth is for user and application authorization logic
package auth

import (
	"strings"

	"github.com/google/uuid"

	"github.com/gilcrest/diy-go-api/domain/errs"
	"github.com/gilcrest/diy-go-api/domain/secure"
	"github.com/gilcrest/diy-go-api/domain/user"
)

// BearerTokenType is used in authorization to access a resource
const BearerTokenType string = "Bearer"

// Provider defines the provider of authorization (Google, Apple, auth0, etc.)
type Provider uint8

// Provider of authorization
//
// The app uses Oauth2 to authorize users with one of the following Providers
const (
	Invalid Provider = iota
	Google           // Google
	Apple            // Apple
)

func (p Provider) String() string {
	switch p {
	case Google:
		return "google"
	case Apple:
		return "apple"
	}
	return "invalid_provider"
}

// ParseProvider initializes a Provider given a case-insensitive string
func ParseProvider(s string) Provider {
	switch strings.ToLower(s) {
	case "google":
		return Google
	case "apple":
		return Apple
	}
	return Invalid
}

// Permission stores an approval of a mode of access to a resource.
type Permission struct {
	// The unique ID for the Permission.
	ID uuid.UUID `json:"-"`
	// Unique External ID to be given to outside callers.
	ExternalID secure.Identifier `json:"external_id"`
	// A human-readable string which represents a resource (e.g. an HTTP route or document, etc.).
	Resource string `json:"resource"`
	// A string representing the action taken on the resource (e.g. POST, GET, edit, etc.)
	Operation string `json:"operation"`
	// A description of what the permission is granting, e.g. "grants ability to edit a billing document".
	Description string `json:"description"`
	// A boolean denoting whether the permission is active (true) or not (false).
	Active bool `json:"active"`
}

// IsValid determines if the Permission is valid
func (p Permission) IsValid() error {
	switch {
	case p.ID == uuid.Nil:
		return errs.E(errs.Validation, "ID is required")
	case p.ExternalID.String() == "":
		return errs.E(errs.Validation, "External ID is required")
	case p.Resource == "":
		return errs.E(errs.Validation, "Resource is required")
	case p.Description == "":
		return errs.E(errs.Validation, "Description is required")
	}
	return nil
}

// Role is a job function or title which defines an authority level.
type Role struct {
	// The unique ID for the Role.
	ID uuid.UUID `json:"-"`
	// Unique External ID to be given to outside callers.
	ExternalID secure.Identifier `json:"external_id"`
	// A human-readable code which represents the role.
	Code string `json:"role_cd"`
	// A longer description of the role.
	Description string `json:"role_description"`
	// A boolean denoting whether the role is active (true) or not (false).
	Active bool `json:"active"`
	// Permissions is the list of permissions allowed for the role.
	Permissions []Permission
	// Users is the list of users for assigned to the role
	Users []user.User
}

// IsValid determines if the Role is valid.
func (r Role) IsValid() error {
	switch {
	case r.ID == uuid.Nil:
		return errs.E(errs.Validation, "ID is required")
	case r.ExternalID.String() == "":
		return errs.E(errs.Validation, "External ID is required")
	case r.Code == "":
		return errs.E(errs.Validation, "Code is required")
	case r.Description == "":
		return errs.E(errs.Validation, "Description is required")
	}
	return nil
}
