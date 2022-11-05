package diy

import (
	"context"

	"github.com/google/uuid"

	"github.com/gilcrest/diy-go-api/errs"
	"github.com/gilcrest/diy-go-api/secure"
)

// App is an application that interacts with the system
type App struct {
	ID               uuid.UUID
	ExternalID       secure.Identifier
	Org              *Org
	Name             string
	Description      string
	Provider         Provider
	ProviderClientID string
	APIKeys          []APIKey
}

// AddKey validates and adds an API key to the slice of App API keys
func (a *App) AddKey(key APIKey) error {
	err := key.validate()
	if err != nil {
		return errs.E(errs.Internal, err)
	}
	a.APIKeys = append(a.APIKeys, key)

	return nil
}

// ValidateKey determines if the app has a matching key for the input
// and if that key is valid
func (a *App) ValidateKey(realm, matchKey string) error {
	key, err := a.matchKey(realm, matchKey)
	if err != nil {
		return err
	}
	err = key.validate()
	if err != nil {
		return errs.E(errs.Unauthenticated, errs.Realm(realm), err)
	}
	return nil
}

// MatchKey returns the matching Key given the string, if exists.
// An error will be sent if no match is found.
func (a *App) matchKey(realm, matchKey string) (APIKey, error) {
	for _, apiKey := range a.APIKeys {
		if matchKey == apiKey.Key() {
			return apiKey, nil
		}
	}
	return APIKey{}, errs.E(errs.Unauthenticated, errs.Realm(realm), "Key does not match any keys for the App")
}

// CreateAppRequest is the request struct for Creating an App
type CreateAppRequest struct {
	Name                   string `json:"name"`
	Description            string `json:"description"`
	Oauth2Provider         string `json:"oauth2_provider"`
	Oauth2ProviderClientID string `json:"oauth2_provider_client_id"`
}

// Validate determines whether the CreateAppRequest has proper data to be considered valid
func (r CreateAppRequest) Validate() error {
	switch {
	case r.Name == "":
		return errs.E(errs.Validation, "app name is required")
	case r.Description == "":
		return errs.E(errs.Validation, "app description is required")
	case r.Oauth2Provider == "":
		return errs.E(errs.Validation, "oAuth2 provider is required")
	case r.Oauth2ProviderClientID == "":
		return errs.E(errs.Validation, "oAuth2 client ID is required")
	}
	return nil
}

// UpdateAppRequest is the request struct for Updating an App
type UpdateAppRequest struct {
	ExternalID  string
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AppResponse is the response struct for an App
type AppResponse struct {
	ExternalID          string           `json:"external_id"`
	Name                string           `json:"name"`
	Description         string           `json:"description"`
	CreateAppExtlID     string           `json:"create_app_extl_id"`
	CreateUserFirstName string           `json:"create_user_first_name"`
	CreateUserLastName  string           `json:"create_user_last_name"`
	CreateDateTime      string           `json:"create_date_time"`
	UpdateAppExtlID     string           `json:"update_app_extl_id"`
	UpdateUserFirstName string           `json:"update_user_first_name"`
	UpdateUserLastName  string           `json:"update_user_last_name"`
	UpdateDateTime      string           `json:"update_date_time"`
	APIKeys             []APIKeyResponse `json:"api_keys"`
}

// APIKeyResponse is the response fields for an API key
type APIKeyResponse struct {
	Key              string `json:"key"`
	DeactivationDate string `json:"deactivation_date"`
}

// AppServicer manages the retrieval and manipulation of an App
type AppServicer interface {
	Create(ctx context.Context, r *CreateAppRequest, adt Audit) (*AppResponse, error)
	Update(ctx context.Context, r *UpdateAppRequest, adt Audit) (*AppResponse, error)
}
