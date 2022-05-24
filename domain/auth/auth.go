// Package auth is for user and application authorization logic
package auth

import (
	"strings"

	"github.com/google/uuid"

	"github.com/gilcrest/diy-go-api/domain/secure"
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
}
