package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"

	"github.com/gilcrest/diy-go-api"
	"github.com/gilcrest/diy-go-api/errs"
	"github.com/gilcrest/diy-go-api/secure"
	"github.com/gilcrest/diy-go-api/sqldb/datastore"
)

// orgAudit is the combination of a domain Org and its audit data
type orgAudit struct {
	Org         *diy.Org
	SimpleAudit *diy.SimpleAudit
}

// newOrgResponse initializes OrgResponse given a diy.Org.
func newOrgResponse(oa *orgAudit, aa appAudit) *diy.OrgResponse {
	r := &diy.OrgResponse{
		ExternalID:          oa.Org.ExternalID.String(),
		Name:                oa.Org.Name,
		Description:         oa.Org.Description,
		KindExternalID:      oa.Org.Kind.ExternalID,
		CreateAppExtlID:     oa.SimpleAudit.Create.App.ExternalID.String(),
		CreateUserFirstName: oa.SimpleAudit.Create.User.FirstName,
		CreateUserLastName:  oa.SimpleAudit.Create.User.LastName,
		CreateDateTime:      oa.SimpleAudit.Create.Moment.Format(time.RFC3339),
		UpdateAppExtlID:     oa.SimpleAudit.Update.App.ExternalID.String(),
		UpdateUserFirstName: oa.SimpleAudit.Update.User.FirstName,
		UpdateUserLastName:  oa.SimpleAudit.Update.User.LastName,
		UpdateDateTime:      oa.SimpleAudit.Update.Moment.Format(time.RFC3339),
	}

	if aa.App != nil {
		r.App = newAppResponse(aa)
	}

	return r
}

// OrgService is a service for updating, reading and deleting an Org
type OrgService struct {
	Datastorer      diy.Datastorer
	APIKeyGenerator diy.APIKeyGenerator
	EncryptionKey   *[32]byte
}

// Create is used to create an Org
func (s *OrgService) Create(ctx context.Context, r *diy.CreateOrgRequest, adt diy.Audit) (or *diy.OrgResponse, err error) {

	if r == nil {
		return nil, errs.E(errs.Validation, "CreateOrgRequest must have a value when creating an Org")
	}
	err = r.Validate()
	if err != nil {
		return nil, err
	}

	// ensure that CreateAppRequest field is not nil or not the zero value
	if (r.CreateAppRequest != nil) || (*r.CreateAppRequest == (diy.CreateAppRequest{})) {
		return nil, errs.E(errs.Validation, "CreateAppRequest must have a value when creating an Org")
	}

	err = r.CreateAppRequest.Validate()
	if err != nil {
		return nil, err
	}

	sa := &diy.SimpleAudit{
		Create: adt,
		Update: adt,
	}

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	var kind *diy.OrgKind
	kind, err = findOrgKindByExtlID(ctx, tx, r.Kind)
	if err != nil {
		return nil, err
	}

	// initialize Org and inject dependent fields
	o := &diy.Org{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Name:        r.Name,
		Description: r.Description,
		Kind:        kind,
	}
	oa := &orgAudit{
		Org:         o,
		SimpleAudit: sa,
	}

	// if there is an app request along with the Org request, process it as well
	var (
		a        *diy.App
		aa       appAudit
		provider diy.Provider
	)
	provider = diy.ParseProvider(r.CreateAppRequest.Oauth2Provider)

	err = r.CreateAppRequest.Validate()
	if err != nil {
		return nil, err
	}
	nap := newAppParams{
		Name:             r.CreateAppRequest.Name,
		Description:      r.CreateAppRequest.Description,
		Org:              o,
		ApiKeyGenerator:  s.APIKeyGenerator,
		EncryptionKey:    s.EncryptionKey,
		Provider:         provider,
		ProviderClientID: r.CreateAppRequest.Oauth2ProviderClientID,
	}
	a, err = newApp(nap)
	if err != nil {
		return nil, err
	}
	aa = appAudit{
		App:         a,
		SimpleAudit: sa,
	}

	// write org to the db
	err = createOrgTx(ctx, tx, oa)
	if err != nil {
		return nil, err
	}

	// if app is also to be created, write it to the db
	if aa.App != nil {
		err = createAppTx(ctx, tx, aa)
		if err != nil {
			return nil, err
		}
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return nil, err
	}

	response := newOrgResponse(oa, aa)

	return response, nil
}

// createOrgTx writes an Org and its audit information to the database.
// separate function as it's used by genesis service as well
func createOrgTx(ctx context.Context, tx pgx.Tx, oa *orgAudit) error {
	if oa.Org.Kind.ID == uuid.Nil {
		return errs.E("org Kind is required")
	}

	// create database record using datastore
	rowsAffected, err := datastore.New(tx).CreateOrg(ctx, newCreateOrgParams(oa))
	if err != nil {
		return errs.E(errs.Database, err)
	}

	// update should only update exactly one record
	if rowsAffected != 1 {
		return errs.E(errs.Database, fmt.Sprintf("CreateOrg() should insert 1 row, actual: %d", rowsAffected))
	}

	return nil
}

// newCreateOrgParams maps an Org to datastore.CreateOrgParams
func newCreateOrgParams(oa *orgAudit) datastore.CreateOrgParams {
	return datastore.CreateOrgParams{
		OrgID:           oa.Org.ID,
		OrgExtlID:       oa.Org.ExternalID.String(),
		OrgName:         oa.Org.Name,
		OrgDescription:  oa.Org.Description,
		OrgKindID:       oa.Org.Kind.ID,
		CreateAppID:     oa.SimpleAudit.Create.App.ID,
		CreateUserID:    oa.SimpleAudit.Create.User.NullUUID(),
		CreateTimestamp: oa.SimpleAudit.Create.Moment,
		UpdateAppID:     oa.SimpleAudit.Update.App.ID,
		UpdateUserID:    oa.SimpleAudit.Update.User.NullUUID(),
		UpdateTimestamp: oa.SimpleAudit.Update.Moment,
	}
}

// Update is used to update an Org
func (s *OrgService) Update(ctx context.Context, r *diy.UpdateOrgRequest, adt diy.Audit) (or *diy.OrgResponse, err error) {

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	// retrieve existing Org
	var oa *orgAudit
	oa, err = findOrgByExternalIDWithAudit(ctx, tx, r.ExternalID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.E(errs.Validation, "No org exists for the given external ID")
		}
		return nil, errs.E(errs.Database, err)
	}
	// overwrite Last audit with the current audit
	oa.SimpleAudit.Update = adt

	// override fields with data from request
	oa.Org.Name = r.Name
	oa.Org.Description = r.Description

	params := datastore.UpdateOrgParams{
		OrgID:           oa.Org.ID,
		OrgName:         oa.Org.Name,
		OrgDescription:  oa.Org.Description,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    adt.User.NullUUID(),
		UpdateTimestamp: adt.Moment,
	}

	// update database record using datastore
	var rowsAffected int64
	rowsAffected, err = datastore.New(tx).UpdateOrg(ctx, params)
	if err != nil {
		return nil, errs.E(errs.Database, err)
	}

	// update should only update exactly one record
	if rowsAffected != 1 {
		return nil, errs.E(errs.Database, fmt.Sprintf("UpdateOrg() should update 1 row, actual: %d", rowsAffected))
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return nil, err
	}

	return newOrgResponse(oa, appAudit{}), nil
}

// Delete is used to delete an Org
func (s *OrgService) Delete(ctx context.Context, extlID string) (dr diy.DeleteResponse, err error) {

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return diy.DeleteResponse{}, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	// retrieve existing Org
	var o diy.Org
	o, err = findOrgByExternalID(ctx, tx, extlID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return diy.DeleteResponse{}, errs.E(errs.Validation, "No org exists for the given external ID")
		}
		return diy.DeleteResponse{}, errs.E(errs.Database, err)
	}

	var dbApps []datastore.App
	dbApps, err = datastore.New(tx).FindAppsByOrg(ctx, o.ID)
	if err != nil {
		return diy.DeleteResponse{}, errs.E(errs.Database, err)
	}

	for _, aa := range dbApps {
		a := diy.App{ID: aa.AppID}
		err = deleteAppTx(ctx, tx, a)
		if err != nil {
			return diy.DeleteResponse{}, errs.E(errs.Database, err)
		}
	}

	var rowsAffected int64
	rowsAffected, err = datastore.New(tx).DeleteOrg(ctx, o.ID)
	if err != nil {
		return diy.DeleteResponse{}, errs.E(errs.Database, err)
	}

	if rowsAffected != 1 {
		return diy.DeleteResponse{}, errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return diy.DeleteResponse{}, err
	}

	response := diy.DeleteResponse{
		ExternalID: extlID,
		Deleted:    true,
	}

	return response, nil
}

// FindAll is used to list all orgs in the datastore
func (s *OrgService) FindAll(ctx context.Context) (responses []*diy.OrgResponse, err error) {

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	var (
		rows []datastore.FindOrgsWithAuditRow
	)

	rows, err = datastore.New(tx).FindOrgsWithAudit(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, err)
	}

	for _, row := range rows {
		o := diy.Org{
			ID:          row.OrgID,
			ExternalID:  secure.MustParseIdentifier(row.OrgExtlID),
			Name:        row.OrgName,
			Description: row.OrgDescription,
			Kind: &diy.OrgKind{
				ID:          row.OrgKindID,
				ExternalID:  row.OrgKindExtlID,
				Description: row.OrgKindDesc,
			},
		}

		sa := diy.SimpleAudit{
			Create: diy.Audit{
				App: &diy.App{
					ID:          row.CreateAppID,
					ExternalID:  secure.MustParseIdentifier(row.CreateAppExtlID),
					Org:         &diy.Org{ID: row.CreateAppOrgID},
					Name:        row.CreateAppName,
					Description: row.CreateAppDescription,
					APIKeys:     nil,
				},
				User: &diy.User{
					ID:        row.CreateUserID.UUID,
					FirstName: row.CreateUserFirstName,
					LastName:  row.CreateUserLastName,
				},
				Moment: row.CreateTimestamp,
			},
			Update: diy.Audit{
				App: &diy.App{
					ID:          row.UpdateAppID,
					ExternalID:  secure.MustParseIdentifier(row.UpdateAppExtlID),
					Org:         &diy.Org{ID: row.UpdateAppOrgID},
					Name:        row.UpdateAppName,
					Description: row.UpdateAppDescription,
					APIKeys:     nil,
				},
				User: &diy.User{
					ID:        row.UpdateUserID.UUID,
					FirstName: row.UpdateUserFirstName,
					LastName:  row.UpdateUserLastName,
				},
				Moment: row.UpdateTimestamp,
			},
		}
		or := newOrgResponse(&orgAudit{Org: &o, SimpleAudit: &sa}, appAudit{})

		responses = append(responses, or)
	}

	return responses, nil
}

// FindByExternalID is used to find an Org by its External ID
func (s *OrgService) FindByExternalID(ctx context.Context, extlID string) (or *diy.OrgResponse, err error) {

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	var oa *orgAudit
	oa, err = findOrgByExternalIDWithAudit(ctx, tx, extlID)
	if err != nil {
		return nil, err
	}

	return newOrgResponse(oa, appAudit{}), nil
}

// findOrgByExternalID retrieves an Org from the datastore given a unique external ID
func findOrgByExternalID(ctx context.Context, dbtx diy.DBTX, extlID string) (diy.Org, error) {
	row, err := datastore.New(dbtx).FindOrgByExtlID(ctx, extlID)
	if err != nil {
		return diy.Org{}, errs.E(errs.Database, err)
	}

	o := diy.Org{
		ID:          row.OrgID,
		ExternalID:  secure.MustParseIdentifier(row.OrgExtlID),
		Name:        row.OrgName,
		Description: row.OrgDescription,
		Kind: &diy.OrgKind{
			ID:          row.OrgKindID,
			ExternalID:  row.OrgKindExtlID,
			Description: row.OrgKindDesc,
		},
	}

	return o, nil
}

// findOrgByExternalID retrieves Org data from the datastore given a unique external ID.
// This data is then hydrated into the diy.Org struct along with the simple audit struct
func findOrgByExternalIDWithAudit(ctx context.Context, dbtx diy.DBTX, extlID string) (*orgAudit, error) {
	var (
		row datastore.FindOrgByExtlIDWithAuditRow
		err error
	)

	row, err = datastore.New(dbtx).FindOrgByExtlIDWithAudit(ctx, extlID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.E(errs.NotExist, fmt.Sprintf("no org found with external ID: %s", extlID))
		} else {
			return nil, errs.E(errs.Database, err)
		}
	}

	o := &diy.Org{
		ID:          row.OrgID,
		ExternalID:  secure.MustParseIdentifier(row.OrgExtlID),
		Name:        row.OrgName,
		Description: row.OrgDescription,
		Kind: &diy.OrgKind{
			ID:          row.OrgKindID,
			ExternalID:  row.OrgKindExtlID,
			Description: row.OrgKindDesc,
		},
	}

	sa := &diy.SimpleAudit{
		Create: diy.Audit{
			App: &diy.App{
				ID:          row.CreateAppID,
				ExternalID:  secure.MustParseIdentifier(row.CreateAppExtlID),
				Org:         &diy.Org{ID: row.CreateAppOrgID},
				Name:        row.CreateAppName,
				Description: row.CreateAppDescription,
				APIKeys:     nil,
			},
			User: &diy.User{
				ID:        row.CreateUserID.UUID,
				FirstName: row.CreateUserFirstName,
				LastName:  row.CreateUserLastName,
			},
			Moment: row.CreateTimestamp,
		},
		Update: diy.Audit{
			App: &diy.App{
				ID:          row.UpdateAppID,
				ExternalID:  secure.MustParseIdentifier(row.UpdateAppExtlID),
				Org:         &diy.Org{ID: row.UpdateAppOrgID},
				Name:        row.UpdateAppName,
				Description: row.UpdateAppDescription,
				APIKeys:     nil,
			},
			User: &diy.User{
				ID:        row.UpdateUserID.UUID,
				FirstName: row.UpdateUserFirstName,
				LastName:  row.UpdateUserLastName,
			},
			Moment: row.UpdateTimestamp,
		},
	}

	return &orgAudit{Org: o, SimpleAudit: sa}, nil
}

// FindOrgByName finds an Org in the database using its unique name.
func FindOrgByName(ctx context.Context, tx datastore.DBTX, name string) (*diy.Org, error) {
	findOrgByNameRow, err := datastore.New(tx).FindOrgByName(ctx, name)
	if err != nil {
		return nil, errs.E(errs.Database, err)
	}

	o := &diy.Org{
		ID:          findOrgByNameRow.OrgID,
		ExternalID:  secure.MustParseIdentifier(findOrgByNameRow.OrgExtlID),
		Name:        findOrgByNameRow.OrgName,
		Description: findOrgByNameRow.OrgDescription,
		Kind: &diy.OrgKind{
			ID:          findOrgByNameRow.OrgKindID,
			ExternalID:  findOrgByNameRow.OrgKindExtlID,
			Description: findOrgByNameRow.OrgKindDesc,
		},
	}

	return o, nil
}

// findOrgKindByExtlID finds an org kind from the datastore given its External ID
func findOrgKindByExtlID(ctx context.Context, dbtx diy.DBTX, extlID string) (*diy.OrgKind, error) {
	kind, err := datastore.New(dbtx).FindOrgKindByExtlID(ctx, extlID)
	if err != nil {
		return nil, errs.E(errs.Database, err)
	}

	orgKind := &diy.OrgKind{
		ID:          kind.OrgKindID,
		ExternalID:  kind.OrgKindExtlID,
		Description: kind.OrgKindDesc,
	}

	return orgKind, nil
}

// createPrincipalOrgKind initializes the org_kind lookup table with the genesis kind record
func createPrincipalOrgKind(ctx context.Context, tx pgx.Tx, adt diy.Audit) (datastore.CreateOrgKindParams, error) {
	createOrgKindParams := datastore.CreateOrgKindParams{
		OrgKindID:       uuid.New(),
		OrgKindExtlID:   principalOrgKind,
		OrgKindDesc:     "The Principal org represents the first organization created in the database and exists purely for the administrative purpose of creating other organizations, apps and users.",
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
	rowsAffected, err = datastore.New(tx).CreateOrgKind(ctx, createOrgKindParams)
	if err != nil {
		return datastore.CreateOrgKindParams{}, errs.E(errs.Database, err)
	}

	if rowsAffected != 1 {
		return datastore.CreateOrgKindParams{}, errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	return createOrgKindParams, nil
}

// createTestOrgKind initializes the org_kind lookup table with the test kind record
func createTestOrgKind(ctx context.Context, tx pgx.Tx, adt diy.Audit) (datastore.CreateOrgKindParams, error) {
	testParams := datastore.CreateOrgKindParams{
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
	rowsAffected, err = datastore.New(tx).CreateOrgKind(ctx, testParams)
	if err != nil {
		return datastore.CreateOrgKindParams{}, errs.E(errs.Database, err)
	}

	if rowsAffected != 1 {
		return datastore.CreateOrgKindParams{}, errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	return testParams, nil
}

// createStandardOrgKind initializes the org_kind lookup table with the standard kind record
func createStandardOrgKind(ctx context.Context, tx pgx.Tx, adt diy.Audit) (datastore.CreateOrgKindParams, error) {
	standardParams := datastore.CreateOrgKindParams{
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
	rowsAffected, err = datastore.New(tx).CreateOrgKind(ctx, standardParams)
	if err != nil {
		return datastore.CreateOrgKindParams{}, errs.E(errs.Database, err)
	}

	if rowsAffected != 1 {
		return datastore.CreateOrgKindParams{}, errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	return standardParams, nil
}
