package diygoapi

import (
	"context"
	"github.com/rs/zerolog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/secure"
)

// PermissionServicer allows for creating, updating, reading and deleting a Permission
type PermissionServicer interface {
	Create(ctx context.Context, r *CreatePermissionRequest, adt Audit) (*PermissionResponse, error)
	FindAll(ctx context.Context) ([]*PermissionResponse, error)
	Delete(ctx context.Context, extlID string) (DeleteResponse, error)
}

// RoleServicer allows for creating, updating, reading and deleting a Role
// as well as assigning permissions and users to it.
type RoleServicer interface {
	Create(ctx context.Context, r *CreateRoleRequest, adt Audit) (*RoleResponse, error)
}

// AuthenticationServicer represents a service for managing authentication.
//
// For this project, Oauth2 is used for user authentication. It is assumed
// that the actual user interaction is being orchestrated externally and
// the server endpoints are being called after an access token has already
// been retrieved from an authentication provider.
//
// In addition, this project provides for a custom application authentication.
// If an endpoint request is sent using application credentials, then those
// will be used. If none are sent, then the client id from the access token
// must be registered in the system and that is used as the calling application.
// The latter is likely the more common use case.
type AuthenticationServicer interface {

	// SelfRegister is used for first-time registration of a Person/User
	// in the system (associated with an Organization). This is "self
	// registration" as opposed to one person registering another person.
	SelfRegister(ctx context.Context, params AuthenticationParams) (auth Auth, err error)

	// FindAuth looks up a User given a Provider and Access Token.
	// If a User is not found, an error is returned.
	FindAuth(ctx context.Context, params AuthenticationParams) (Auth, error)

	// FindAppByProviderClientID Finds an App given a Provider Client ID as part
	// of an Auth object.
	FindAppByProviderClientID(ctx context.Context, realm string, auth Auth) (a *App, err error)

	// FindAppByAPIKey finds an app given its External ID and determines
	// if the given API key is a valid key for it. It is used as part of
	// app authentication.
	FindAppByAPIKey(ctx context.Context, realm, appExtlID, key string) (*App, error)
}

// AuthorizationServicer represents a service for managing authorization.
type AuthorizationServicer interface {
	Authorize(r *http.Request, lgr zerolog.Logger, adt Audit) error
}

// TokenExchanger exchanges an oauth2.Token for a ProviderUserInfo
// struct populated with information retrieved from an authentication provider.
type TokenExchanger interface {
	Exchange(ctx context.Context, realm string, provider Provider, token *oauth2.Token) (*ProviderInfo, error)
}

// BearerTokenType is used in authorization to access a resource
const BearerTokenType string = "Bearer"

// Provider defines the provider of authorization (Google, Github, Apple, auth0, etc.).
//
// Only Google is used currently.
type Provider uint8

// Provider of authorization
//
// The app uses Oauth2 to authorize users with one of the following Providers
const (
	UnknownProvider Provider = iota
	Google                   // Google
)

func (p Provider) String() string {
	switch p {
	case Google:
		return "google"
	}
	return "unknown_provider"
}

// ParseProvider initializes a Provider given a case-insensitive string
func ParseProvider(s string) Provider {
	switch strings.ToLower(s) {
	case "google":
		return Google
	}
	return UnknownProvider
}

// ProviderInfo contains information returned from an authorization provider
type ProviderInfo struct {
	Provider  Provider
	TokenInfo *ProviderTokenInfo
	UserInfo  *ProviderUserInfo
}

// ProviderTokenInfo contains information gleaned from the Oauth2
// provider's access token
type ProviderTokenInfo struct {
	// Expiration: time of expiration (estimated). This is a moving target as
	// some providers send the actual time of expiration, others
	// just send seconds until expiration, which means it's a
	// calculation and won't have perfect precision.
	Expiration time.Time

	// Client ID: External ID representing the Oauth2 client which
	// authenticated the user.
	ClientID string

	// Scope: The space separated list of scopes granted to this token.
	Scope string
}

// ProviderUserInfo contains common fields from the various Oauth2 providers.
// Currently only using Google, so looks a lot like Google's.
type ProviderUserInfo struct {
	// ID: The obfuscated ID of the user assigned by the authentication provider.
	ExternalID string

	// Email: The user's email address.
	Email string

	// NamePrefix: The name prefix for the Profile (e.g. Mx., Ms., Mr., etc.)
	NamePrefix string

	// MiddleName: The person's middle name.
	MiddleName string

	// FirstName: The user's first name.
	FirstName string

	// FamilyName: The user's last name.
	LastName string

	// FullName: The user's full name.
	FullName string

	// NameSuffix: The name suffix for the person's name (e.g. "PhD", "CCNA", "OBE").
	// Other examples include generational designations like "Sr." and "Jr." and "I", "II", "III", etc.
	NameSuffix string

	// Nickname: The person's nickname
	Nickname string

	// Gender: The user's gender. TODO - setup Gender properly. not binary.
	Gender string

	// BirthDate: The full birthdate of a person (e.g. Dec 18, 1953)
	BirthDate time.Time

	// Hd: The hosted domain e.g. example.com if the user is Google apps
	// user.
	HostedDomain string

	// Link: URL of the profile page.
	ProfileLink string

	// Locale: The user's preferred locale.
	Locale string

	// Picture: URL of the user's picture image.
	Picture string
}

// Auth represents user's OAuth2 credentials.
// Users are linked to a Person. A single Person could authenticate through multiple providers.
type Auth struct {
	// ID is the unique identifier for authorization record in database
	ID uuid.UUID

	// User is the unique user associated to the authorization record.
	//
	// A Person can have one or more methods of authentication, however,
	// only one per authorization provider is allowed per User.
	User *User

	// Provider is the authentication provider
	Provider Provider

	// ProviderClientID is the external ID representing the Oauth2 client which
	// authenticated the user.
	ProviderClientID string

	// ProviderPersonID is the authentication provider's unique person/user ID.
	ProviderPersonID string

	// Token is the Oauth2 token used to determine user identity
	Token *oauth2.Token
}

// Permission stores an approval of a mode of access to a resource.
type Permission struct {
	// ID is the unique ID for the Permission.
	ID uuid.UUID
	// ExternalID is the unique External ID to be given to outside callers.
	ExternalID secure.Identifier
	// Resource is a human-readable string which represents a resource (e.g. an HTTP route or document, etc.).
	Resource string
	// Operation represents the action taken on the resource (e.g. POST, GET, edit, etc.)
	Operation string
	// Description is what the permission is granting, e.g. "grants ability to edit a billing document".
	Description string
	// Active is a boolean denoting whether the permission is active (true) or not (false).
	Active bool
}

// Validate determines if the Permission is valid
func (p Permission) Validate() error {
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

// CreatePermissionRequest is the request struct for creating a permission
type CreatePermissionRequest struct {
	// A human-readable string which represents a resource (e.g. an HTTP route or document, etc.).
	Resource string `json:"resource"`
	// A string representing the action taken on the resource (e.g. POST, GET, edit, etc.)
	Operation string `json:"operation"`
	// A description of what the permission is granting, e.g. "grants ability to edit a billing document".
	Description string `json:"description"`
	// A boolean denoting whether the permission is active (true) or not (false).
	Active bool `json:"active"`
}

// FindPermissionRequest is the response struct for finding a permission
type FindPermissionRequest struct {
	// Unique External ID to be given to outside callers.
	ExternalID string `json:"external_id"`
	// A human-readable string which represents a resource (e.g. an HTTP route or document, etc.).
	Resource string `json:"resource"`
	// A string representing the action taken on the resource (e.g. POST, GET, edit, etc.)
	Operation string `json:"operation"`
}

// PermissionResponse is the response struct for a permission
type PermissionResponse struct {
	// Unique External ID to be given to outside callers.
	ExternalID string `json:"external_id"`
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
	ID uuid.UUID
	// Unique External ID to be given to outside callers.
	ExternalID secure.Identifier
	// A human-readable code which represents the role.
	Code string
	// A longer description of the role.
	Description string
	// A boolean denoting whether the role is active (true) or not (false).
	Active bool
	// Permissions is the list of permissions allowed for the role.
	Permissions []*Permission
}

// Validate determines if the Role is valid.
func (r Role) Validate() error {
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

// CreateRoleRequest is the request struct for creating a role
type CreateRoleRequest struct {
	// A human-readable code which represents the role.
	Code string `json:"role_cd"`
	// A longer description of the role.
	Description string `json:"role_description"`
	// A boolean denoting whether the role is active (true) or not (false).
	Active bool `json:"active"`
	// The list of permissions to be given to the role
	Permissions []*FindPermissionRequest
}

// RoleResponse is the response struct for a Role.
type RoleResponse struct {
	// Unique External ID to be given to outside callers.
	ExternalID string `json:"external_id"`
	// A human-readable code which represents the role.
	Code string `json:"role_cd"`
	// A longer description of the role.
	Description string `json:"role_description"`
	// A boolean denoting whether the role is active (true) or not (false).
	Active bool `json:"active"`
	// Permissions is the list of permissions allowed for the role.
	Permissions []*Permission
}

// AuthenticationParams is the parameters needed for authenticating a User.
type AuthenticationParams struct {
	// Realm is a description of a protected area, used in the WWW-Authenticate header.
	Realm string
	// Provider is the authentication provider.
	Provider Provider
	// Token is the authentication token sent as part of Oauth2.
	Token *oauth2.Token
}
