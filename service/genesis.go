package service

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"golang.org/x/oauth2"
	"golang.org/x/text/language"

	"github.com/gilcrest/diygoapi"
	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/secure"
	"github.com/gilcrest/diygoapi/sqldb/datastore"
)

const (
	// PrincipalOrgName is the first organization created as part of
	// the Genesis event and is the central administration org.
	PrincipalOrgName               = "Principal"
	principalOrgDescription        = "The Principal org represents the first organization created in the database and exists for the administrative purpose of creating other organizations, apps and users."
	principalOrgKind        string = "principal"
	// PrincipalAppName is the first app created as part of the
	// Genesis event and is the central administration app.
	PrincipalAppName        = "Developer Dashboard"
	principalAppDescription = "App created as part of Genesis event. To be used solely for creating other apps, orgs and users."
	// Peter Gabriel is the test user created as part of the Genesis event.
	testUserFirstName = "Peter"
	testUserLastName  = "Gabriel"
	// TestOrgName is the organization created as part of the Genesis
	// event solely for the purpose of testing
	TestOrgName        = "Test Org"
	testOrgDescription = "The test org is used solely for the purpose of testing."
	// TestAppName is the test app created as part of the Genesis
	// event solely for the purpose of testing
	TestAppName        = "Test App"
	testAppDescription = "The test app is used solely for the purpose of testing."
	// TestRoleCode is the role created to flag the test account in the test org.
	TestRoleCode = "TestAdminRole"
	// LocalJSONGenesisResponseFile is the local JSON Genesis Response File path
	// (relative to project root)
	LocalJSONGenesisResponseFile = "./config/genesis/response.json"
)

// principalSeed is used in the Genesis event to seed the principal
// org, app, initial org Kinds, and Audit object from the event.
type principalSeed struct {
	PrincipalOrg    *diygoapi.Org
	PrincipalApp    *diygoapi.App
	StandardOrgKind *diygoapi.OrgKind
	TestOrgKind     *diygoapi.OrgKind
	Audit           diygoapi.Audit
}

// testSeed is the Test Org, App and User
type testSeed struct {
	TestOrg  *diygoapi.Org
	TestApp  *diygoapi.App
	TestUser *diygoapi.User
}

// userInitiatedSeed is the User Initiated Org, App and User
type userInitiatedSeed struct {
	UserInitiatedOrg *diygoapi.Org
	UserInitiatedApp *diygoapi.App
}

// GenesisService seeds the database. It should be run only once on initial database setup.
type GenesisService struct {
	Datastorer      diygoapi.Datastorer
	APIKeyGenerator diygoapi.APIKeyGenerator
	EncryptionKey   *[32]byte
	TokenExchanger  diygoapi.TokenExchanger
	LanguageMatcher language.Matcher
}

// Arche creates the initial seed data in the database.
func (s *GenesisService) Arche(ctx context.Context, r *diygoapi.GenesisRequest) (gr diygoapi.GenesisResponse, err error) {
	const op errs.Op = "service/GenesisService.Arche"

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return gr, errs.E(op, err)
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	// ensure the Genesis event has not already taken place
	err = genesisHasOccurred(ctx, tx)
	if err != nil {
		return gr, errs.E(op, err)
	}

	var (
		ps  principalSeed
		ts  testSeed
		uis userInitiatedSeed
	)

	// seed Principal org, app and user data as well as initial OrgKind structs.
	// principalSeed struct is returned for use in subsequent steps
	ps, err = s.seedPrincipal(ctx, tx, r)
	if err != nil {
		return gr, errs.E(op, err)
	}

	// seed Test org data.
	ts, err = s.seedTest(ctx, tx, ps)
	if err != nil {
		return gr, errs.E(op, err)
	}

	// seed User Initiated org data.
	uis, err = s.seedUserInitiatedData(ctx, tx, ps, r)
	if err != nil {
		return gr, errs.E(op, err)
	}

	// seed Permissions
	err = seedPermissions(ctx, tx, r, ps.Audit)
	if err != nil {
		return gr, errs.E(op, err)
	}

	// seed Roles
	var gRoles genesisRoles
	gRoles, err = seedRoles(ctx, tx, r, ps.Audit)
	if err != nil {
		return gr, errs.E(op, err)
	}

	gr2gup := grantRoles2GenesisUsersParams{
		Roles:             gRoles,
		PrincipalSeed:     ps,
		TestSeed:          ts,
		UserInitiatedSeed: uis,
	}

	// grant Roles to users
	err = grantRoles2GenesisUsers(ctx, tx, gr2gup)
	if err != nil {
		return gr, errs.E(op, err)
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return gr, errs.E(op, err)
	}

	pOrg := newOrgResponse(&orgAudit{Org: ps.PrincipalOrg, SimpleAudit: &diygoapi.SimpleAudit{Create: ps.Audit, Update: ps.Audit}},
		appAudit{App: ps.PrincipalApp, SimpleAudit: &diygoapi.SimpleAudit{Create: ps.Audit, Update: ps.Audit}})

	tOrg := newOrgResponse(&orgAudit{Org: ts.TestOrg, SimpleAudit: &diygoapi.SimpleAudit{Create: ps.Audit, Update: ps.Audit}},
		appAudit{App: ts.TestApp, SimpleAudit: &diygoapi.SimpleAudit{Create: ps.Audit, Update: ps.Audit}})

	uiOrg := newOrgResponse(&orgAudit{Org: uis.UserInitiatedOrg, SimpleAudit: &diygoapi.SimpleAudit{Create: ps.Audit, Update: ps.Audit}},
		appAudit{App: uis.UserInitiatedApp, SimpleAudit: &diygoapi.SimpleAudit{Create: ps.Audit, Update: ps.Audit}})

	response := diygoapi.GenesisResponse{
		Principal:     pOrg,
		Test:          tOrg,
		UserInitiated: uiOrg,
	}

	return response, nil
}

func (s *GenesisService) seedPrincipal(ctx context.Context, tx pgx.Tx, r *diygoapi.GenesisRequest) (principalSeed, error) {
	const (
		seedPrincipalRealm string  = "seedPrincipal"
		op                 errs.Op = "service/GenesisService.seedPrincipal"
	)

	var (
		provider   diygoapi.Provider
		token      *oauth2.Token
		authParams *diygoapi.AuthenticationParams
		realm      string
		err        error
	)

	authParams, _ = diygoapi.AuthParamsFromContext(ctx)
	if authParams != nil {
		provider = authParams.Provider
		token = authParams.Token
		realm = authParams.Realm
	} else {
		provider = diygoapi.ParseProvider(r.User.Provider)
		token = &oauth2.Token{AccessToken: r.User.Token, TokenType: diygoapi.BearerTokenType}
		realm = seedPrincipalRealm
	}

	// create Org
	o := &diygoapi.Org{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Name:        PrincipalOrgName,
		Description: principalOrgDescription,
	}

	// initialize App and inject dependent fields
	nap := newAppParams{
		Name:            PrincipalAppName,
		Description:     principalAppDescription,
		Org:             o,
		ApiKeyGenerator: s.APIKeyGenerator,
		EncryptionKey:   s.EncryptionKey,
	}

	var a *diygoapi.App
	a, err = newApp(nap)
	if err != nil {
		return principalSeed{}, errs.E(op, err)
	}

	// initialize "The Creator" user from request data
	// auth could not be found by access token in the db
	// get ProviderInfo from provider API
	var providerInfo *diygoapi.ProviderInfo
	providerInfo, err = s.TokenExchanger.Exchange(ctx, realm, provider, token)
	if err != nil {
		return principalSeed{}, errs.E(op, err)
	}

	gUser := diygoapi.NewUserFromProviderInfo(providerInfo, s.LanguageMatcher)

	err = gUser.Validate()
	if err != nil {
		return principalSeed{}, errs.E(op, err)
	}

	gPerson := diygoapi.Person{
		ID:         uuid.New(),
		ExternalID: secure.NewID(),
		Users:      []*diygoapi.User{gUser},
	}

	adt := diygoapi.Audit{
		App:    a,
		User:   gUser,
		Moment: time.Now(),
	}

	cg := datastore.CreateAuthProviderParams{
		AuthProviderID:   int64(diygoapi.Google),
		AuthProviderCd:   diygoapi.Google.String(),
		AuthProviderDesc: "Google Oauth2",
		CreateAppID:      adt.App.ID,
		CreateUserID:     adt.User.NullUUID(),
		CreateTimestamp:  adt.Moment,
		UpdateAppID:      adt.App.ID,
		UpdateUserID:     adt.User.NullUUID(),
		UpdateTimestamp:  adt.Moment,
	}
	_, err = datastore.New(tx).CreateAuthProvider(ctx, cg)
	if err != nil {
		return principalSeed{}, errs.E(op, errs.Database, err)
	}

	// write Person/User from request to the database
	err = createPersonTx(ctx, tx, gPerson, adt)
	if err != nil {
		return principalSeed{}, errs.E(op, err)
	}

	// associate Genesis org to "The Creator"
	aoaParams := attachOrgAssociationParams{
		Org:   o,
		User:  gUser,
		Audit: adt,
	}
	err = attachOrgAssociation(ctx, tx, aoaParams)
	if err != nil {
		return principalSeed{}, errs.E(op, err)
	}

	// create Auth for "The Creator"
	auth := diygoapi.Auth{
		ID:                        uuid.New(),
		User:                      gUser,
		Provider:                  providerInfo.Provider,
		ProviderClientID:          providerInfo.TokenInfo.ClientID,
		ProviderPersonID:          providerInfo.UserInfo.ExternalID,
		ProviderAccessToken:       providerInfo.TokenInfo.Token.AccessToken,
		ProviderAccessTokenExpiry: providerInfo.TokenInfo.Token.Expiry,
		ProviderRefreshToken:      providerInfo.TokenInfo.Token.RefreshToken,
	}

	err = createAuthTx(ctx, tx, createAuthTxParams{Auth: auth, Audit: adt})
	if err != nil {
		return principalSeed{}, errs.E(op, err)
	}

	// create Principal org kind
	var principalKindParams datastore.CreateOrgKindParams
	principalKindParams, err = createPrincipalOrgKind(ctx, tx, adt)
	if err != nil {
		return principalSeed{}, errs.E(op, errs.Database, err)
	}
	o.Kind = &diygoapi.OrgKind{
		ID:          principalKindParams.OrgKindID,
		ExternalID:  principalKindParams.OrgKindExtlID,
		Description: principalKindParams.OrgKindDesc,
	}

	// create other org kinds (test, standard)
	var testKindParams datastore.CreateOrgKindParams
	testKindParams, err = createTestOrgKind(ctx, tx, adt)
	if err != nil {
		return principalSeed{}, errs.E(op, errs.Database, err)
	}
	tk := &diygoapi.OrgKind{
		ID:          testKindParams.OrgKindID,
		ExternalID:  testKindParams.OrgKindExtlID,
		Description: testKindParams.OrgKindDesc,
	}

	var standardOrgParams datastore.CreateOrgKindParams
	standardOrgParams, err = createStandardOrgKind(ctx, tx, adt)
	if err != nil {
		return principalSeed{}, errs.E(op, errs.Database, err)
	}
	sk := &diygoapi.OrgKind{
		ID:          standardOrgParams.OrgKindID,
		ExternalID:  standardOrgParams.OrgKindExtlID,
		Description: standardOrgParams.OrgKindDesc,
	}

	sa := &diygoapi.SimpleAudit{
		Create: adt,
		Update: adt,
	}

	// write the Org to the database
	err = createOrgTx(ctx, tx, &orgAudit{Org: o, SimpleAudit: sa})
	if err != nil {
		return principalSeed{}, errs.E(op, err)
	}

	// app is also to be created, write it to the db
	err = createAppTx(ctx, tx, appAudit{App: a, SimpleAudit: sa})
	if err != nil {
		return principalSeed{}, errs.E(op, err)
	}

	seed := principalSeed{
		PrincipalOrg:    o,
		PrincipalApp:    a,
		StandardOrgKind: sk,
		TestOrgKind:     tk,
		Audit:           adt,
	}

	return seed, nil
}

func (s *GenesisService) seedTest(ctx context.Context, tx pgx.Tx, ps principalSeed) (testSeed, error) {
	const op errs.Op = "service/GenesisService.seedTest"

	var err error

	// initialize test user in Genesis org
	testUser := &diygoapi.User{
		ID:         uuid.New(),
		ExternalID: secure.NewID(),
		FirstName:  testUserFirstName,
		LastName:   testUserLastName,
	}

	// make Test User a whole Person
	testPerson := diygoapi.Person{
		ID:         uuid.New(),
		ExternalID: secure.NewID(),
		Users:      []*diygoapi.User{testUser},
	}

	// write Test Person/User to the database
	err = createPersonTx(ctx, tx, testPerson, ps.Audit)
	if err != nil {
		return testSeed{}, errs.E(op, err)
	}

	// associate Principal org to the Test User
	aoaParams := attachOrgAssociationParams{
		Org:   ps.PrincipalOrg,
		User:  testUser,
		Audit: ps.Audit,
	}
	err = attachOrgAssociation(ctx, tx, aoaParams)
	if err != nil {
		return testSeed{}, errs.E(op, err)
	}

	// create Org
	o := &diygoapi.Org{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Name:        TestOrgName,
		Description: testOrgDescription,
		Kind:        ps.TestOrgKind,
	}

	nap := newAppParams{
		Name:            TestAppName,
		Description:     testAppDescription,
		Org:             o,
		ApiKeyGenerator: s.APIKeyGenerator,
		EncryptionKey:   s.EncryptionKey,
	}

	var a *diygoapi.App
	a, err = newApp(nap)
	if err != nil {
		return testSeed{}, errs.E(op, errs.Internal, err)
	}

	sa := &diygoapi.SimpleAudit{
		Create: ps.Audit,
		Update: ps.Audit,
	}

	// write the Org to the database
	err = createOrgTx(ctx, tx, &orgAudit{Org: o, SimpleAudit: sa})
	if err != nil {
		return testSeed{}, errs.E(op, err)
	}

	// app is also to be created, write it to the db
	err = createAppTx(ctx, tx, appAudit{App: a, SimpleAudit: sa})
	if err != nil {
		return testSeed{}, errs.E(op, err)
	}

	// attach test org to test user
	aoaParams = attachOrgAssociationParams{
		Org:   o,
		User:  testUser,
		Audit: ps.Audit,
	}
	err = attachOrgAssociation(ctx, tx, aoaParams)
	if err != nil {
		return testSeed{}, errs.E(op, err)
	}

	seed := testSeed{
		TestOrg:  o,
		TestApp:  a,
		TestUser: testUser,
	}

	return seed, nil
}

func (s *GenesisService) seedUserInitiatedData(ctx context.Context, tx pgx.Tx, ps principalSeed, r *diygoapi.GenesisRequest) (userInitiatedSeed, error) {
	const op errs.Op = "service/GenesisService.seedUserInitiatedData"

	var err error

	// create Org
	o := &diygoapi.Org{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Name:        r.UserInitiatedOrg.Name,
		Description: r.UserInitiatedOrg.Description,
		Kind:        ps.StandardOrgKind,
	}

	provider := diygoapi.ParseProvider(r.UserInitiatedOrg.CreateAppRequest.Oauth2Provider)

	nap := newAppParams{
		Name:             r.UserInitiatedOrg.CreateAppRequest.Name,
		Description:      r.UserInitiatedOrg.CreateAppRequest.Description,
		Org:              o,
		ApiKeyGenerator:  s.APIKeyGenerator,
		EncryptionKey:    s.EncryptionKey,
		Provider:         provider,
		ProviderClientID: r.UserInitiatedOrg.CreateAppRequest.Oauth2ProviderClientID,
	}

	var a *diygoapi.App
	a, err = newApp(nap)
	if err != nil {
		return userInitiatedSeed{}, errs.E(op, errs.Internal, err)
	}

	sa := &diygoapi.SimpleAudit{
		Create: ps.Audit,
		Update: ps.Audit,
	}

	// write the Org to the database
	err = createOrgTx(ctx, tx, &orgAudit{Org: o, SimpleAudit: sa})
	if err != nil {
		return userInitiatedSeed{}, errs.E(op, err)
	}

	// app is also to be created, write it to the db
	err = createAppTx(ctx, tx, appAudit{App: a, SimpleAudit: sa})
	if err != nil {
		return userInitiatedSeed{}, errs.E(op, err)
	}

	// associate existing User to newly created Org
	aoaParams := attachOrgAssociationParams{
		Org:   o,
		User:  ps.Audit.User,
		Audit: ps.Audit,
	}
	err = attachOrgAssociation(ctx, tx, aoaParams)
	if err != nil {
		return userInitiatedSeed{}, errs.E(op, err)
	}

	ui := userInitiatedSeed{
		UserInitiatedOrg: o,
		UserInitiatedApp: a,
	}

	return ui, nil
}

func genesisHasOccurred(ctx context.Context, dbtx datastore.DBTX) (err error) {
	const op errs.Op = "service/genesisHasOccurred"

	var (
		existingOrgs         []datastore.FindOrgsByKindExtlIDRow
		hasGenesisOrgTypeRow = true
		hasGenesisOrgRow     = true
	)

	// validate Principal records do not exist already
	// first: check org_type
	_, err = datastore.New(dbtx).FindOrgKindByExtlID(ctx, principalOrgKind)
	if err != nil {
		if err != pgx.ErrNoRows {
			return errs.E(op, errs.Database, err)
		}
		hasGenesisOrgTypeRow = false
	}

	// last: check org
	existingOrgs, err = datastore.New(dbtx).FindOrgsByKindExtlID(ctx, principalOrgKind)
	if err != nil {
		return errs.E(op, errs.Database, err)
	}
	if len(existingOrgs) == 0 {
		hasGenesisOrgRow = false
	}

	if hasGenesisOrgTypeRow || hasGenesisOrgRow {
		return errs.E(op, errs.Validation, "No prior data should exist when executing Genesis Service")
	}

	return nil
}

// ReadConfig reads the generated config file from Genesis
// and returns it in the response body
func (s *GenesisService) ReadConfig() (gr diygoapi.GenesisResponse, err error) {
	const op errs.Op = "service/GenesisService.ReadConfig"

	var b []byte
	b, err = os.ReadFile(LocalJSONGenesisResponseFile)
	if err != nil {
		return diygoapi.GenesisResponse{}, errs.E(op, err)
	}
	err = json.Unmarshal(b, &gr)
	if err != nil {
		return diygoapi.GenesisResponse{}, errs.E(op, err)
	}

	return gr, nil
}

func seedPermissions(ctx context.Context, tx pgx.Tx, r *diygoapi.GenesisRequest, adt diygoapi.Audit) (err error) {
	const op errs.Op = "service/seedPermissions"

	for _, p := range r.CreatePermissionRequests {
		_, err = createPermissionTx(ctx, tx, &p, adt)
		if err != nil {
			return errs.E(op, err)
		}
	}

	return nil
}

// roles created as part of Genesis event
type genesisRoles struct {
	// RequestRoles are roles created from the user request input
	RequestRoles []diygoapi.Role
	// TestRole is the role created for designating test users
	TestRole diygoapi.Role
}

func seedRoles(ctx context.Context, tx pgx.Tx, r *diygoapi.GenesisRequest, adt diygoapi.Audit) (genesisRoles, error) {
	const op errs.Op = "service/seedRoles"

	var (
		requestRoles    []diygoapi.Role
		rolePermissions []*diygoapi.Permission
		err             error
	)

	// seed roles from Genesis request (slice of CreateRoleRequest contained within)
	for _, crr := range r.CreateRoleRequests {

		role := diygoapi.Role{
			ID:          uuid.New(),
			ExternalID:  secure.NewID(),
			Code:        crr.Code,
			Description: crr.Description,
			Active:      crr.Active,
		}

		// find and add corresponding Permissions to the role
		rolePermissions, err = findPermissions(ctx, tx, crr.Permissions)
		if err != nil {
			return genesisRoles{}, errs.E(op, err)
		}
		role.Permissions = rolePermissions

		// add the Test user and Genesis input user to roles by attaching their external ids
		err = createRoleTx(ctx, tx, role, adt)
		if err != nil {
			return genesisRoles{}, errs.E(op, err)
		}
		requestRoles = append(requestRoles, role)
	}

	// seed testAdmin role
	// all permissions are given to this role as it's for testing only
	testRole := diygoapi.Role{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Code:        TestRoleCode,
		Description: "Role created for the sole purpose of unit and db integration testing.",
		Active:      false,
		Permissions: rolePermissions,
	}

	err = createRoleTx(ctx, tx, testRole, adt)
	if err != nil {
		return genesisRoles{}, errs.E(op, err)
	}

	roles := genesisRoles{
		RequestRoles: requestRoles,
		TestRole:     testRole,
	}

	return roles, nil
}

type grantRoles2GenesisUsersParams struct {
	Roles             genesisRoles
	PrincipalSeed     principalSeed
	TestSeed          testSeed
	UserInitiatedSeed userInitiatedSeed
}

// grantRoles2GenesisUsers grants whichever roles are included as
// part of the Genesis request as well as the testAdmin role to flag
// the test user
func grantRoles2GenesisUsers(ctx context.Context, tx pgx.Tx, params grantRoles2GenesisUsersParams) error {
	const op errs.Op = "service/grantRoles2GenesisUsers"

	var err error

	// grant roles from the Genesis request to The Creator
	for _, requestRole := range params.Roles.RequestRoles {
		aorParams := grantOrgRoleParams{
			Role:  requestRole,
			User:  params.PrincipalSeed.Audit.User,
			Org:   params.PrincipalSeed.PrincipalOrg,
			Audit: params.PrincipalSeed.Audit,
		}

		err = grantOrgRole(ctx, tx, aorParams)
		if err != nil {
			return errs.E(op, err)
		}

		p := grantOrgRoleParams{
			Role:  requestRole,
			User:  params.PrincipalSeed.Audit.User,
			Org:   params.UserInitiatedSeed.UserInitiatedOrg,
			Audit: params.PrincipalSeed.Audit,
		}

		err = grantOrgRole(ctx, tx, p)
		if err != nil {
			return errs.E(op, err)
		}
	}

	aorParams := grantOrgRoleParams{
		Role:  params.Roles.TestRole,
		User:  params.TestSeed.TestUser,
		Org:   params.TestSeed.TestOrg,
		Audit: params.PrincipalSeed.Audit,
	}

	err = grantOrgRole(ctx, tx, aorParams)
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}
