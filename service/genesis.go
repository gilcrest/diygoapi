package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"

	"github.com/gilcrest/go-api-basic/datastore/orgstore"
	"github.com/gilcrest/go-api-basic/domain/app"
	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/org"
	"github.com/gilcrest/go-api-basic/domain/person"
	"github.com/gilcrest/go-api-basic/domain/secure"
	"github.com/gilcrest/go-api-basic/domain/user"
)

const genesisOrgTypeString string = "genesis"

type FullGenesisResponse struct {
	GenesisResponse GenesisResponse `json:"genesis"`
	TestResponse    TestResponse    `json:"test"`
}

// GenesisResponse is the response struct for the genesis org and app
type GenesisResponse struct {
	OrgResponse OrgResponse `json:"org"`
	AppResponse AppResponse `json:"app"`
}

// TestResponse is the response struct for the test org and app
type TestResponse struct {
	OrgResponse OrgResponse `json:"org"`
	AppResponse AppResponse `json:"app"`
}

// GenesisService seeds the database. It should be run only once on initial database setup.
type GenesisService struct {
	Datastorer            Datastorer
	RandomStringGenerator CryptoRandomGenerator
	EncryptionKey         *[32]byte
}

type seedSet struct {
	org   org.Org
	app   app.App
	user  user.User
	audit audit.SimpleAudit
}

// Seed method seeds the database
func (s GenesisService) Seed(ctx context.Context) (FullGenesisResponse, error) {

	var (
		tx  pgx.Tx
		err error
	)

	// ensure the Genesis seed event has not already taken place
	err = genesisHasOccurred(ctx, s.Datastorer.Pool())
	if err != nil {
		return FullGenesisResponse{}, err
	}

	// start db txn using pgxpool
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return FullGenesisResponse{}, err
	}

	var (
		genesisSet seedSet
		testSet    seedSet
		testKind   org.Kind
	)
	genesisSet, testKind, err = s.seedGenesis(ctx, tx)
	if err != nil {
		return FullGenesisResponse{}, err
	}

	testSet, err = s.seedTest(ctx, tx, testKind)
	if err != nil {
		return FullGenesisResponse{}, err
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return FullGenesisResponse{}, err
	}

	genesisResponse := GenesisResponse{
		OrgResponse: newOrgResponse(orgAudit{Org: genesisSet.org, SimpleAudit: genesisSet.audit}),
		AppResponse: newAppResponse(genesisSet.app),
	}

	testResponse := TestResponse{
		OrgResponse: newOrgResponse(orgAudit{Org: testSet.org, SimpleAudit: testSet.audit}),
		AppResponse: newAppResponse(testSet.app),
	}

	response := FullGenesisResponse{
		GenesisResponse: genesisResponse,
		TestResponse:    testResponse,
	}

	return response, nil
}

func (s GenesisService) seedGenesis(ctx context.Context, tx pgx.Tx) (seedSet, org.Kind, error) {
	var err error

	// create Org
	o := org.Org{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Name:        "genesis",
		Description: "The genesis org represents the first organization created in the database and exists purely for the administrative purpose of creating other organizations, apps and users.",
	}

	// initialize App and inject dependent fields
	a := app.App{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Org:         o,
		Name:        "WOPR",
		Description: "App created as part of Genesis event. To be used solely for creating other apps, orgs and users.",
		APIKeys:     nil,
	}

	keyDeactivation := time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC)
	err = a.AddNewKey(s.RandomStringGenerator, s.EncryptionKey, keyDeactivation)
	if err != nil {
		return seedSet{}, org.Kind{}, errs.E(errs.Internal, s.Datastorer.RollbackTx(ctx, tx, err))
	}

	pgUser, pgAudit := createPeterGabriel(o, a)
	pcUser, pcAudit := createPhilCollins(o, a)

	// create Genesis org kind
	var genesisKindParams orgstore.CreateOrgKindParams
	genesisKindParams, err = createGenesisOrgKind(ctx, s.Datastorer, tx, pgAudit)
	if err != nil {
		return seedSet{}, org.Kind{}, errs.E(errs.Database, s.Datastorer.RollbackTx(ctx, tx, err))
	}
	o.Kind = org.Kind{
		ID:          genesisKindParams.OrgKindID,
		ExternalID:  genesisKindParams.OrgKindExtlID,
		Description: genesisKindParams.OrgKindDesc,
	}

	// create other org kinds (test, standard)
	var testKindParams orgstore.CreateOrgKindParams
	testKindParams, err = createTestOrgKind(ctx, s.Datastorer, tx, pgAudit)
	if err != nil {
		return seedSet{}, org.Kind{}, errs.E(errs.Database, s.Datastorer.RollbackTx(ctx, tx, err))
	}
	tk := org.Kind{
		ID:          testKindParams.OrgKindID,
		ExternalID:  testKindParams.OrgKindExtlID,
		Description: testKindParams.OrgKindDesc,
	}

	err = createStandardOrgKind(ctx, s.Datastorer, tx, pgAudit)
	if err != nil {
		return seedSet{}, org.Kind{}, errs.E(errs.Database, s.Datastorer.RollbackTx(ctx, tx, err))
	}

	sa := audit.SimpleAudit{
		First: pgAudit,
		Last:  pgAudit,
	}

	// write the Org to the database
	err = createOrgDB(ctx, s.Datastorer, tx, orgAudit{Org: o, SimpleAudit: sa})
	if err != nil {
		return seedSet{}, org.Kind{}, err
	}

	// write the App to the database
	err = createAppDB(ctx, s.Datastorer, tx, a, pgAudit)
	if err != nil {
		return seedSet{}, org.Kind{}, err
	}

	// write Peter Gabriel to the database
	err = createUserDB(ctx, s.Datastorer, tx, pgUser, pgAudit)
	if err != nil {
		return seedSet{}, org.Kind{}, err
	}

	// write Phil Collins to the database
	err = createUserDB(ctx, s.Datastorer, tx, pcUser, pcAudit)
	if err != nil {
		return seedSet{}, org.Kind{}, err
	}

	return seedSet{org: o, app: a, user: pgUser, audit: sa}, tk, nil
}

func createPeterGabriel(o org.Org, a app.App) (user.User, audit.Audit) {
	// Peter Gabriel Person
	pgPrsn := person.Person{
		ID:  uuid.New(),
		Org: o,
	}

	// Peter Gabriel Person Profile
	pgPfl := person.Profile{ID: uuid.New(), Person: pgPrsn}
	pgPfl.FirstName = "Peter"
	pgPfl.LastName = "Gabriel"

	// Peter Gabriel User
	pgUser := user.User{
		ID:       uuid.New(),
		Username: strings.TrimSpace("pgabriel"),
		Org:      o,
		Profile:  pgPfl,
	}

	// Peter Gabriel Audit
	pgAudit := audit.Audit{
		App:    a,
		User:   pgUser,
		Moment: time.Now(),
	}

	return pgUser, pgAudit
}

func createPhilCollins(o org.Org, a app.App) (user.User, audit.Audit) {
	// Peter Gabriel Person
	pcPrsn := person.Person{
		ID:  uuid.New(),
		Org: o,
	}

	// Peter Gabriel Person Profile
	pgPfl := person.Profile{ID: uuid.New(), Person: pcPrsn}
	pgPfl.FirstName = "Phil"
	pgPfl.LastName = "Collins"

	// Peter Gabriel User
	pcUser := user.User{
		ID:       uuid.New(),
		Username: strings.TrimSpace("pcollins"),
		Org:      o,
		Profile:  pgPfl,
	}

	// Peter Gabriel Audit
	pcAudit := audit.Audit{
		App:    a,
		User:   pcUser,
		Moment: time.Now(),
	}

	return pcUser, pcAudit
}

func (s GenesisService) seedTest(ctx context.Context, tx pgx.Tx, k org.Kind) (seedSet, error) {
	var err error

	// create Org
	o := org.Org{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Name:        "test",
		Description: "The test org is self explanatory",
		Kind:        k,
	}

	// initialize App and inject dependent fields
	a := app.App{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Org:         o,
		Name:        "test",
		Description: "The test app is self explanatory",
		APIKeys:     nil,
	}

	keyDeactivation := time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC)
	err = a.AddNewKey(s.RandomStringGenerator, s.EncryptionKey, keyDeactivation)
	if err != nil {
		return seedSet{}, errs.E(errs.Internal, s.Datastorer.RollbackTx(ctx, tx, err))
	}

	// create Person
	prsn := person.Person{
		ID:  uuid.New(),
		Org: o,
	}

	// create Person Profile
	pfl := person.Profile{ID: uuid.New(), Person: prsn}
	pfl.FirstName = "Steve"
	pfl.LastName = "Hackett"

	// create User
	u := user.User{
		ID:       uuid.New(),
		Username: strings.TrimSpace("shackett"),
		Org:      o,
		Profile:  pfl,
	}

	//create Audit
	adt := audit.Audit{
		App:    a,
		User:   u,
		Moment: time.Now(),
	}

	sa := audit.SimpleAudit{
		First: adt,
		Last:  adt,
	}

	// write the Org to the database
	err = createOrgDB(ctx, s.Datastorer, tx, orgAudit{Org: o, SimpleAudit: sa})
	if err != nil {
		return seedSet{}, err
	}

	// write the App to the database
	err = createAppDB(ctx, s.Datastorer, tx, a, adt)
	if err != nil {
		return seedSet{}, err
	}

	// write the User to the database
	err = createUserDB(ctx, s.Datastorer, tx, u, adt)
	if err != nil {
		return seedSet{}, err
	}

	return seedSet{org: o, app: a, user: u, audit: sa}, nil
}

func genesisHasOccurred(ctx context.Context, dbtx orgstore.DBTX) (err error) {
	var (
		existingOrgs         []orgstore.FindOrgsByKindExtlIDRow
		hasGenesisOrgTypeRow = true
		hasGenesisOrgRow     = true
	)

	// validate Genesis records do not exist already
	// first: check org_type
	_, err = orgstore.New(dbtx).FindOrgKindByExtlID(ctx, genesisOrgTypeString)
	if err != nil {
		if err != pgx.ErrNoRows {
			return errs.E(errs.Database, err)
		}
		hasGenesisOrgTypeRow = false
	}

	// last: check org
	existingOrgs, err = orgstore.New(dbtx).FindOrgsByKindExtlID(ctx, genesisOrgTypeString)
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
