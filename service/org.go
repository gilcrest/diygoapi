package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/datastore/orgstore"
	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/org"
	"github.com/gilcrest/go-api-basic/domain/secure"
)

// Datastorer is an interface for working with the Database
type Datastorer interface {
	// Pool returns *pgxpool.Pool
	Pool() *pgxpool.Pool
	// BeginTx starts a pgx.Tx using the input context
	BeginTx(ctx context.Context) (pgx.Tx, error)
	// RollbackTx rolls back the input pgx.Tx
	RollbackTx(ctx context.Context, tx pgx.Tx, err error) error
	// CommitTx commits the Tx
	CommitTx(ctx context.Context, tx pgx.Tx) error
}

// CreateOrgRequest is the request struct for Creating a Movie
type CreateOrgRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// OrgResponse is the response struct for a Movie
type OrgResponse struct {
	ExternalID  string        `json:"external_id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	CreateAudit auditResponse `json:"create_audit"`
	UpdateAudit auditResponse `json:"update_audit"`
}

// newOrgResponse initializes OrgResponse given an org.Org.
// org.Org does not embed create/update App and User (intentionally),
// so these structs are retrieved from the datastore as well.
func newOrgResponse(ctx context.Context, dbtx DBTX, o org.Org) (OrgResponse, error) {
	createApp, err := findAppByID(ctx, dbtx, o.CreateAppID)
	if err != nil {
		return OrgResponse{}, err
	}
	createUser, err := findUserByID(ctx, dbtx, o.CreateUserID)
	if err != nil {
		return OrgResponse{}, err
	}
	updateApp, err := findAppByID(ctx, dbtx, o.UpdateAppID)
	if err != nil {
		return OrgResponse{}, err
	}
	updateUser, err := findUserByID(ctx, dbtx, o.UpdateUserID)
	if err != nil {
		return OrgResponse{}, err
	}
	ca := audit.Audit{
		App:    createApp,
		User:   createUser,
		Moment: o.CreateTime,
	}

	ua := audit.Audit{
		App:    updateApp,
		User:   updateUser,
		Moment: o.UpdateTime,
	}

	return OrgResponse{
		ExternalID:  o.ExternalID.String(),
		Name:        o.Name,
		Description: o.Description,
		CreateAudit: newAuditResponse(ca),
		UpdateAudit: newAuditResponse(ua),
	}, nil
}

// CreateOrgService is a service for creating an Org
type CreateOrgService struct {
	Datastorer Datastorer
}

// Create is used to create an Org
func (cos CreateOrgService) Create(ctx context.Context, r *CreateOrgRequest, adt audit.Audit) (OrgResponse, error) {

	// initialize Org and inject dependent fields
	o := org.Org{
		ID:           uuid.New(),
		ExternalID:   secure.NewID(),
		Name:         r.Name,
		Description:  r.Description,
		CreateAppID:  adt.App.ID,
		CreateUserID: adt.User.ID,
		CreateTime:   adt.Moment,
		UpdateAppID:  adt.App.ID,
		UpdateUserID: adt.User.ID,
		UpdateTime:   adt.Moment,
	}

	// start db txn using pgxpool
	tx, err := cos.Datastorer.BeginTx(ctx)
	if err != nil {
		return OrgResponse{}, err
	}

	// create database record using orgstore
	_, err = orgstore.New(tx).CreateOrg(ctx, NewCreateOrgParams(o))
	if err != nil {
		return OrgResponse{}, errs.E(errs.Database, cos.Datastorer.RollbackTx(ctx, tx, err))
	}

	// commit db txn using pgxpool
	err = cos.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return OrgResponse{}, err
	}

	or := OrgResponse{
		ExternalID:  o.ExternalID.String(),
		Name:        o.Name,
		Description: o.Description,
		CreateAudit: newAuditResponse(adt),
		UpdateAudit: newAuditResponse(adt),
	}

	return or, nil
}

// NewCreateOrgParams maps an Org to orgstore.CreateOrgParams
func NewCreateOrgParams(o org.Org) orgstore.CreateOrgParams {
	return orgstore.CreateOrgParams{
		OrgID:           uuid.New(),
		OrgExtlID:       o.ExternalID.String(),
		OrgName:         o.Name,
		OrgDescription:  o.Description,
		CreateAppID:     o.CreateAppID,
		CreateUserID:    datastore.NewNullUUID(o.CreateUserID),
		CreateTimestamp: o.CreateTime,
		UpdateAppID:     o.UpdateAppID,
		UpdateUserID:    datastore.NewNullUUID(o.UpdateUserID),
		UpdateTimestamp: o.UpdateTime,
	}
}

// newUpdateOrgParams maps an Org to orgstore.UpdateOrgParams
func newUpdateOrgParams(o org.Org) orgstore.UpdateOrgParams {
	return orgstore.UpdateOrgParams{
		OrgID:           o.ID,
		OrgName:         o.Name,
		OrgDescription:  o.Description,
		UpdateAppID:     o.UpdateAppID,
		UpdateUserID:    datastore.NewNullUUID(o.UpdateUserID),
		UpdateTimestamp: o.UpdateTime,
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
	o, err := findOrgByExternalID(ctx, tx, r.ExternalID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return OrgResponse{}, errs.E(errs.Validation, "No org exists for the given external ID")
		}
		return OrgResponse{}, errs.E(errs.Database, err)
	}

	// override fields with data from request
	o.Name = r.Name
	o.Description = r.Description
	o.UpdateAppID = adt.App.ID
	o.UpdateUserID = adt.User.ID
	o.UpdateTime = adt.Moment

	// update database record using orgstore
	err = orgstore.New(tx).UpdateOrg(ctx, newUpdateOrgParams(o))
	if err != nil {
		return OrgResponse{}, errs.E(errs.Database, cos.Datastorer.RollbackTx(ctx, tx, err))
	}

	var or OrgResponse
	or, err = newOrgResponse(ctx, tx, o)
	if err != nil {
		return OrgResponse{}, err
	}

	// commit db txn using pgxpool
	err = cos.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return OrgResponse{}, err
	}

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
		var (
			or   OrgResponse
			extl secure.Identifier
		)
		extl, err = secure.ParseIdentifier(dbo.OrgExtlID)
		if err != nil {
			return nil, err
		}

		// add data from db
		o := org.Org{
			ID:           dbo.OrgID,
			ExternalID:   extl,
			Name:         dbo.OrgName,
			Description:  dbo.OrgDescription,
			CreateAppID:  dbo.CreateAppID,
			CreateUserID: dbo.CreateUserID.UUID,
			CreateTime:   dbo.CreateTimestamp,
			UpdateAppID:  dbo.UpdateAppID,
			UpdateUserID: dbo.UpdateUserID.UUID,
		}
		or, err = newOrgResponse(ctx, dbtx, o)
		if err != nil {
			return nil, err
		}
		response = append(response, or)
	}

	return response, nil
}

// FindByExternalID is used to find an Org by its External ID
func (fos FindOrgService) FindByExternalID(ctx context.Context, extlID string) (OrgResponse, error) {

	dbtx := fos.Datastorer.Pool()

	o, err := findOrgByExternalID(ctx, dbtx, extlID)
	if err != nil {
		return OrgResponse{}, err
	}

	or, err := newOrgResponse(ctx, dbtx, o)
	if err != nil {
		return OrgResponse{}, err
	}

	return or, nil
}

// findOrgByID retrieves an Org from the datastore given a unique ID
func findOrgByID(ctx context.Context, dbtx DBTX, id uuid.UUID) (org.Org, error) {
	var (
		dbo  orgstore.Org
		extl secure.Identifier
		err  error
	)

	dbo, err = orgstore.New(dbtx).FindOrgByID(ctx, id)
	if err != nil {
		return org.Org{}, errs.E(errs.Database, err)
	}

	extl, err = secure.ParseIdentifier(dbo.OrgExtlID)
	if err != nil {
		return org.Org{}, err
	}

	o := org.Org{
		ID:           dbo.OrgID,
		ExternalID:   extl,
		Name:         dbo.OrgName,
		Description:  dbo.OrgDescription,
		CreateAppID:  dbo.CreateAppID,
		CreateUserID: dbo.CreateUserID.UUID,
		CreateTime:   dbo.CreateTimestamp,
		UpdateAppID:  dbo.UpdateAppID,
		UpdateUserID: dbo.UpdateUserID.UUID,
		UpdateTime:   dbo.UpdateTimestamp,
	}

	return o, nil
}

// findOrgByExternalID retrieves an Org from the datastore given a unique external ID
func findOrgByExternalID(ctx context.Context, dbtx DBTX, extlID string) (org.Org, error) {
	var (
		dbo  orgstore.Org
		extl secure.Identifier
		err  error
	)

	dbo, err = orgstore.New(dbtx).FindOrgByExtlID(ctx, extlID)
	if err != nil {
		return org.Org{}, errs.E(errs.Database, err)
	}

	extl, err = secure.ParseIdentifier(dbo.OrgExtlID)
	if err != nil {
		return org.Org{}, err
	}

	o := org.Org{
		ID:           dbo.OrgID,
		ExternalID:   extl,
		Name:         dbo.OrgName,
		Description:  dbo.OrgDescription,
		CreateAppID:  dbo.CreateAppID,
		CreateUserID: dbo.CreateUserID.UUID,
		CreateTime:   dbo.CreateTimestamp,
		UpdateAppID:  dbo.UpdateAppID,
		UpdateUserID: dbo.UpdateUserID.UUID,
		UpdateTime:   dbo.UpdateTimestamp,
	}

	return o, nil
}
