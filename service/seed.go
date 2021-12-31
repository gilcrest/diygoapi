package service

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/datastore/appstore"
	"github.com/gilcrest/go-api-basic/datastore/orgstore"
	"github.com/gilcrest/go-api-basic/datastore/personstore"
	"github.com/gilcrest/go-api-basic/datastore/userstore"
	"github.com/gilcrest/go-api-basic/domain/app"
	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/org"
	"github.com/gilcrest/go-api-basic/domain/person"
	"github.com/gilcrest/go-api-basic/domain/secure"
	"github.com/gilcrest/go-api-basic/domain/user"
	"github.com/google/uuid"
)

// SeedRequest is the request struct for seeding the database
type SeedRequest struct {
	OrgName           string `json:"org_name"`
	OrgDescription    string `json:"org_description"`
	AppName           string `json:"app_name"`
	AppDescription    string `json:"app_description"`
	SeedUsername      string `json:"seed_username"`
	SeedUserFirstName string `json:"seed_user_first_name"`
	SeedUserLastName  string `json:"seed_user_last_name"`
}

// SeedResponse is the response struct for seeding the database
type SeedResponse struct {
	OrgResponse OrgResponse `json:"org"`
	AppResponse AppResponse `json:"app"`
}

// SeedService seeds the database. It is run only once on initial database setup.
type SeedService struct {
	Datastorer            Datastorer
	CryptoRandomGenerator CryptoRandomGenerator
	EncryptionKey         *[32]byte
}

// Seed method seeds the database
func (sr SeedService) Seed(ctx context.Context, r *SeedRequest) (SeedResponse, error) {

	// start db txn using pgxpool
	tx, err := sr.Datastorer.BeginTx(ctx)
	if err != nil {
		return SeedResponse{}, err
	}

	orgCount, err := orgstore.New(tx).CountOrgs(ctx)
	if err != nil {
		return SeedResponse{}, errs.E(errs.Database, err)
	}
	if orgCount != 0 {
		return SeedResponse{}, errs.E(errs.Validation, "database has already been seeded.")
	}

	// create Org
	o := org.Org{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Name:        r.OrgName,
		Description: r.OrgDescription,
	}

	// initialize App and inject dependent fields
	a := app.App{
		ID:          uuid.New(),
		ExternalID:  secure.NewID(),
		Org:         o,
		Name:        r.AppName,
		Description: r.AppDescription,
		APIKeys:     nil,
	}

	// generate App API key
	aak, err := app.NewAPIKey(sr.CryptoRandomGenerator, sr.EncryptionKey)
	if err != nil {
		return SeedResponse{}, errs.E(errs.Internal, sr.Datastorer.RollbackTx(ctx, tx, err))
	}
	aak.SetDeactivationDate(time.Date(2099, 12, 31, 0, 0, 0, 0, time.UTC))
	a.APIKeys = []app.APIKey{aak}

	// create Person
	prsn := person.Person{
		ID:  uuid.New(),
		Org: o,
	}

	// create Person Profile
	pfl := person.Profile{ID: uuid.New(), Person: prsn}
	pfl.FirstName = r.SeedUserFirstName
	pfl.LastName = r.SeedUserLastName

	// create User
	u := user.User{
		ID:       uuid.New(),
		Username: strings.TrimSpace(r.SeedUsername),
		Org:      o,
		Profile:  pfl,
	}

	//create Audit
	adt := audit.Audit{
		App:    a,
		User:   u,
		Moment: time.Now(),
	}

	// add audit fields to Org
	o.CreateAppID = adt.App.ID
	o.CreateUserID = adt.User.ID
	o.CreateTime = adt.Moment
	o.UpdateAppID = adt.App.ID
	o.UpdateUserID = adt.User.ID
	o.UpdateTime = adt.Moment

	cop := orgstore.CreateOrgParams{
		OrgID:           o.ID,
		OrgExtlID:       o.ExternalID.String(),
		OrgName:         o.Name,
		OrgDescription:  o.Description,
		CreateAppID:     a.ID,
		CreateUserID:    datastore.NewNullUUID(u.ID),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     a.ID,
		UpdateUserID:    datastore.NewNullUUID(u.ID),
		UpdateTimestamp: adt.Moment,
	}

	_, err = orgstore.New(tx).CreateOrg(ctx, cop)
	if err != nil {
		return SeedResponse{}, errs.E(errs.Database, sr.Datastorer.RollbackTx(ctx, tx, err))
	}

	// create App database record using appstore
	_, err = appstore.New(tx).CreateApp(ctx, NewCreateAppParams(a, adt))
	if err != nil {
		return SeedResponse{}, errs.E(errs.Database, sr.Datastorer.RollbackTx(ctx, tx, err))
	}

	// create database record using appstore
	_, err = appstore.New(tx).CreateAppAPIKey(ctx, NewCreateAppAPIKeyParams(a, aak, adt))
	if err != nil {
		return SeedResponse{}, errs.E(errs.Database, sr.Datastorer.RollbackTx(ctx, tx, err))
	}

	cpp := personstore.CreatePersonParams{
		PersonID:        prsn.ID,
		OrgID:           prsn.Org.ID,
		CreateAppID:     a.ID,
		CreateUserID:    datastore.NewNullUUID(u.ID),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     a.ID,
		UpdateUserID:    datastore.NewNullUUID(u.ID),
		UpdateTimestamp: adt.Moment,
	}

	// create Person db record
	_, err = personstore.New(tx).CreatePerson(ctx, cpp)
	if err != nil {
		return SeedResponse{}, errs.E(errs.Database, sr.Datastorer.RollbackTx(ctx, tx, err))
	}

	// create Person Profile db record
	cppp := personstore.CreatePersonProfileParams{
		PersonProfileID: pfl.ID,
		PersonID:        prsn.ID,
		NamePrefix:      sql.NullString{},
		FirstName:       pfl.FirstName,
		MiddleName:      sql.NullString{},
		LastName:        pfl.LastName,
		NameSuffix:      sql.NullString{},
		Nickname:        sql.NullString{},
		CompanyName:     sql.NullString{},
		CompanyDept:     sql.NullString{},
		JobTitle:        sql.NullString{},
		BirthDate:       sql.NullTime{},
		BirthYear:       sql.NullInt64{},
		BirthMonth:      sql.NullInt64{},
		BirthDay:        sql.NullInt64{},
		LanguageID:      uuid.NullUUID{},
		CreateAppID:     uuid.NullUUID{},
		CreateUserID:    datastore.NewNullUUID(u.ID),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     uuid.NullUUID{},
		UpdateUserID:    datastore.NewNullUUID(u.ID),
		UpdateTimestamp: adt.Moment,
	}

	_, err = personstore.New(tx).CreatePersonProfile(ctx, cppp)
	if err != nil {
		return SeedResponse{}, errs.E(errs.Database, sr.Datastorer.RollbackTx(ctx, tx, err))
	}

	cup := userstore.CreateUserParams{
		UserID:          u.ID,
		Username:        u.Username,
		OrgID:           u.Org.ID,
		PersonProfileID: u.Profile.ID,
		CreateAppID:     a.ID,
		CreateUserID:    datastore.NewNullUUID(u.ID),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     a.ID,
		UpdateUserID:    datastore.NewNullUUID(u.ID),
		UpdateTimestamp: adt.Moment,
	}

	_, err = userstore.New(tx).CreateUser(ctx, cup)
	if err != nil {
		return SeedResponse{}, errs.E(errs.Database, sr.Datastorer.RollbackTx(ctx, tx, err))
	}

	orgResponse, err := newOrgResponse(ctx, tx, o)
	if err != nil {
		return SeedResponse{}, errs.E(errs.Database, sr.Datastorer.RollbackTx(ctx, tx, err))
	}

	// commit db txn using pgxpool
	err = sr.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return SeedResponse{}, err
	}

	response := SeedResponse{
		OrgResponse: orgResponse,
		AppResponse: newAppResponse(a),
	}

	return response, nil
}
