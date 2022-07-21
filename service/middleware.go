package service

import (
	"context"
	"net/http"

	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"

	"github.com/gilcrest/diy-go-api/datastore/appstore"
	"github.com/gilcrest/diy-go-api/datastore/userstore"
	"github.com/gilcrest/diy-go-api/domain/app"
	"github.com/gilcrest/diy-go-api/domain/audit"
	"github.com/gilcrest/diy-go-api/domain/auth"
	"github.com/gilcrest/diy-go-api/domain/errs"
	"github.com/gilcrest/diy-go-api/domain/org"
	"github.com/gilcrest/diy-go-api/domain/secure"
	"github.com/gilcrest/diy-go-api/domain/user"
	"github.com/gilcrest/diy-go-api/gateway/authgateway"
)

// Authorizer determines if an app/user (as part of an audit.Audit struct) is
// authorized for the route in the request
type Authorizer interface {
	Authorize(lgr zerolog.Logger, r *http.Request, sub audit.Audit) error
}

// MiddlewareService holds methods used by server middleware handlers
type MiddlewareService struct {
	Datastorer                 Datastorer
	GoogleOauth2TokenConverter GoogleOauth2TokenConverter
	Authorizer                 Authorizer
	EncryptionKey              *[32]byte
}

// FindAppByAPIKey finds an app given its External ID and determines
// if the given API key is a valid key for it. It is used as part of
// app authentication
func (s MiddlewareService) FindAppByAPIKey(ctx context.Context, realm, appExtlID, key string) (app.App, error) {

	var (
		kr  []appstore.FindAppAPIKeysByAppExtlIDRow
		err error
	)

	// retrieve the list of encrypted API keys from the database
	kr, err = appstore.New(s.Datastorer.Pool()).FindAppAPIKeysByAppExtlID(ctx, appExtlID)
	if err != nil {
		return app.App{}, errs.E(errs.Unauthenticated, errs.Realm(realm), err)
	}

	var (
		a   app.App
		ak  app.APIKey
		aks []app.APIKey
	)

	// for each row, decrypt the API key using the encryption key,
	// initialize an app.APIKey and set to a slice of API keys.
	for i, row := range kr {
		if i == 0 { // only need to fill the app struct on first iteration
			var extl secure.Identifier
			extl, err = secure.ParseIdentifier(row.OrgExtlID)
			if err != nil {
				return app.App{}, err
			}
			a.ID = row.AppID
			a.ExternalID = extl
			a.Org = org.Org{
				ID:          row.OrgID,
				ExternalID:  extl,
				Name:        row.OrgName,
				Description: row.OrgDescription,
			}
			a.Name = row.AppName
			a.Description = row.AppDescription
		}
		ak, err = app.NewAPIKeyFromCipher(row.ApiKey, s.EncryptionKey)
		if err != nil {
			return app.App{}, err
		}
		ak.SetDeactivationDate(row.DeactvDate)
		aks = append(aks, ak)
	}
	a.APIKeys = aks

	// ValidKey determines if any of the keys attached to the app
	// match the input key and are still valid.
	err = a.ValidKey(realm, key)
	if err != nil {
		return app.App{}, err
	}

	return a, nil
}

// GoogleOauth2TokenConverter converts an oauth2.Token to an authgateway.Userinfo struct
type GoogleOauth2TokenConverter interface {
	Convert(ctx context.Context, realm string, token oauth2.Token) (authgateway.ProviderUserInfo, error)
}

// FindUserParams is parameters for finding a User
type FindUserParams struct {
	Realm          string
	App            app.App
	Provider       auth.Provider
	Token          oauth2.Token
	RetrieveFromDB bool
}

// FindUserByOauth2Token retrieves a users' identity from a Provider
// and then retrieves the associated registered user from the datastore
func (s MiddlewareService) FindUserByOauth2Token(ctx context.Context, params FindUserParams) (user.User, error) {
	var (
		uInfo authgateway.ProviderUserInfo
		err   error
	)

	if params.Provider == auth.Invalid {
		return user.User{}, errs.E(errs.Unauthenticated, errs.Realm(params.Realm), "provider not recognized")
	}

	if params.Provider == auth.Apple {
		return user.User{}, errs.E(errs.Unauthenticated, errs.Realm(params.Realm), "apple authentication not yet implemented")
	}

	if params.Provider == auth.Google {
		uInfo, err = s.GoogleOauth2TokenConverter.Convert(ctx, params.Realm, params.Token)
		if err != nil {
			return user.User{}, err
		}
	}

	findUserByUsernameParams := userstore.FindUserByUsernameParams{
		Username: uInfo.Username,
		OrgID:    params.App.Org.ID,
	}

	if params.RetrieveFromDB {
		var findUserByUsernameRow userstore.FindUserByUsernameRow
		findUserByUsernameRow, err = userstore.New(s.Datastorer.Pool()).FindUserByUsername(ctx, findUserByUsernameParams)
		if err != nil {
			if err == pgx.ErrNoRows {
				return user.User{}, errs.E(errs.Unauthenticated, errs.Realm(params.Realm), "No user registered in database")
			}
			return user.User{}, errs.E(errs.Unauthenticated, errs.Realm(params.Realm), err)
		}

		return hydrateUserFromUsernameRow(findUserByUsernameRow), nil
	}

	return hydrateUserFromProviderUserInfo(params, uInfo), nil
}

// Authorize determines if an app/user (as part of an Audit) is
// authorized for the route in the request
func (s MiddlewareService) Authorize(lgr zerolog.Logger, r *http.Request, sub audit.Audit) error {
	return s.Authorizer.Authorize(lgr, r, sub)
}
