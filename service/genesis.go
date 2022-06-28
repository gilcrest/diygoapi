package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

const (
	// PrincipalOrgName is the first organization created as part of
	// the Genesis event and is the central administration org.
	PrincipalOrgName        = "Principal"
	principalOrgDescription = "The Principal org represents the first organization created in the database and exists for the administrative purpose of creating other organizations, apps and users."
	// PrincipalAppName is the first app created as part of the
	// Genesis event and is the central administration app.
	PrincipalAppName        = "Developer Dashboard"
	principalAppDescription = "App created as part of Genesis event. To be used solely for creating other apps, orgs and users."
	// PrincipalTestUsername is for the test user created as part of the
	// Genesis event and is needed for testing some features of the Principal org.
	PrincipalTestUsername      = "pgabriel"
	principalTestUserFirstName = "Peter"
	principalTestUserLastName  = "Gabriel"
	// TestOrgName is the organization created as part of the Genesis
	// event solely for the purpose of testing
	TestOrgName        = "Test Org"
	testOrgDescription = "The test org is used solely for the purpose of testing."
	// TestAppName is the test app created as part of the Genesis
	// event solely for the purpose of testing
	TestAppName        = "Test App"
	testAppDescription = "The test app is used solely for the purpose of testing."
	// TestUsername is the test user created as part of the Genesis
	// event solely for the purpose of testing
	TestUsername      = "shackett"
	testUserFirstName = "Steve"
	testUserLastName  = "Hackett"

	genesisOrgKind string = "genesis"
	// LocalJSONGenesisResponseFile is the local JSON Genesis Response File path
	// (relative to project root)
	LocalJSONGenesisResponseFile = "./config/genesis/response.json"
)

// GenesisResponse contains both the Genesis response and the Test response
type GenesisResponse struct {
	Principal     OrgResponse  `json:"principal"`
	Test          OrgResponse  `json:"test"`
	UserInitiated *OrgResponse `json:"userInitiated,omitempty"`
}

// GenesisRequest is the request struct for the genesis service
type GenesisRequest struct {
	User struct {
		// Email: The Genesis user email address.
		Email string `json:"email"`

		// FirstName: The Genesis user first name.
		FirstName string `json:"first_name"`

		// LastName: The Genesis user last name.
		LastName string `json:"last_name"`
	} `json:"user"`

	UserInitiatedOrg CreateOrgRequest `json:"org"`

	// Permissions: The list of permissions to be created as part of Genesis
	Permissions []PermissionRequest `json:"permissions"`

	// Roles: The list of Roles to be created as part of Genesis
	Roles []CreateRoleRequest `json:"roles"`
}

// seedGenesisReturnParams returns several structs needed for subsequent actions
// in Genesis.
type seedGenesisReturnParams struct {
	org          org.Org
	app          app.App
	standardKind org.Kind
	testKind     org.Kind
	audit        audit.Audit
}

// seedGenesisReturnParams returns several structs needed for subsequent actions
// in Genesis.
type seedTestReturnParams struct {
	org      org.Org
	app      app.App
	testUser user.User
	audit    audit.Audit
}

// GenesisService seeds the database. It should be run only once on initial database setup.
type GenesisService struct {
	Datastorer            Datastorer
	RandomStringGenerator CryptoRandomGenerator
	EncryptionKey         *[32]byte
}

// Seed method seeds the database
func (s GenesisService) Seed(ctx context.Context, r *GenesisRequest) (gr GenesisResponse, err error) {

	// ensure the Genesis seed event has not already taken place
	err = genesisHasOccurred(ctx, s.Datastorer.Pool())
	if err != nil {
		return gr, err
	}

	var (
		sgrp seedGenesisReturnParams
		strp seedTestReturnParams
	)

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return gr, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	// seed Genesis org data. As part of this method, the initial org.Kind
	// structs are added to the db. The test kind is returned for use
	// in the seedTest method
	sgrp, err = s.seedGenesis(ctx, tx, r)
	if err != nil {
		return gr, err
	}

	// seed Test org data.
	strp, err = s.seedTest(ctx, tx, sgrp)
	if err != nil {
		return gr, err
	}

	// seed User Initiated org data.
	var uirp seedUserInitiatedDataReturnParams
	uirp, err = s.seedUserInitiatedData(ctx, tx, sgrp, r)
	if err != nil {
		return gr, err
	}

	// seed Permissions
	err = seedPermissions(ctx, tx, r, sgrp.audit)
	if err != nil {
		return gr, err
	}

	// seed Roles
	for _, crr := range r.Roles {
		// add 3 users created as part of Genesis to the roles created
		crr.UserExternals = append(crr.UserExternals, sgrp.audit.User.ExternalID.String(), strp.testUser.ExternalID.String(), uirp.SeedUser.ExternalID.String())
		_, err = createRoleTx(ctx, tx, &crr, sgrp.audit)
		if err != nil {
			return gr, err
		}
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return gr, err
	}

	pOrg := newOrgResponse(orgAudit{Org: sgrp.org, SimpleAudit: audit.SimpleAudit{First: sgrp.audit, Last: sgrp.audit}},
		appAudit{App: sgrp.app, SimpleAudit: audit.SimpleAudit{First: sgrp.audit, Last: sgrp.audit}})

	tOrg := newOrgResponse(orgAudit{Org: strp.org, SimpleAudit: audit.SimpleAudit{First: strp.audit, Last: strp.audit}},
		appAudit{App: strp.app, SimpleAudit: audit.SimpleAudit{First: strp.audit, Last: strp.audit}})

	response := GenesisResponse{
		Principal:     pOrg,
		Test:          tOrg,
		UserInitiated: &uirp.OrgResponse,
	}

	return response, nil
}

func (s GenesisService) seedGenesis(ctx context.Context, tx pgx.Tx, r *GenesisRequest) (seedGenesisReturnParams, error) {
	var err error

	// create Org
	o := org.Org{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Name:        PrincipalOrgName,
		Description: principalOrgDescription,
	}

	// initialize App and inject dependent fields

	a := app.App{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Org:         o,
		Name:        PrincipalAppName,
		Description: principalAppDescription,
		APIKeys:     nil,
	}

	// create API key
	keyDeactivation := time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC)
	err = a.AddNewKey(s.RandomStringGenerator, s.EncryptionKey, keyDeactivation)
	if err != nil {
		return seedGenesisReturnParams{}, errs.E(errs.Internal, err)
	}

	// initialize Peter Gabriel test user in Genesis org
	pgUser := user.User{
		ID:         uuid.New(),
		ExternalID: secure.NewID(),
		Username:   strings.TrimSpace(PrincipalTestUsername),
		Org:        o,
		Profile: person.Profile{
			ID:        uuid.New(),
			Person:    person.Person{ID: uuid.New(), Org: o},
			FirstName: principalTestUserFirstName,
			LastName:  principalTestUserLastName,
		},
	}

	// initialize Genesis user from request data
	gUser := user.User{
		ID:         uuid.New(),
		ExternalID: secure.NewID(),
		Username:   strings.TrimSpace(r.User.Email),
		Org:        o,
		Profile: person.Profile{
			ID:        uuid.New(),
			Person:    person.Person{ID: uuid.New(), Org: o},
			FirstName: strings.TrimSpace(r.User.FirstName),
			LastName:  strings.TrimSpace(r.User.LastName),
		},
	}

	err = gUser.IsValid()
	if err != nil {
		return seedGenesisReturnParams{}, err
	}

	adt := audit.Audit{
		App:    a,
		User:   gUser,
		Moment: time.Now(),
	}

	// create Genesis org kind
	var genesisKindParams orgstore.CreateOrgKindParams
	genesisKindParams, err = createGenesisOrgKind(ctx, tx, adt)
	if err != nil {
		return seedGenesisReturnParams{}, errs.E(errs.Database, err)
	}
	o.Kind = org.Kind{
		ID:          genesisKindParams.OrgKindID,
		ExternalID:  genesisKindParams.OrgKindExtlID,
		Description: genesisKindParams.OrgKindDesc,
	}

	// create other org kinds (test, standard)
	var testKindParams orgstore.CreateOrgKindParams
	testKindParams, err = createTestOrgKind(ctx, tx, adt)
	if err != nil {
		return seedGenesisReturnParams{}, errs.E(errs.Database, err)
	}
	tk := org.Kind{
		ID:          testKindParams.OrgKindID,
		ExternalID:  testKindParams.OrgKindExtlID,
		Description: testKindParams.OrgKindDesc,
	}

	var standardOrgParams orgstore.CreateOrgKindParams
	standardOrgParams, err = createStandardOrgKind(ctx, tx, adt)
	if err != nil {
		return seedGenesisReturnParams{}, errs.E(errs.Database, err)
	}
	sk := org.Kind{
		ID:          standardOrgParams.OrgKindID,
		ExternalID:  standardOrgParams.OrgKindExtlID,
		Description: standardOrgParams.OrgKindDesc,
	}

	sa := audit.SimpleAudit{
		First: adt,
		Last:  adt,
	}

	// write the Org to the database
	err = createOrgTx(ctx, tx, orgAudit{Org: o, SimpleAudit: sa})
	if err != nil {
		return seedGenesisReturnParams{}, err
	}

	createAppParams := appstore.CreateAppParams{
		AppID:           a.ID,
		OrgID:           a.Org.ID,
		AppExtlID:       a.ExternalID.String(),
		AppName:         a.Name,
		AppDescription:  a.Description,
		CreateAppID:     adt.App.ID,
		CreateUserID:    adt.User.NullUUID(),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    adt.User.NullUUID(),
		UpdateTimestamp: adt.Moment,
	}

	// create app database record using appstore
	var rowsAffected int64
	rowsAffected, err = appstore.New(tx).CreateApp(ctx, createAppParams)
	if err != nil {
		return seedGenesisReturnParams{}, errs.E(errs.Database, err)
	}

	if rowsAffected != 1 {
		return seedGenesisReturnParams{}, errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	for _, key := range a.APIKeys {

		createAppAPIKeyParams := appstore.CreateAppAPIKeyParams{
			ApiKey:          key.Ciphertext(),
			AppID:           a.ID,
			DeactvDate:      key.DeactivationDate(),
			CreateAppID:     adt.App.ID,
			CreateUserID:    adt.User.NullUUID(),
			CreateTimestamp: adt.Moment,
			UpdateAppID:     adt.App.ID,
			UpdateUserID:    adt.User.NullUUID(),
			UpdateTimestamp: adt.Moment,
		}

		// create app API key database record using appstore
		var apiKeyRowsAffected int64
		apiKeyRowsAffected, err = appstore.New(tx).CreateAppAPIKey(ctx, createAppAPIKeyParams)
		if err != nil {
			return seedGenesisReturnParams{}, errs.E(errs.Database, err)
		}

		if apiKeyRowsAffected != 1 {
			return seedGenesisReturnParams{}, errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", apiKeyRowsAffected))
		}
	}

	// write user from request to the database
	err = createUserTx(ctx, tx, gUser, adt)
	if err != nil {
		return seedGenesisReturnParams{}, err
	}

	// write user to the database
	err = createUserTx(ctx, tx, pgUser, adt)
	if err != nil {
		return seedGenesisReturnParams{}, err
	}

	sgrp := seedGenesisReturnParams{
		org:          o,
		app:          a,
		standardKind: sk,
		testKind:     tk,
		audit:        adt,
	}

	return sgrp, nil
}

func (s GenesisService) seedTest(ctx context.Context, tx pgx.Tx, sgrp seedGenesisReturnParams) (seedTestReturnParams, error) {
	var err error

	// create Org
	o := org.Org{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Name:        TestOrgName,
		Description: testOrgDescription,
		Kind:        sgrp.testKind,
	}

	// initialize App and inject dependent fields
	a := app.App{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Org:         o,
		Name:        TestAppName,
		Description: testAppDescription,
		APIKeys:     nil,
	}

	keyDeactivation := time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC)
	err = a.AddNewKey(s.RandomStringGenerator, s.EncryptionKey, keyDeactivation)
	if err != nil {
		return seedTestReturnParams{}, errs.E(errs.Internal, err)
	}

	// create Person
	prsn := person.Person{
		ID:  uuid.New(),
		Org: o,
	}

	// create Person Profile
	pfl := person.Profile{ID: uuid.New(), Person: prsn}
	pfl.FirstName = testUserFirstName
	pfl.LastName = testUserLastName

	// create User
	u := user.User{
		ID:         uuid.New(),
		ExternalID: secure.NewID(),
		Username:   TestUsername,
		Org:        o,
		Profile:    pfl,
	}

	sa := audit.SimpleAudit{
		First: sgrp.audit,
		Last:  sgrp.audit,
	}

	// write the Org to the database
	err = createOrgTx(ctx, tx, orgAudit{Org: o, SimpleAudit: sa})
	if err != nil {
		return seedTestReturnParams{}, err
	}

	createAppParams := appstore.CreateAppParams{
		AppID:           a.ID,
		OrgID:           a.Org.ID,
		AppExtlID:       a.ExternalID.String(),
		AppName:         a.Name,
		AppDescription:  a.Description,
		CreateAppID:     sgrp.audit.App.ID,
		CreateUserID:    sgrp.audit.User.NullUUID(),
		CreateTimestamp: sgrp.audit.Moment,
		UpdateAppID:     sgrp.audit.App.ID,
		UpdateUserID:    sgrp.audit.User.NullUUID(),
		UpdateTimestamp: sgrp.audit.Moment,
	}

	// create app database record using appstore
	var rowsAffected int64
	rowsAffected, err = appstore.New(tx).CreateApp(ctx, createAppParams)
	if err != nil {
		return seedTestReturnParams{}, errs.E(errs.Database, err)
	}

	if rowsAffected != 1 {
		return seedTestReturnParams{}, errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	for _, key := range a.APIKeys {

		createAppAPIKeyParams := appstore.CreateAppAPIKeyParams{
			ApiKey:          key.Ciphertext(),
			AppID:           a.ID,
			DeactvDate:      key.DeactivationDate(),
			CreateAppID:     sgrp.audit.App.ID,
			CreateUserID:    sgrp.audit.User.NullUUID(),
			CreateTimestamp: sgrp.audit.Moment,
			UpdateAppID:     sgrp.audit.App.ID,
			UpdateUserID:    sgrp.audit.User.NullUUID(),
			UpdateTimestamp: sgrp.audit.Moment,
		}

		// create app API key database record using appstore
		var apiKeyRowsAffected int64
		apiKeyRowsAffected, err = appstore.New(tx).CreateAppAPIKey(ctx, createAppAPIKeyParams)
		if err != nil {
			return seedTestReturnParams{}, errs.E(errs.Database, err)
		}

		if apiKeyRowsAffected != 1 {
			return seedTestReturnParams{}, errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", apiKeyRowsAffected))
		}
	}

	// write the User to the database
	err = createUserTx(ctx, tx, u, sgrp.audit)
	if err != nil {
		return seedTestReturnParams{}, err
	}

	strp := seedTestReturnParams{
		org:      o,
		app:      a,
		testUser: u,
		audit:    sgrp.audit,
	}

	return strp, nil
}

type seedUserInitiatedDataReturnParams struct {
	OrgResponse OrgResponse
	SeedUser    user.User
}

func (s GenesisService) seedUserInitiatedData(ctx context.Context, tx pgx.Tx, sgrp seedGenesisReturnParams, r *GenesisRequest) (seedUserInitiatedDataReturnParams, error) {
	var err error

	// create Org
	o := org.Org{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Name:        r.UserInitiatedOrg.Name,
		Description: r.UserInitiatedOrg.Description,
		Kind:        sgrp.standardKind,
	}

	// initialize App and inject dependent fields
	a := app.App{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Org:         o,
		Name:        r.UserInitiatedOrg.App.Name,
		Description: r.UserInitiatedOrg.App.Description,
		APIKeys:     nil,
	}

	keyDeactivation := time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC)
	err = a.AddNewKey(s.RandomStringGenerator, s.EncryptionKey, keyDeactivation)
	if err != nil {
		return seedUserInitiatedDataReturnParams{}, errs.E(errs.Internal, err)
	}

	// initialize Genesis user from request data
	gUser := user.User{
		ID:         uuid.New(),
		ExternalID: secure.NewID(),
		Username:   strings.TrimSpace(r.User.Email),
		Org:        o,
		Profile: person.Profile{
			ID:        uuid.New(),
			Person:    person.Person{ID: uuid.New(), Org: o},
			FirstName: strings.TrimSpace(r.User.FirstName),
			LastName:  strings.TrimSpace(r.User.LastName),
		},
	}

	sa := audit.SimpleAudit{
		First: sgrp.audit,
		Last:  sgrp.audit,
	}

	// write the Org to the database
	err = createOrgTx(ctx, tx, orgAudit{Org: o, SimpleAudit: sa})
	if err != nil {
		return seedUserInitiatedDataReturnParams{}, err
	}

	createAppParams := appstore.CreateAppParams{
		AppID:           a.ID,
		OrgID:           a.Org.ID,
		AppExtlID:       a.ExternalID.String(),
		AppName:         a.Name,
		AppDescription:  a.Description,
		CreateAppID:     sgrp.audit.App.ID,
		CreateUserID:    sgrp.audit.User.NullUUID(),
		CreateTimestamp: sgrp.audit.Moment,
		UpdateAppID:     sgrp.audit.App.ID,
		UpdateUserID:    sgrp.audit.User.NullUUID(),
		UpdateTimestamp: sgrp.audit.Moment,
	}

	// create app database record using appstore
	var rowsAffected int64
	rowsAffected, err = appstore.New(tx).CreateApp(ctx, createAppParams)
	if err != nil {
		return seedUserInitiatedDataReturnParams{}, errs.E(errs.Database, err)
	}

	if rowsAffected != 1 {
		return seedUserInitiatedDataReturnParams{}, errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	for _, key := range a.APIKeys {

		createAppAPIKeyParams := appstore.CreateAppAPIKeyParams{
			ApiKey:          key.Ciphertext(),
			AppID:           a.ID,
			DeactvDate:      key.DeactivationDate(),
			CreateAppID:     sgrp.audit.App.ID,
			CreateUserID:    sgrp.audit.User.NullUUID(),
			CreateTimestamp: sgrp.audit.Moment,
			UpdateAppID:     sgrp.audit.App.ID,
			UpdateUserID:    sgrp.audit.User.NullUUID(),
			UpdateTimestamp: sgrp.audit.Moment,
		}

		// create app API key database record using appstore
		var apiKeyRowsAffected int64
		apiKeyRowsAffected, err = appstore.New(tx).CreateAppAPIKey(ctx, createAppAPIKeyParams)
		if err != nil {
			return seedUserInitiatedDataReturnParams{}, errs.E(errs.Database, err)
		}

		if apiKeyRowsAffected != 1 {
			return seedUserInitiatedDataReturnParams{}, errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", apiKeyRowsAffected))
		}
	}

	// write the User to the database
	err = createUserTx(ctx, tx, gUser, sgrp.audit)
	if err != nil {
		return seedUserInitiatedDataReturnParams{}, err
	}

	oa := orgAudit{Org: o, SimpleAudit: sa}
	aa := appAudit{App: a, SimpleAudit: sa}

	rp := seedUserInitiatedDataReturnParams{
		OrgResponse: newOrgResponse(oa, aa),
		SeedUser:    gUser,
	}

	return rp, nil
}

func genesisHasOccurred(ctx context.Context, dbtx orgstore.DBTX) (err error) {
	var (
		existingOrgs         []orgstore.FindOrgsByKindExtlIDRow
		hasGenesisOrgTypeRow = true
		hasGenesisOrgRow     = true
	)

	// validate Genesis records do not exist already
	// first: check org_type
	_, err = orgstore.New(dbtx).FindOrgKindByExtlID(ctx, genesisOrgKind)
	if err != nil {
		if err != pgx.ErrNoRows {
			return errs.E(errs.Database, err)
		}
		hasGenesisOrgTypeRow = false
	}

	// last: check org
	existingOrgs, err = orgstore.New(dbtx).FindOrgsByKindExtlID(ctx, genesisOrgKind)
	if err != nil {
		return errs.E(errs.Database, err)
	}
	if len(existingOrgs) == 0 {
		hasGenesisOrgRow = false
	}

	if hasGenesisOrgTypeRow || hasGenesisOrgRow {
		return errs.E(errs.Validation, "No prior data should exist when executing Genesis Service")
	}

	return nil
}

// ReadConfig reads the generated config file from Genesis
// and returns it in the response body
func (s GenesisService) ReadConfig() (gr GenesisResponse, err error) {
	var b []byte
	b, err = os.ReadFile(LocalJSONGenesisResponseFile)
	if err != nil {
		return GenesisResponse{}, errs.E(err)
	}
	err = json.Unmarshal(b, &gr)
	if err != nil {
		return GenesisResponse{}, errs.E(err)
	}

	return gr, nil
}

func seedPermissions(ctx context.Context, tx pgx.Tx, r *GenesisRequest, adt audit.Audit) (err error) {
	for _, p := range r.Permissions {
		_, err = createPermissionTx(ctx, tx, &p, adt)
		if err != nil {
			return err
		}
	}

	return nil
}
