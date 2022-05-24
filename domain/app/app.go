package app

import (
	"context"
	"net/http"
	"time"

	"github.com/gilcrest/diy-go-api/domain/errs"
	"github.com/gilcrest/diy-go-api/domain/org"
	"github.com/gilcrest/diy-go-api/domain/secure"
	"github.com/google/uuid"
)

// App is an application that interacts with the system
type App struct {
	ID          uuid.UUID
	ExternalID  secure.Identifier
	Org         org.Org
	Name        string
	Description string
	APIKeys     []APIKey
}

// AddKey adds the API key to slice of API keys for the App
func (a *App) AddKey(key APIKey) error {
	err := key.isValid()
	if err != nil {
		return errs.E(errs.Internal, err)
	}
	a.APIKeys = append(a.APIKeys, key)

	return nil
}

// AddNewKey adds a newly generated API key to the slice of API keys for the App
func (a *App) AddNewKey(g APIKeyStringGenerator, ek *[32]byte, deactivation time.Time) error {
	var (
		key APIKey
		err error
	)

	// generate App API key
	key, err = NewAPIKey(g, ek)
	if err != nil {
		return err
	}
	key.SetDeactivationDate(deactivation)

	err = key.isValid()
	if err != nil {
		return errs.E(errs.Internal, err)
	}
	a.APIKeys = append(a.APIKeys, key)

	return nil
}

// ValidKey determines if the app has a matching key for the input
// and if that key is valid
func (a App) ValidKey(realm, matchKey string) error {
	key, err := a.matchKey(realm, matchKey)
	if err != nil {
		return err
	}
	err = key.isValid()
	if err != nil {
		return errs.E(errs.Unauthenticated, errs.Realm(realm), err)
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
