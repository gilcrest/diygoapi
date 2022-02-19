package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/datastore/orgstore"
	"github.com/gilcrest/go-api-basic/domain/app"
	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/org"
	"github.com/gilcrest/go-api-basic/domain/secure"
	"github.com/gilcrest/go-api-basic/domain/user"
)

// CreateOrgRequest is the request struct for Creating an Org
type CreateOrgRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Kind        string `json:"kind"`
}

// OrgResponse is the response struct for an Org
type OrgResponse struct {
	ExternalID  string        `json:"external_id"`
	Name        string        `json:"name"`
	Kind        string        `json:"kind"`
	Description string        `json:"description"`
	CreateAudit auditResponse `json:"create_audit"`
	UpdateAudit auditResponse `json:"update_audit"`
}

// newOrgResponse initializes OrgResponse given an org.Org.
func newOrgResponse(o org.Org, sa audit.SimpleAudit) OrgResponse {
	return OrgResponse{
		ExternalID:  o.ExternalID.String(),
		Name:        o.Name,
		Description: o.Description,
		Kind:        o.Kind.ExternalID,
		CreateAudit: newAuditResponse(sa.First),
		UpdateAudit: newAuditResponse(sa.Last),
	}
}

// CreateOrgService is a service for creating an Org
type CreateOrgService struct {
	Datastorer Datastorer
}

// Create is used to create an Org
func (s CreateOrgService) Create(ctx context.Context, r *CreateOrgRequest, adt audit.Audit) (OrgResponse, error) {
	var err error

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

	sa := audit.SimpleAudit{
		First: adt,
		Last:  adt,
	}

	// start db txn using pgxpool
	tx, err := s.Datastorer.BeginTx(ctx)
	if err != nil {
		return OrgResponse{}, err
	}

	err = createOrgDB(ctx, s.Datastorer, tx, o, sa)
	if err != nil {
		return OrgResponse{}, err
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return OrgResponse{}, err
	}

	return newOrgResponse(o, sa), nil
}

// created separate function as it's used by genesis service as well
func createOrgDB(ctx context.Context, ds Datastorer, tx pgx.Tx, o org.Org, sa audit.SimpleAudit) error {
	if o.Kind.ID == uuid.Nil {
		return errs.E(errs.Database, ds.RollbackTx(ctx, tx, errs.E("org Kind is required")))
	}

	// create database record using orgstore
	_, err := orgstore.New(tx).CreateOrg(ctx, newCreateOrgParams(o, sa))
	if err != nil {
		return errs.E(errs.Database, ds.RollbackTx(ctx, tx, err))
	}

	return nil
}

// newCreateOrgParams maps an Org to orgstore.CreateOrgParams
func newCreateOrgParams(o org.Org, simpleAudit audit.SimpleAudit) orgstore.CreateOrgParams {
	return orgstore.CreateOrgParams{
		OrgID:           o.ID,
		OrgExtlID:       o.ExternalID.String(),
		OrgName:         o.Name,
		OrgDescription:  o.Description,
		OrgKindID:       o.Kind.ID,
		CreateAppID:     simpleAudit.First.App.ID,
		CreateUserID:    datastore.NewNullUUID(simpleAudit.First.User.ID),
		CreateTimestamp: simpleAudit.First.Moment,
		UpdateAppID:     simpleAudit.Last.App.ID,
		UpdateUserID:    datastore.NewNullUUID(simpleAudit.Last.User.ID),
		UpdateTimestamp: simpleAudit.Last.Moment,
	}
}

// newUpdateOrgParams maps an Org to orgstore.UpdateOrgParams
func newUpdateOrgParams(o org.Org, ua audit.Audit) orgstore.UpdateOrgParams {
	return orgstore.UpdateOrgParams{
		OrgID:           o.ID,
		OrgName:         o.Name,
		OrgDescription:  o.Description,
		UpdateAppID:     ua.App.ID,
		UpdateUserID:    datastore.NewNullUUID(ua.User.ID),
		UpdateTimestamp: ua.Moment,
	}
}

// UpdateOrgRequest is the request struct for Updating an Org
type UpdateOrgRequest struct {
	ExternalID  string
	Name        string `json:"name"`
	Description string `json:"description"`
}

// UpdateOrgService is a service for updating an Org
type UpdateOrgService struct {
	Datastorer Datastorer
}

// Update is used to update an Org
func (cos UpdateOrgService) Update(ctx context.Context, r *UpdateOrgRequest, adt audit.Audit) (OrgResponse, error) {
	// start db txn using pgxpool
	tx, err := cos.Datastorer.BeginTx(ctx)
	if err != nil {
		return OrgResponse{}, err
	}

	// retrieve existing Org
	o, sa, err := findOrgByExternalID(ctx, tx, r.ExternalID, true)
	if err != nil {
		if err == pgx.ErrNoRows {
			return OrgResponse{}, errs.E(errs.Validation, "No org exists for the given external ID")
		}
		return OrgResponse{}, errs.E(errs.Database, err)
	}

	// override fields with data from request
	o.Name = r.Name
	o.Description = r.Description

	// update database record using orgstore
	err = orgstore.New(tx).UpdateOrg(ctx, newUpdateOrgParams(o, adt))
	if err != nil {
		return OrgResponse{}, errs.E(errs.Database, cos.Datastorer.RollbackTx(ctx, tx, err))
	}

	// commit db txn using pgxpool
	err = cos.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return OrgResponse{}, err
	}

	sa.Last = adt
	or := newOrgResponse(o, *sa)

	return or, nil
}

// FindOrgService interface reads Orgs form the datastore
type FindOrgService struct {
	Datastorer Datastorer
}

// FindAll is used to list all orgs in the datastore
func (fos FindOrgService) FindAll(ctx context.Context) ([]OrgResponse, error) {

	var (
		dbos     []orgstore.Org
		response []OrgResponse
		err      error
	)

	dbtx := fos.Datastorer.Pool()
	dbos, err = orgstore.New(dbtx).FindOrgs(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, err)
	}

	for _, dbo := range dbos {
		var o org.Org
		sa := new(audit.SimpleAudit)
		o, sa, err = hydrateOrgFromDB(ctx, dbtx, dbo, true)
		if err != nil {
			return nil, err
		}
		or := newOrgResponse(o, *sa)

		response = append(response, or)
	}

	return response, nil
}

// FindByExternalID is used to find an Org by its External ID
func (fos FindOrgService) FindByExternalID(ctx context.Context, extlID string) (OrgResponse, error) {

	dbtx := fos.Datastorer.Pool()

	o, sa, err := findOrgByExternalID(ctx, dbtx, extlID, true)
	if err != nil {
		return OrgResponse{}, err
	}

	return newOrgResponse(o, *sa), nil
}

// findOrgByID retrieves an Org from the datastore given a unique ID
func findOrgByID(ctx context.Context, dbtx DBTX, id uuid.UUID, withAudit bool) (org.Org, *audit.SimpleAudit, error) {
	dbo, err := orgstore.New(dbtx).FindOrgByID(ctx, id)
	if err != nil {
		return org.Org{}, nil, errs.E(errs.Database, err)
	}

	return hydrateOrgFromDB(ctx, dbtx, dbo, withAudit)
}

// findOrgByExternalID retrieves an Org from the datastore given a unique external ID
// if withAudit is set to true, audit.SimpleAudit will be returned, otherwise nil is returned
func findOrgByExternalID(ctx context.Context, dbtx DBTX, extlID string, withAudit bool) (org.Org, *audit.SimpleAudit, error) {
	dbo, err := orgstore.New(dbtx).FindOrgByExtlID(ctx, extlID)
	if err != nil {
		return org.Org{}, nil, errs.E(errs.Database, err)
	}

	return hydrateOrgFromDB(ctx, dbtx, dbo, withAudit)
}

// hydrateOrgFromDB populates an org.Org and an audit.SimpleAudit (if withAudit is true) given an orgstore.Org
func hydrateOrgFromDB(ctx context.Context, dbtx DBTX, dbo orgstore.Org, withAudit bool) (org.Org, *audit.SimpleAudit, error) {
	var (
		extl                   secure.Identifier
		err                    error
		createApp, updateApp   app.App
		createUser, updateUser user.User
	)

	extl, err = secure.ParseIdentifier(dbo.OrgExtlID)
	if err != nil {
		return org.Org{}, nil, err
	}

	o := org.Org{
		ID:          dbo.OrgID,
		ExternalID:  extl,
		Name:        dbo.OrgName,
		Description: dbo.OrgDescription,
	}

	sa := new(audit.SimpleAudit)
	if withAudit {
		createApp, _, err = findAppByID(ctx, dbtx, dbo.CreateAppID, false)
		if err != nil {
			return org.Org{}, nil, err
		}
		createUser, err = findUserByID(ctx, dbtx, dbo.CreateUserID.UUID)
		if err != nil {
			return org.Org{}, nil, err
		}
		createAudit := audit.Audit{
			App:    createApp,
			User:   createUser,
			Moment: dbo.CreateTimestamp,
		}

		updateApp, _, err = findAppByID(ctx, dbtx, dbo.UpdateAppID, false)
		if err != nil {
			return org.Org{}, nil, err
		}
		updateUser, err = findUserByID(ctx, dbtx, dbo.UpdateUserID.UUID)
		if err != nil {
			return org.Org{}, nil, err
		}
		updateAudit := audit.Audit{
			App:    updateApp,
			User:   updateUser,
			Moment: dbo.UpdateTimestamp,
		}

		sa = &audit.SimpleAudit{First: createAudit, Last: updateAudit}
	}

	return o, sa, nil
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
func createGenesisOrgKind(ctx context.Context, ds Datastorer, tx pgx.Tx, adt audit.Audit) (orgstore.CreateOrgKindParams, error) {
	genesisParams := orgstore.CreateOrgKindParams{
		OrgKindID:       uuid.New(),
		OrgKindExtlID:   genesisOrgTypeString,
		OrgKindDesc:     "The Genesis org represents the first organization created in the database and exists purely for the administrative purpose of creating other organizations, apps and users.",
		CreateAppID:     adt.App.ID,
		CreateUserID:    datastore.NewNullUUID(adt.User.ID),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    datastore.NewNullUUID(adt.User.ID),
		UpdateTimestamp: adt.Moment,
	}

	_, err := orgstore.New(tx).CreateOrgKind(ctx, genesisParams)
	if err != nil {
		return orgstore.CreateOrgKindParams{}, errs.E(errs.Database, ds.RollbackTx(ctx, tx, err))
	}

	return genesisParams, nil
}

// createTestOrgKind initializes the org_kind lookup table with the test kind record
func createTestOrgKind(ctx context.Context, ds Datastorer, tx pgx.Tx, adt audit.Audit) (orgstore.CreateOrgKindParams, error) {
	testParams := orgstore.CreateOrgKindParams{
		OrgKindID:       uuid.New(),
		OrgKindExtlID:   "test",
		OrgKindDesc:     "The test org is used strictly for testing",
		CreateAppID:     adt.App.ID,
		CreateUserID:    datastore.NewNullUUID(adt.User.ID),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    datastore.NewNullUUID(adt.User.ID),
		UpdateTimestamp: adt.Moment,
	}

	_, err := orgstore.New(tx).CreateOrgKind(ctx, testParams)
	if err != nil {
		return orgstore.CreateOrgKindParams{}, errs.E(errs.Database, ds.RollbackTx(ctx, tx, err))
	}

	return testParams, nil
}

// createStandardOrgKind initializes the org_kind lookup table with the standard kind record
func createStandardOrgKind(ctx context.Context, ds Datastorer, tx pgx.Tx, adt audit.Audit) (err error) {
	standardParams := orgstore.CreateOrgKindParams{
		OrgKindID:       uuid.New(),
		OrgKindExtlID:   "standard",
		OrgKindDesc:     "The standard org is used for myriad business purposes",
		CreateAppID:     adt.App.ID,
		CreateUserID:    datastore.NewNullUUID(adt.User.ID),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    datastore.NewNullUUID(adt.User.ID),
		UpdateTimestamp: adt.Moment,
	}

	_, err = orgstore.New(tx).CreateOrgKind(ctx, standardParams)
	if err != nil {
		return errs.E(errs.Database, ds.RollbackTx(ctx, tx, err))
	}

	return nil
}
