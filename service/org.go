package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"

	"github.com/gilcrest/diy-go-api/datastore/appstore"
	"github.com/gilcrest/diy-go-api/datastore/orgstore"
	"github.com/gilcrest/diy-go-api/domain/app"
	"github.com/gilcrest/diy-go-api/domain/audit"
	"github.com/gilcrest/diy-go-api/domain/errs"
	"github.com/gilcrest/diy-go-api/domain/org"
	"github.com/gilcrest/diy-go-api/domain/person"
	"github.com/gilcrest/diy-go-api/domain/secure"
	"github.com/gilcrest/diy-go-api/domain/user"
)

// orgAudit is the combination of a domain Org and its audit data
type orgAudit struct {
	Org         org.Org
	SimpleAudit audit.SimpleAudit
}

// CreateOrgRequest is the request struct for Creating an Org
type CreateOrgRequest struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Kind        string           `json:"kind"`
	App         CreateAppRequest `json:"app"`
}

func (r CreateOrgRequest) isValid() error {
	switch {
	case r.Name == "":
		return errs.E(errs.Validation, "org name is required")
	case r.Description == "":
		return errs.E(errs.Validation, "org description is required")
	case r.Kind == "":
		return errs.E(errs.Validation, "org kind is required")
	}
	return nil
}

// OrgResponse is the response struct for an Org
type OrgResponse struct {
	ExternalID          string      `json:"external_id"`
	Name                string      `json:"name"`
	KindExternalID      string      `json:"kind_description"`
	Description         string      `json:"description"`
	CreateAppExtlID     string      `json:"create_app_extl_id"`
	CreateUsername      string      `json:"create_username"`
	CreateUserFirstName string      `json:"create_user_first_name"`
	CreateUserLastName  string      `json:"create_user_last_name"`
	CreateDateTime      string      `json:"create_date_time"`
	UpdateAppExtlID     string      `json:"update_app_extl_id"`
	UpdateUsername      string      `json:"update_username"`
	UpdateUserFirstName string      `json:"update_user_first_name"`
	UpdateUserLastName  string      `json:"update_user_last_name"`
	UpdateDateTime      string      `json:"update_date_time"`
	App                 AppResponse `json:"app"`
}

// newOrgResponse initializes OrgResponse given an org.Org.
func newOrgResponse(oa orgAudit, aa appAudit) OrgResponse {
	r := OrgResponse{
		ExternalID:          oa.Org.ExternalID.String(),
		Name:                oa.Org.Name,
		Description:         oa.Org.Description,
		KindExternalID:      oa.Org.Kind.ExternalID,
		CreateAppExtlID:     oa.SimpleAudit.First.App.ExternalID.String(),
		CreateUsername:      oa.SimpleAudit.First.User.Username,
		CreateUserFirstName: oa.SimpleAudit.First.User.Profile.FirstName,
		CreateUserLastName:  oa.SimpleAudit.First.User.Profile.LastName,
		CreateDateTime:      oa.SimpleAudit.First.Moment.Format(time.RFC3339),
		UpdateAppExtlID:     oa.SimpleAudit.Last.App.ExternalID.String(),
		UpdateUsername:      oa.SimpleAudit.Last.User.Username,
		UpdateUserFirstName: oa.SimpleAudit.Last.User.Profile.FirstName,
		UpdateUserLastName:  oa.SimpleAudit.Last.User.Profile.LastName,
		UpdateDateTime:      oa.SimpleAudit.Last.Moment.Format(time.RFC3339),
	}

	if aa.App.ID != uuid.Nil {
		r.App = newAppResponse(aa)
	}

	return r
}

// CreateOrgService is a service for creating an Org (and optionally an App with it)
type CreateOrgService struct {
	Datastorer            Datastorer
	RandomStringGenerator CryptoRandomGenerator
	EncryptionKey         *[32]byte
}

// Create is used to create an Org
func (s CreateOrgService) Create(ctx context.Context, r *CreateOrgRequest, adt audit.Audit) (or OrgResponse, err error) {
	err = r.isValid()
	if err != nil {
		return OrgResponse{}, err
	}

	sa := audit.SimpleAudit{
		First: adt,
		Last:  adt,
	}

	var kind org.Kind
	kind, err = findOrgKindByExtlID(ctx, s.Datastorer.Pool(), r.Kind)
	if err != nil {
		return OrgResponse{}, err
	}

	// initialize Org and inject dependent fields
	o := org.Org{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Name:        r.Name,
		Description: r.Description,
		Kind:        kind,
	}
	oa := orgAudit{
		Org:         o,
		SimpleAudit: sa,
	}

	// if there is an app request along with the Org request, process it as well
	var (
		car CreateAppRequest
		a   app.App
		aa  appAudit
	)
	if r.App != car {
		err = r.App.isValid()
		if err != nil {
			return OrgResponse{}, err
		}
		nap := newAppParams{
			r:                     &r.App,
			org:                   o,
			adt:                   adt,
			randomStringGenerator: s.RandomStringGenerator,
			encryptionKey:         s.EncryptionKey,
		}
		a, err = newApp(nap)
		if err != nil {
			return OrgResponse{}, err
		}
		aa = appAudit{
			App:         a,
			SimpleAudit: sa,
		}
	}

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return OrgResponse{}, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	// write org to the db
	err = createOrgTx(ctx, tx, oa)
	if err != nil {
		return OrgResponse{}, err
	}

	// if app is also to be created, write it to the db
	if aa.App.ID != uuid.Nil {
		err = createAppTx(ctx, tx, aa)
		if err != nil {
			return OrgResponse{}, err
		}
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return OrgResponse{}, err
	}

	response := newOrgResponse(oa, aa)

	return response, nil
}

// createOrgTx writes an Org and its audit information to the database.
// separate function as it's used by genesis service as well
func createOrgTx(ctx context.Context, tx pgx.Tx, oa orgAudit) error {
	if oa.Org.Kind.ID == uuid.Nil {
		return errs.E("org Kind is required")
	}

	// create database record using orgstore
	rowsAffected, err := orgstore.New(tx).CreateOrg(ctx, newCreateOrgParams(oa))
	if err != nil {
		return errs.E(errs.Database, err)
	}

	// update should only update exactly one record
	if rowsAffected != 1 {
		return errs.E(errs.Database, fmt.Sprintf("CreateOrg() should insert 1 row, actual: %d", rowsAffected))
	}

	return nil
}

// newCreateOrgParams maps an Org to orgstore.CreateOrgParams
func newCreateOrgParams(oa orgAudit) orgstore.CreateOrgParams {
	return orgstore.CreateOrgParams{
		OrgID:           oa.Org.ID,
		OrgExtlID:       oa.Org.ExternalID.String(),
		OrgName:         oa.Org.Name,
		OrgDescription:  oa.Org.Description,
		OrgKindID:       oa.Org.Kind.ID,
		CreateAppID:     oa.SimpleAudit.First.App.ID,
		CreateUserID:    oa.SimpleAudit.First.User.NullUUID(),
		CreateTimestamp: oa.SimpleAudit.First.Moment,
		UpdateAppID:     oa.SimpleAudit.Last.App.ID,
		UpdateUserID:    oa.SimpleAudit.Last.User.NullUUID(),
		UpdateTimestamp: oa.SimpleAudit.Last.Moment,
	}
}

// OrgService is a service for updating, reading and deleting an Org
type OrgService struct {
	Datastorer Datastorer
}

// UpdateOrgRequest is the request struct for Updating an Org
type UpdateOrgRequest struct {
	ExternalID  string
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Update is used to update an Org
func (s OrgService) Update(ctx context.Context, r *UpdateOrgRequest, adt audit.Audit) (or OrgResponse, err error) {

	// retrieve existing Org
	var (
		oa orgAudit
	)
	oa, err = findOrgByExternalIDWithAudit(ctx, s.Datastorer.Pool(), r.ExternalID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return OrgResponse{}, errs.E(errs.Validation, "No org exists for the given external ID")
		}
		return OrgResponse{}, errs.E(errs.Database, err)
	}
	// overwrite Last audit with the current audit
	oa.SimpleAudit.Last = adt

	// override fields with data from request
	oa.Org.Name = r.Name
	oa.Org.Description = r.Description

	params := orgstore.UpdateOrgParams{
		OrgID:           oa.Org.ID,
		OrgName:         oa.Org.Name,
		OrgDescription:  oa.Org.Description,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    adt.User.NullUUID(),
		UpdateTimestamp: adt.Moment,
	}

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return OrgResponse{}, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	// update database record using orgstore
	var rowsAffected int64
	rowsAffected, err = orgstore.New(tx).UpdateOrg(ctx, params)
	if err != nil {
		return OrgResponse{}, errs.E(errs.Database, err)
	}

	// update should only update exactly one record
	if rowsAffected != 1 {
		return OrgResponse{}, errs.E(errs.Database, fmt.Sprintf("UpdateOrg() should update 1 row, actual: %d", rowsAffected))
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return OrgResponse{}, err
	}

	return newOrgResponse(oa, appAudit{}), nil
}

// Delete is used to delete an Org
func (s OrgService) Delete(ctx context.Context, extlID string) (dr DeleteResponse, err error) {

	// retrieve existing Org
	var o org.Org
	o, err = findOrgByExternalID(ctx, s.Datastorer.Pool(), extlID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return DeleteResponse{}, errs.E(errs.Validation, "No org exists for the given external ID")
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

	var apps []appstore.App
	apps, err = appstore.New(tx).FindAppsByOrg(ctx, o.ID)
	if err != nil {
		return DeleteResponse{}, errs.E(errs.Database, err)
	}

	for _, aa := range apps {
		a := app.App{ID: aa.AppID}
		err = deleteAppTx(ctx, tx, a)
		if err != nil {
			return DeleteResponse{}, errs.E(errs.Database, err)
		}
	}

	var rowsAffected int64
	rowsAffected, err = orgstore.New(tx).DeleteOrg(ctx, o.ID)
	if err != nil {
		return DeleteResponse{}, errs.E(errs.Database, err)
	}

	if rowsAffected != 1 {
		return DeleteResponse{}, errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))

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

// FindAll is used to list all orgs in the datastore
func (s OrgService) FindAll(ctx context.Context) ([]OrgResponse, error) {

	var (
		rows      []orgstore.FindOrgsWithAuditRow
		responses []OrgResponse
		err       error
	)

	dbtx := s.Datastorer.Pool()
	rows, err = orgstore.New(dbtx).FindOrgsWithAudit(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, err)
	}

	for _, row := range rows {
		o := org.Org{
			ID:          row.OrgID,
			ExternalID:  secure.MustParseIdentifier(row.OrgExtlID),
			Name:        row.OrgName,
			Description: row.OrgDescription,
			Kind: org.Kind{
				ID:          row.OrgKindID,
				ExternalID:  row.OrgKindExtlID,
				Description: row.OrgKindDesc,
			},
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
		or := newOrgResponse(orgAudit{Org: o, SimpleAudit: sa}, appAudit{})

		responses = append(responses, or)
	}

	return responses, nil
}

// FindByExternalID is used to find an Org by its External ID
func (s OrgService) FindByExternalID(ctx context.Context, extlID string) (OrgResponse, error) {

	dbtx := s.Datastorer.Pool()

	oa, err := findOrgByExternalIDWithAudit(ctx, dbtx, extlID)
	if err != nil {
		return OrgResponse{}, err
	}

	return newOrgResponse(oa, appAudit{}), nil
}

// findOrgByID retrieves an Org from the datastore given a unique ID
func findOrgByID(ctx context.Context, dbtx DBTX, id uuid.UUID) (org.Org, error) {
	dbo, err := orgstore.New(dbtx).FindOrgByID(ctx, id)
	if err != nil {
		return org.Org{}, errs.E(errs.Database, err)
	}

	o := org.Org{
		ID:          dbo.OrgID,
		ExternalID:  secure.MustParseIdentifier(dbo.OrgExtlID),
		Name:        dbo.OrgName,
		Description: dbo.OrgDescription,
		Kind: org.Kind{
			ID:          dbo.OrgKindID,
			ExternalID:  dbo.OrgKindExtlID,
			Description: dbo.OrgKindDesc,
		},
	}

	return o, nil
}

// findOrgByExternalID retrieves an Org from the datastore given a unique external ID
func findOrgByExternalID(ctx context.Context, dbtx DBTX, extlID string) (org.Org, error) {
	row, err := orgstore.New(dbtx).FindOrgByExtlID(ctx, extlID)
	if err != nil {
		return org.Org{}, errs.E(errs.Database, err)
	}

	o := org.Org{
		ID:          row.OrgID,
		ExternalID:  secure.MustParseIdentifier(row.OrgExtlID),
		Name:        row.OrgName,
		Description: row.OrgDescription,
		Kind: org.Kind{
			ID:          row.OrgKindID,
			ExternalID:  row.OrgKindExtlID,
			Description: row.OrgKindDesc,
		},
	}

	return o, nil
}

// findOrgByExternalID retrieves Org data from the datastore given a unique external ID.
// This data is then hydrated into the org.Org struct along with the simple audit struct
func findOrgByExternalIDWithAudit(ctx context.Context, dbtx DBTX, extlID string) (orgAudit, error) {
	var (
		row orgstore.FindOrgByExtlIDWithAuditRow
		err error
	)

	row, err = orgstore.New(dbtx).FindOrgByExtlIDWithAudit(ctx, extlID)
	if err != nil {
		return orgAudit{}, errs.E(errs.Database, err)
	}

	o := org.Org{
		ID:          row.OrgID,
		ExternalID:  secure.MustParseIdentifier(row.OrgExtlID),
		Name:        row.OrgName,
		Description: row.OrgDescription,
		Kind: org.Kind{
			ID:          row.OrgKindID,
			ExternalID:  row.OrgKindExtlID,
			Description: row.OrgKindDesc,
		},
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

	return orgAudit{Org: o, SimpleAudit: sa}, nil
}

// findOrgKindByExtlID finds an org kind from the datastore given its External ID
func findOrgKindByExtlID(ctx context.Context, dbtx DBTX, extlID string) (org.Kind, error) {
	kind, err := orgstore.New(dbtx).FindOrgKindByExtlID(ctx, extlID)
	if err != nil {
		return org.Kind{}, errs.E(errs.Database, err)
	}

	orgKind := org.Kind{
		ID:          kind.OrgKindID,
		ExternalID:  kind.OrgKindExtlID,
		Description: kind.OrgKindDesc,
	}

	return orgKind, nil
}

// createGenesisOrgKind initializes the org_kind lookup table with the genesis kind record
func createGenesisOrgKind(ctx context.Context, tx pgx.Tx, adt audit.Audit) (orgstore.CreateOrgKindParams, error) {
	genesisParams := orgstore.CreateOrgKindParams{
		OrgKindID:       uuid.New(),
		OrgKindExtlID:   genesisOrgKind,
		OrgKindDesc:     "The Genesis org represents the first organization created in the database and exists purely for the administrative purpose of creating other organizations, apps and users.",
		CreateAppID:     adt.App.ID,
		CreateUserID:    adt.User.NullUUID(),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    adt.User.NullUUID(),
		UpdateTimestamp: adt.Moment,
	}

	var (
		rowsAffected int64
		err          error
	)
	rowsAffected, err = orgstore.New(tx).CreateOrgKind(ctx, genesisParams)
	if err != nil {
		return orgstore.CreateOrgKindParams{}, errs.E(errs.Database, err)
	}

	if rowsAffected != 1 {
		return orgstore.CreateOrgKindParams{}, errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	return genesisParams, nil
}

// createTestOrgKind initializes the org_kind lookup table with the test kind record
func createTestOrgKind(ctx context.Context, tx pgx.Tx, adt audit.Audit) (orgstore.CreateOrgKindParams, error) {
	testParams := orgstore.CreateOrgKindParams{
		OrgKindID:       uuid.New(),
		OrgKindExtlID:   "test",
		OrgKindDesc:     "The test org is used strictly for testing",
		CreateAppID:     adt.App.ID,
		CreateUserID:    adt.User.NullUUID(),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    adt.User.NullUUID(),
		UpdateTimestamp: adt.Moment,
	}

	var (
		rowsAffected int64
		err          error
	)
	rowsAffected, err = orgstore.New(tx).CreateOrgKind(ctx, testParams)
	if err != nil {
		return orgstore.CreateOrgKindParams{}, errs.E(errs.Database, err)
	}

	if rowsAffected != 1 {
		return orgstore.CreateOrgKindParams{}, errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	return testParams, nil
}

// createStandardOrgKind initializes the org_kind lookup table with the standard kind record
func createStandardOrgKind(ctx context.Context, tx pgx.Tx, adt audit.Audit) (orgstore.CreateOrgKindParams, error) {
	standardParams := orgstore.CreateOrgKindParams{
		OrgKindID:       uuid.New(),
		OrgKindExtlID:   "standard",
		OrgKindDesc:     "The standard org is used for myriad business purposes",
		CreateAppID:     adt.App.ID,
		CreateUserID:    adt.User.NullUUID(),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    adt.User.NullUUID(),
		UpdateTimestamp: adt.Moment,
	}

	var (
		rowsAffected int64
		err          error
	)
	rowsAffected, err = orgstore.New(tx).CreateOrgKind(ctx, standardParams)
	if err != nil {
		return orgstore.CreateOrgKindParams{}, errs.E(errs.Database, err)
	}

	if rowsAffected != 1 {
		return orgstore.CreateOrgKindParams{}, errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	return standardParams, nil
}
