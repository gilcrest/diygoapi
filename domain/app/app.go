package app

import (
	"context"
	"net/http"
	"time"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/org"
	"github.com/gilcrest/go-api-basic/domain/secure"
	"github.com/google/uuid"
)

// App is an application that interacts with the system
type App struct {
	ID           uuid.UUID
	ExternalID   secure.Identifier
	Org          org.Org
	Name         string
	Description  string
	CreateAppID  uuid.UUID
	CreateUserID uuid.UUID
	CreateTime   time.Time
	UpdateAppID  uuid.UUID
	UpdateUserID uuid.UUID
	UpdateTime   time.Time
	APIKeys      []APIKey
}

// ValidKey determines if the app has a matching key for the input
// and if that key is valid
func (a App) ValidKey(realm, matchKey string) error {
	key, err := a.matchKey(realm, matchKey)
	if err != nil {
		return err
	}
	err = key.isValid(realm)
	if err != nil {
		return err
	}
	return nil
}

// MatchKey returns the matching Key given the string, if exists
// An error will be sent if no match is found
func (a App) matchKey(realm, matchKey string) (APIKey, error) {
	for _, apiKey := range a.APIKeys {
		if matchKey == apiKey.Key() {
			return apiKey, nil
		}
	}
	return APIKey{}, errs.E(errs.Unauthenticated, errs.Realm(realm), "Key does not match any keys for the App")
}

type contextKey string

const contextKeyUser = contextKey("app")

// FromRequest gets the App from the request
func FromRequest(r *http.Request) (App, error) {
	adt, ok := r.Context().Value(contextKeyUser).(App)
	if !ok {
		return adt, errs.E(errs.Internal, "App not set properly to context")
	}
	return adt, nil
}

// CtxWithApp sets the App to the given context
func CtxWithApp(ctx context.Context, a App) context.Context {
	return context.WithValue(ctx, contextKeyUser, a)
}
