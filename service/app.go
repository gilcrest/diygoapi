package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/datastore/appstore"
	"github.com/gilcrest/go-api-basic/domain/app"
	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/org"
	"github.com/gilcrest/go-api-basic/domain/secure"
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
	CryptoRandomGenerator CryptoRandomGenerator
	EncryptionKey         *[32]byte
}

// Create is used to create an App
func (cas CreateAppService) Create(ctx context.Context, r *CreateAppRequest, adt audit.Audit) (AppResponse, error) {
	var a app.App
	a.ID = uuid.New()
	a.ExternalID = secure.NewID()
	a.Org = adt.App.Org
	a.Name = r.Name
	a.Description = r.Description

	aak, err := app.NewAPIKey(cas.CryptoRandomGenerator, cas.EncryptionKey)
	if err != nil {
		return AppResponse{}, err
	}
	aak.SetDeactivationDate(time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC))

	var keys []app.APIKey
	keys = append(keys, aak)
	a.APIKeys = keys

	// start db txn using pgxpool
	tx, err := cas.Datastorer.BeginTx(ctx)
	if err != nil {
		return AppResponse{}, err
	}

	// create app database record using appstore
	_, err = appstore.New(tx).CreateApp(ctx, NewCreateAppParams(a, adt))
	if err != nil {
		return AppResponse{}, errs.E(errs.Database, cas.Datastorer.RollbackTx(ctx, tx, err))
	}

	// create app API key database record using appstore
	_, err = appstore.New(tx).CreateAppAPIKey(ctx, NewCreateAppAPIKeyParams(a, aak, adt))
	if err != nil {
		return AppResponse{}, errs.E(errs.Database, cas.Datastorer.RollbackTx(ctx, tx, err))
	}

	// commit db txn using pgxpool
	err = cas.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return AppResponse{}, err
	}

	return newAppResponse(a), nil
}

// NewCreateAppParams maps an App to appstore.CreateAppParams
func NewCreateAppParams(a app.App, adt audit.Audit) appstore.CreateAppParams {
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

// NewCreateAppAPIKeyParams maps an AppAPIKey to appstore.CreateAppAPIKeyParams
func NewCreateAppAPIKeyParams(a app.App, k app.APIKey, adt audit.Audit) appstore.CreateAppAPIKeyParams {
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

// DBTX interface mirrors the interface generated by sqlc to allow
// passing a Pool or a Tx
type DBTX interface {
	Exec(context.Context, string, ...interface{}) (pgconn.CommandTag, error)
	Query(context.Context, string, ...interface{}) (pgx.Rows, error)
	QueryRow(context.Context, string, ...interface{}) pgx.Row
}

// findAppByID finds an app from the datastore given its ID
func findAppByID(ctx context.Context, dbtx DBTX, id uuid.UUID) (app.App, error) {
	dba, err := appstore.New(dbtx).FindAppByID(ctx, id)
	if err != nil {
		return app.App{}, errs.E(errs.Database, err)
	}
	o, err := findOrgByID(ctx, dbtx, dba.OrgID)
	if err != nil {
		return app.App{}, err
	}
	var extl secure.Identifier
	extl, err = secure.ParseIdentifier(dba.AppExtlID)
	if err != nil {
		return app.App{}, err
	}

	a := app.App{
		ID:           dba.AppID,
		ExternalID:   extl,
		Org:          o,
		Name:         dba.AppName,
		Description:  dba.AppDescription,
		CreateAppID:  dba.CreateAppID,
		CreateUserID: dba.CreateUserID.UUID,
		CreateTime:   dba.CreateTimestamp,
		UpdateAppID:  dba.UpdateAppID,
		UpdateUserID: dba.UpdateUserID.UUID,
		UpdateTime:   dba.UpdateTimestamp,
		APIKeys:      nil,
	}

	return a, nil
}
