package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/datastore/appstore"
	"github.com/gilcrest/go-api-basic/domain/app"
	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/org"
	"github.com/gilcrest/go-api-basic/domain/secure"
	"github.com/gilcrest/go-api-basic/domain/user"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
)

// CreateAppRequest is the request struct for Creating an App
type CreateAppRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AppResponse is the response struct for an App
type AppResponse struct {
	ExternalID  string           `json:"external_id"`
	Name        string           `json:"name"`
	Description string           `json:"description"`
	APIKeys     []APIKeyResponse `json:"api_keys"`
}

// APIKeyResponse is the response fields for an API key
type APIKeyResponse struct {
	Key              string `json:"key"`
	DeactivationDate string `json:"deactivation_date"`
}

// newAPIKeyResponse initializes an APIKeyResponse. The app.APIKey is
// decrypted and set to the Key field as part of initialization.
func newAPIKeyResponse(key app.APIKey) APIKeyResponse {
	return APIKeyResponse{Key: key.Key(), DeactivationDate: key.DeactivationDate().String()}
}

// newAppResponse initializes an AppResponse given an app.App
func newAppResponse(a app.App) AppResponse {
	var keys []APIKeyResponse
	for _, key := range a.APIKeys {
		akr := newAPIKeyResponse(key)
		keys = append(keys, akr)
	}
	return AppResponse{
		ExternalID:  a.ExternalID.String(),
		Name:        a.Name,
		Description: a.Description,
		APIKeys:     keys,
	}
}

// CreateAppService is a service for creating an App
type CreateAppService struct {
	Datastorer            Datastorer
	RandomStringGenerator CryptoRandomGenerator
	EncryptionKey         *[32]byte
}

// Create is used to create an App
func (s CreateAppService) Create(ctx context.Context, r *CreateAppRequest, adt audit.Audit) (AppResponse, error) {
	var (
		a   app.App
		err error
	)
	a.ID = uuid.New()
	a.ExternalID = secure.NewID()
	a.Org = adt.App.Org
	a.Name = r.Name
	a.Description = r.Description

	keyDeactivation := time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC)
	err = a.AddNewKey(s.RandomStringGenerator, s.EncryptionKey, keyDeactivation)
	if err != nil {
		return AppResponse{}, err
	}

	// start db txn using pgxpool
	tx, err := s.Datastorer.BeginTx(ctx)
	if err != nil {
		return AppResponse{}, err
	}

	err = createAppDB(ctx, s.Datastorer, tx, a, adt)
	if err != nil {
		return AppResponse{}, err
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return AppResponse{}, err
	}

	return newAppResponse(a), nil
}

// createAppDB creates an app in the database given a domain app.App and audit.Audit
func createAppDB(ctx context.Context, ds Datastorer, tx pgx.Tx, a app.App, adt audit.Audit) error {
	var err error

	if len(a.APIKeys) == 0 {
		return errs.E(errs.Internal, ds.RollbackTx(ctx, tx, errs.E("app must have at least one API key.")))
	}

	// create app database record using appstore
	_, err = appstore.New(tx).CreateApp(ctx, newCreateAppParams(a, adt))
	if err != nil {
		return errs.E(errs.Database, ds.RollbackTx(ctx, tx, err))
	}

	for _, aak := range a.APIKeys {
		// create app API key database record using appstore
		_, err = appstore.New(tx).CreateAppAPIKey(ctx, newCreateAppAPIKeyParams(a, aak, adt))
		if err != nil {
			return errs.E(errs.Database, ds.RollbackTx(ctx, tx, err))
		}
	}

	return nil
}

// newCreateAppParams maps an App to appstore.CreateAppParams
func newCreateAppParams(a app.App, adt audit.Audit) appstore.CreateAppParams {
	return appstore.CreateAppParams{
		AppID:           a.ID,
		OrgID:           a.Org.ID,
		AppExtlID:       a.ExternalID.String(),
		AppName:         a.Name,
		AppDescription:  a.Description,
		Active:          sql.NullBool{Bool: true, Valid: true},
		CreateAppID:     adt.App.ID,
		CreateUserID:    datastore.NewNullUUID(adt.User.ID),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    datastore.NewNullUUID(adt.User.ID),
		UpdateTimestamp: adt.Moment,
	}
}

// newCreateAppAPIKeyParams maps an AppAPIKey to appstore.CreateAppAPIKeyParams
func newCreateAppAPIKeyParams(a app.App, k app.APIKey, adt audit.Audit) appstore.CreateAppAPIKeyParams {
	return appstore.CreateAppAPIKeyParams{
		ApiKey:          k.Ciphertext(),
		AppID:           a.ID,
		DeactvDate:      k.DeactivationDate(),
		CreateAppID:     adt.App.ID,
		CreateUserID:    datastore.NewNullUUID(adt.User.ID),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    datastore.NewNullUUID(adt.User.ID),
		UpdateTimestamp: adt.Moment,
	}
}

// FindAppService is a service for retrieving an App from the datastore
type FindAppService struct {
	Datastorer    Datastorer
	EncryptionKey *[32]byte
}

// FindAppByAPIKey finds an app given its External ID and determines
// if the given API key is a valid key for it. It is used as part of
// app authentication
func (fas FindAppService) FindAppByAPIKey(ctx context.Context, realm, appExtlID, apiKey string) (app.App, error) {

	var (
		kr  []appstore.FindAppAPIKeysByAppExtlIDRow
		err error
	)

	kr, err = appstore.New(fas.Datastorer.Pool()).FindAppAPIKeysByAppExtlID(ctx, appExtlID)
	if err != nil {
		return app.App{}, errs.E(errs.Unauthenticated, errs.Realm(realm), err)
	}

	var (
		a    app.App
		keys []app.APIKey
	)
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
		var key app.APIKey
		key, err = app.NewAPIKeyFromCipher(row.ApiKey, fas.EncryptionKey)
		if err != nil {
			return app.App{}, err
		}
		key.SetDeactivationDate(row.DeactvDate)
		keys = append(keys, key)
	}
	a.APIKeys = keys

	err = a.ValidKey(realm, apiKey)
	if err != nil {
		return app.App{}, err
	}

	return a, nil
}

// findAppByID finds an app from the datastore given its ID
func findAppByID(ctx context.Context, dbtx DBTX, id uuid.UUID, withAudit bool) (app.App, *audit.SimpleAudit, error) {
	dba, err := appstore.New(dbtx).FindAppByID(ctx, id)
	if err != nil {
		return app.App{}, nil, errs.E(errs.Database, err)
	}

	a, sa, err := newApp(ctx, dbtx, dba, withAudit)
	if err != nil {
		return app.App{}, nil, err
	}

	return a, sa, nil
}

// newApp hydrates an appstore.App into an app.App and an audit.SimpleAudit (if withAudit is true)
func newApp(ctx context.Context, dbtx DBTX, dba appstore.App, withAudit bool) (app.App, *audit.SimpleAudit, error) {
	var (
		extl                   secure.Identifier
		err                    error
		createApp, updateApp   app.App
		createUser, updateUser user.User
		o                      org.Org
	)

	o, err = findOrgByID(ctx, dbtx, dba.OrgID)
	if err != nil {
		return app.App{}, nil, err
	}

	extl, err = secure.ParseIdentifier(dba.AppExtlID)
	if err != nil {
		return app.App{}, nil, err
	}

	a := app.App{
		ID:          dba.AppID,
		ExternalID:  extl,
		Org:         o,
		Name:        dba.AppName,
		Description: dba.AppDescription,
		APIKeys:     nil,
	}

	sa := new(audit.SimpleAudit)
	if withAudit {
		createApp, _, err = findAppByID(ctx, dbtx, dba.CreateAppID, false)
		if err != nil {
			return app.App{}, nil, err
		}
		createUser, err = findUserByID(ctx, dbtx, dba.CreateUserID.UUID)
		if err != nil {
			return app.App{}, nil, err
		}
		createAudit := audit.Audit{
			App:    createApp,
			User:   createUser,
			Moment: dba.CreateTimestamp,
		}

		updateApp, _, err = findAppByID(ctx, dbtx, dba.UpdateAppID, false)
		if err != nil {
			return app.App{}, nil, err
		}
		updateUser, err = findUserByID(ctx, dbtx, dba.UpdateUserID.UUID)
		if err != nil {
			return app.App{}, nil, err
		}
		updateAudit := audit.Audit{
			App:    updateApp,
			User:   updateUser,
			Moment: dba.UpdateTimestamp,
		}

		sa = &audit.SimpleAudit{First: createAudit, Last: updateAudit}
	}

	return a, sa, nil
}
