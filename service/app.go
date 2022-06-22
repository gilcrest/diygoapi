package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"

	"github.com/gilcrest/diy-go-api/datastore/appstore"
	"github.com/gilcrest/diy-go-api/domain/app"
	"github.com/gilcrest/diy-go-api/domain/audit"
	"github.com/gilcrest/diy-go-api/domain/errs"
	"github.com/gilcrest/diy-go-api/domain/org"
	"github.com/gilcrest/diy-go-api/domain/person"
	"github.com/gilcrest/diy-go-api/domain/secure"
	"github.com/gilcrest/diy-go-api/domain/user"
)

// appAudit is the combination of a domain App and its audit data
type appAudit struct {
	App         app.App
	SimpleAudit audit.SimpleAudit
}

// CreateAppRequest is the request struct for Creating an App
type CreateAppRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func (r CreateAppRequest) isValid() error {
	switch {
	case r.Name == "":
		return errs.E(errs.Validation, "app name is required")
	case r.Description == "":
		return errs.E(errs.Validation, "app description is required")
	}
	return nil
}

// AppResponse is the response struct for an App
type AppResponse struct {
	ExternalID          string           `json:"external_id"`
	Name                string           `json:"name"`
	Description         string           `json:"description"`
	CreateAppExtlID     string           `json:"create_app_extl_id"`
	CreateUsername      string           `json:"create_username"`
	CreateUserFirstName string           `json:"create_user_first_name"`
	CreateUserLastName  string           `json:"create_user_last_name"`
	CreateDateTime      string           `json:"create_date_time"`
	UpdateAppExtlID     string           `json:"update_app_extl_id"`
	UpdateUsername      string           `json:"update_username"`
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

// newAPIKeyResponse initializes an APIKeyResponse. The app.APIKey is
// decrypted and set to the Key field as part of initialization.
func newAPIKeyResponse(key app.APIKey) APIKeyResponse {
	return APIKeyResponse{Key: key.Key(), DeactivationDate: key.DeactivationDate().String()}
}

// newAppResponse initializes an AppResponse given an app.App
func newAppResponse(aa appAudit) AppResponse {
	var keys []APIKeyResponse
	for _, key := range aa.App.APIKeys {
		akr := newAPIKeyResponse(key)
		keys = append(keys, akr)
	}
	return AppResponse{
		ExternalID:          aa.App.ExternalID.String(),
		Name:                aa.App.Name,
		Description:         aa.App.Description,
		CreateAppExtlID:     aa.SimpleAudit.First.App.ExternalID.String(),
		CreateUsername:      aa.SimpleAudit.First.User.Username,
		CreateUserFirstName: aa.SimpleAudit.First.User.Profile.FirstName,
		CreateUserLastName:  aa.SimpleAudit.First.User.Profile.LastName,
		CreateDateTime:      aa.SimpleAudit.First.Moment.Format(time.RFC3339),
		UpdateAppExtlID:     aa.SimpleAudit.Last.App.ExternalID.String(),
		UpdateUsername:      aa.SimpleAudit.Last.User.Username,
		UpdateUserFirstName: aa.SimpleAudit.Last.User.Profile.FirstName,
		UpdateUserLastName:  aa.SimpleAudit.Last.User.Profile.LastName,
		UpdateDateTime:      aa.SimpleAudit.Last.Moment.Format(time.RFC3339),
		APIKeys:             keys,
	}
}

// AppService is a service for creating an App
type AppService struct {
	Datastorer            Datastorer
	RandomStringGenerator CryptoRandomGenerator
	EncryptionKey         *[32]byte
}

// Create is used to create an App
func (s AppService) Create(ctx context.Context, r *CreateAppRequest, adt audit.Audit) (ar AppResponse, err error) {

	var (
		a  app.App
		aa appAudit
	)
	nap := newAppParams{
		r: r,
		// when creating an app, the org the app belongs to must be
		// the same as the org which the user is transacting.
		org:                   adt.App.Org,
		adt:                   adt,
		randomStringGenerator: s.RandomStringGenerator,
		encryptionKey:         s.EncryptionKey,
	}
	a, err = newApp(nap)
	if err != nil {
		return AppResponse{}, err
	}
	aa = appAudit{
		App: a,
		SimpleAudit: audit.SimpleAudit{
			First: adt,
			Last:  adt,
		},
	}

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return AppResponse{}, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	err = createAppTx(ctx, tx, aa)
	if err != nil {
		return AppResponse{}, err
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return AppResponse{}, err
	}

	return newAppResponse(appAudit{App: a, SimpleAudit: audit.SimpleAudit{First: adt, Last: adt}}), nil
}

type newAppParams struct {
	// the request details for the app
	r *CreateAppRequest
	// the org the app belongs to
	org org.Org
	// the audit details of who is creating the app
	adt                   audit.Audit
	randomStringGenerator CryptoRandomGenerator
	encryptionKey         *[32]byte
}

func newApp(nap newAppParams) (a app.App, err error) {
	a = app.App{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Org:         nap.org,
		Name:        nap.r.Name,
		Description: nap.r.Description,
	}

	keyDeactivation := time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC)
	err = a.AddNewKey(nap.randomStringGenerator, nap.encryptionKey, keyDeactivation)
	if err != nil {
		return app.App{}, err
	}

	return a, nil
}

// createAppTx creates the app in the database using a pgx.Tx. This is moved out of the
// app create handler function as it's also used when creating an org.
func createAppTx(ctx context.Context, tx pgx.Tx, aa appAudit) (err error) {
	createAppParams := appstore.CreateAppParams{
		AppID:           aa.App.ID,
		OrgID:           aa.App.Org.ID,
		AppExtlID:       aa.App.ExternalID.String(),
		AppName:         aa.App.Name,
		AppDescription:  aa.App.Description,
		CreateAppID:     aa.SimpleAudit.First.App.ID,
		CreateUserID:    aa.SimpleAudit.First.User.NullUUID(),
		CreateTimestamp: aa.SimpleAudit.First.Moment,
		UpdateAppID:     aa.SimpleAudit.Last.App.ID,
		UpdateUserID:    aa.SimpleAudit.Last.User.NullUUID(),
		UpdateTimestamp: aa.SimpleAudit.Last.Moment,
	}

	// create app database record using appstore
	var rowsAffected int64
	rowsAffected, err = appstore.New(tx).CreateApp(ctx, createAppParams)
	if err != nil {
		return errs.E(errs.Database, err)
	}

	if rowsAffected != 1 {
		return errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	for _, key := range aa.App.APIKeys {

		createAppAPIKeyParams := appstore.CreateAppAPIKeyParams{
			ApiKey:          key.Ciphertext(),
			AppID:           aa.App.ID,
			DeactvDate:      key.DeactivationDate(),
			CreateAppID:     aa.SimpleAudit.First.App.ID,
			CreateUserID:    aa.SimpleAudit.First.User.NullUUID(),
			CreateTimestamp: aa.SimpleAudit.First.Moment,
			UpdateAppID:     aa.SimpleAudit.Last.App.ID,
			UpdateUserID:    aa.SimpleAudit.Last.User.NullUUID(),
			UpdateTimestamp: aa.SimpleAudit.Last.Moment,
		}

		// create app API key database record using appstore
		var apiKeyRowsAffected int64
		apiKeyRowsAffected, err = appstore.New(tx).CreateAppAPIKey(ctx, createAppAPIKeyParams)
		if err != nil {
			return errs.E(errs.Database, err)
		}

		if apiKeyRowsAffected != 1 {
			return errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", apiKeyRowsAffected))
		}
	}

	return nil
}

// UpdateAppRequest is the request struct for Updating an App
type UpdateAppRequest struct {
	ExternalID  string
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Update is used to update an App. API Keys for an App cannot be updated.
func (s AppService) Update(ctx context.Context, r *UpdateAppRequest, adt audit.Audit) (ar AppResponse, err error) {

	// retrieve existing Org
	var aa appAudit
	aa, err = findAppByExternalIDWithAudit(ctx, s.Datastorer.Pool(), r.ExternalID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return AppResponse{}, errs.E(errs.Validation, "No app exists for the given external ID")
		}
		return AppResponse{}, errs.E(errs.Database, err)
	}
	// overwrite Last audit with the current audit
	aa.SimpleAudit.Last = adt

	// override fields with data from request
	aa.App.Name = r.Name
	aa.App.Description = r.Description

	updateAppParams := appstore.UpdateAppParams{
		AppName:         aa.App.Name,
		AppDescription:  aa.App.Description,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    adt.User.NullUUID(),
		UpdateTimestamp: adt.Moment,
		AppID:           aa.App.ID,
	}

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return AppResponse{}, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	var rowsAffected int64
	rowsAffected, err = appstore.New(tx).UpdateApp(ctx, updateAppParams)
	if err != nil {
		return AppResponse{}, errs.E(errs.Database, err)
	}

	if rowsAffected != 1 {
		return AppResponse{}, errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return AppResponse{}, err
	}

	return newAppResponse(aa), nil
}

// Delete is used to delete an App
func (s AppService) Delete(ctx context.Context, extlID string) (dr DeleteResponse, err error) {

	// retrieve existing App
	var a app.App
	a, err = findAppByExternalID(ctx, s.Datastorer.Pool(), extlID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return DeleteResponse{}, errs.E(errs.Validation, "No app exists for the given external ID")
		}
		return DeleteResponse{}, errs.E(errs.Database, err)
	}

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return DeleteResponse{}, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	err = deleteAppTx(ctx, tx, a)
	if err != nil {
		return DeleteResponse{}, err
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return DeleteResponse{}, err
	}

	response := DeleteResponse{
		ExternalID: extlID,
		Deleted:    true,
	}

	return response, nil
}

func deleteAppTx(ctx context.Context, tx pgx.Tx, a app.App) (err error) {
	// one-to-many API keys can be associated with an App. This will
	// delete them all.
	var apiKeysRowsAffected int64
	apiKeysRowsAffected, err = appstore.New(tx).DeleteAppAPIKeys(ctx, a.ID)
	if err != nil {
		return errs.E(errs.Database, err)
	}

	if apiKeysRowsAffected < 1 {
		return errs.E(errs.Database, fmt.Sprintf("rows affected should be at least 1, actual: %d", apiKeysRowsAffected))
	}

	var rowsAffected int64
	rowsAffected, err = appstore.New(tx).DeleteApp(ctx, a.ID)
	if err != nil {
		return errs.E(errs.Database, err)
	}

	if rowsAffected != 1 {
		return errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	return nil
}

// FindByExternalID is used to find an App by its External ID
func (s AppService) FindByExternalID(ctx context.Context, extlID string) (ar AppResponse, err error) {

	var aa appAudit
	aa, err = findAppByExternalIDWithAudit(ctx, s.Datastorer.Pool(), extlID)
	if err != nil {
		return AppResponse{}, err
	}

	return newAppResponse(aa), nil
}

// FindAll is used to list all apps in the datastore
func (s AppService) FindAll(ctx context.Context) (sar []AppResponse, err error) {

	var (
		rows      []appstore.FindAppsWithAuditRow
		responses []AppResponse
	)
	rows, err = appstore.New(s.Datastorer.Pool()).FindAppsWithAudit(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, err)
	}

	for _, row := range rows {
		a := app.App{
			ID:         row.AppID,
			ExternalID: secure.MustParseIdentifier(row.AppExtlID),
			Org: org.Org{
				ID:          row.OrgID,
				ExternalID:  secure.MustParseIdentifier(row.OrgExtlID),
				Name:        row.OrgName,
				Description: row.OrgDescription,
				Kind: org.Kind{
					ID:          row.OrgKindID,
					ExternalID:  row.OrgKindExtlID,
					Description: row.OrgKindDesc,
				},
			},
			Name:        row.AppName,
			Description: row.AppDescription,
			APIKeys:     nil,
		}

		sa := audit.SimpleAudit{
			First: audit.Audit{
				App: app.App{
					ID:          row.CreateAppID,
					ExternalID:  secure.MustParseIdentifier(row.CreateAppExtlID),
					Org:         org.Org{ID: row.CreateAppOrgID},
					Name:        row.CreateAppName,
					Description: row.CreateAppDescription,
					APIKeys:     nil,
				},
				User: user.User{
					ID:       row.CreateUserID.UUID,
					Username: row.CreateUsername,
					Org:      org.Org{ID: row.CreateUserOrgID},
					Profile: person.Profile{
						FirstName: row.CreateUserFirstName,
						LastName:  row.CreateUserLastName,
					},
				},
				Moment: row.CreateTimestamp,
			},
			Last: audit.Audit{
				App: app.App{
					ID:          row.UpdateAppID,
					ExternalID:  secure.MustParseIdentifier(row.UpdateAppExtlID),
					Org:         org.Org{ID: row.UpdateAppOrgID},
					Name:        row.UpdateAppName,
					Description: row.UpdateAppDescription,
					APIKeys:     nil,
				},
				User: user.User{
					ID:       row.UpdateUserID.UUID,
					Username: row.UpdateUsername,
					Org:      org.Org{ID: row.UpdateUserOrgID},
					Profile: person.Profile{
						FirstName: row.UpdateUserFirstName,
						LastName:  row.UpdateUserLastName,
					},
				},
				Moment: row.UpdateTimestamp,
			},
		}
		or := newAppResponse(appAudit{App: a, SimpleAudit: sa})

		responses = append(responses, or)
	}

	return responses, nil
}

func findAppByExternalID(ctx context.Context, dbtx DBTX, extlID string) (app.App, error) {
	row, err := appstore.New(dbtx).FindAppByExternalID(ctx, extlID)
	if err != nil {
		return app.App{}, errs.E(errs.Database, err)
	}

	a := app.App{
		ID:         row.AppID,
		ExternalID: secure.MustParseIdentifier(row.AppExtlID),
		Org: org.Org{
			ID:          row.OrgID,
			ExternalID:  secure.MustParseIdentifier(row.OrgExtlID),
			Name:        row.OrgName,
			Description: row.OrgDescription,
			Kind: org.Kind{
				ID:          row.OrgKindID,
				ExternalID:  row.OrgKindExtlID,
				Description: row.OrgKindDesc,
			},
		},
		Name:        row.AppName,
		Description: row.AppDescription,
		APIKeys:     nil,
	}

	return a, nil
}

// findAppByExternalIDWithAudit retrieves App data from the datastore given a unique external ID.
// This data is then hydrated into the app.App struct along with the simple audit struct
func findAppByExternalIDWithAudit(ctx context.Context, dbtx DBTX, extlID string) (appAudit, error) {
	var (
		row appstore.FindAppByExternalIDWithAuditRow
		err error
	)

	row, err = appstore.New(dbtx).FindAppByExternalIDWithAudit(ctx, extlID)
	if err != nil {
		return appAudit{}, errs.E(errs.Database, err)
	}

	a := app.App{
		ID:         row.AppID,
		ExternalID: secure.MustParseIdentifier(row.AppExtlID),
		Org: org.Org{
			ID:          row.OrgID,
			ExternalID:  secure.MustParseIdentifier(row.OrgExtlID),
			Name:        row.OrgName,
			Description: row.OrgDescription,
			Kind: org.Kind{
				ID:          row.OrgKindID,
				ExternalID:  row.OrgKindExtlID,
				Description: row.OrgKindDesc,
			},
		},
		Name:        row.AppName,
		Description: row.AppDescription,
		APIKeys:     nil,
	}

	sa := audit.SimpleAudit{
		First: audit.Audit{
			App: app.App{
				ID:          row.CreateAppID,
				ExternalID:  secure.MustParseIdentifier(row.CreateAppExtlID),
				Org:         org.Org{ID: row.CreateAppOrgID},
				Name:        row.CreateAppName,
				Description: row.CreateAppDescription,
				APIKeys:     nil,
			},
			User: user.User{
				ID:       row.CreateUserID.UUID,
				Username: row.CreateUsername,
				Org:      org.Org{ID: row.CreateUserOrgID},
				Profile: person.Profile{
					FirstName: row.CreateUserFirstName,
					LastName:  row.CreateUserLastName,
				},
			},
			Moment: row.CreateTimestamp,
		},
		Last: audit.Audit{
			App: app.App{
				ID:          row.UpdateAppID,
				ExternalID:  secure.MustParseIdentifier(row.UpdateAppExtlID),
				Org:         org.Org{ID: row.UpdateAppOrgID},
				Name:        row.UpdateAppName,
				Description: row.UpdateAppDescription,
				APIKeys:     nil,
			},
			User: user.User{
				ID:       row.UpdateUserID.UUID,
				Username: row.UpdateUsername,
				Org:      org.Org{ID: row.UpdateUserOrgID},
				Profile: person.Profile{
					FirstName: row.UpdateUserFirstName,
					LastName:  row.UpdateUserLastName,
				},
			},
			Moment: row.UpdateTimestamp,
		},
	}

	return appAudit{App: a, SimpleAudit: sa}, nil
}
