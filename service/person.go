package service

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"golang.org/x/text/language"

	"github.com/gilcrest/diygoapi"
	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/secure"
	"github.com/gilcrest/diygoapi/sqldb/datastore"
)

// createPersonTx creates a Person in the database
// Any Users attached to the Person will also be created.
// The created User will be associated to any Orgs attached.
func createPersonTx(ctx context.Context, tx pgx.Tx, p diygoapi.Person, adt diygoapi.Audit) error {
	const op errs.Op = "service/createPersonTx"

	var err error

	createPersonParams := datastore.CreatePersonParams{
		PersonID:        p.ID,
		PersonExtlID:    p.ExternalID.String(),
		CreateAppID:     adt.App.ID,
		CreateUserID:    adt.User.NullUUID(),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    adt.User.NullUUID(),
		UpdateTimestamp: adt.Moment,
	}

	// create Person db record
	var rowsAffected int64
	rowsAffected, err = datastore.New(tx).CreatePerson(ctx, createPersonParams)
	if err != nil {
		return errs.E(op, errs.Database, err)
	}

	if rowsAffected != 1 {
		return errs.E(op, errs.Database, fmt.Sprintf("person rows affected should be 1, actual: %d", rowsAffected))
	}

	// loop through all users associated to the Person and create them
	for _, u := range p.Users {
		cuTxParams := createUserTxParams{
			PersonID: p.ID,
			User:     u,
			Audit:    adt,
		}
		err = createUserTx(ctx, tx, cuTxParams)
		if err != nil {
			return errs.E(op, err)
		}
	}

	return nil

}

func newUserResponse(u *diygoapi.User) *diygoapi.UserResponse {
	r := &diygoapi.UserResponse{
		ID:                  u.ID,
		ExternalID:          u.ExternalID,
		NamePrefix:          u.NamePrefix,
		FirstName:           u.FirstName,
		MiddleName:          u.MiddleName,
		LastName:            u.LastName,
		FullName:            u.FullName,
		NameSuffix:          u.NameSuffix,
		Nickname:            u.Nickname,
		Email:               u.Email,
		CompanyName:         u.CompanyName,
		CompanyDepartment:   u.CompanyDepartment,
		JobTitle:            u.JobTitle,
		BirthDate:           u.BirthDate,
		LanguagePreferences: u.LanguagePreferences,
		HostedDomain:        u.HostedDomain,
		PictureURL:          u.PictureURL,
		ProfileLink:         u.ProfileLink,
		Source:              u.Source,
	}

	return r
}

type createUserTxParams struct {
	// The ID of the Person the User is associated to
	PersonID uuid.UUID
	// The User to be created
	User *diygoapi.User
	// The details for which app and user created/updated the User
	Audit diygoapi.Audit
}

// createUserTx creates a User in the database
func createUserTx(ctx context.Context, tx pgx.Tx, params createUserTxParams) error {
	const op errs.Op = "service/createUserTx"

	var err error

	var birthYear, birthMonth, birthDay sql.NullInt64
	if !params.User.BirthDate.IsZero() {
		birthYear = diygoapi.NewNullInt64(int64(params.User.BirthDate.Year()))
		birthMonth = diygoapi.NewNullInt64(int64(params.User.BirthDate.Month()))
		birthDay = diygoapi.NewNullInt64(int64(params.User.BirthDate.Day()))
	}

	cuParams := datastore.CreateUserParams{
		UserID:          params.User.ID,
		UserExtlID:      params.User.ExternalID.String(),
		PersonID:        params.PersonID,
		NamePrefix:      diygoapi.NewNullString(params.User.NamePrefix),
		FirstName:       params.User.FirstName,
		MiddleName:      diygoapi.NewNullString(params.User.MiddleName),
		LastName:        params.User.LastName,
		NameSuffix:      diygoapi.NewNullString(params.User.NameSuffix),
		Nickname:        diygoapi.NewNullString(params.User.Nickname),
		Email:           diygoapi.NewNullString(params.User.Email),
		CompanyName:     diygoapi.NewNullString(params.User.CompanyName),
		CompanyDept:     diygoapi.NewNullString(params.User.CompanyDepartment),
		JobTitle:        diygoapi.NewNullString(params.User.JobTitle),
		BirthDate:       diygoapi.NewNullTime(params.User.BirthDate),
		BirthYear:       birthYear,
		BirthMonth:      birthMonth,
		BirthDay:        birthDay,
		CreateAppID:     params.Audit.App.ID,
		CreateUserID:    params.Audit.User.NullUUID(),
		CreateTimestamp: params.Audit.Moment,
		UpdateAppID:     params.Audit.App.ID,
		UpdateUserID:    params.Audit.User.NullUUID(),
		UpdateTimestamp: params.Audit.Moment,
	}

	// create User db record
	var rowsAffected int64
	rowsAffected, err = datastore.New(tx).CreateUser(ctx, cuParams)
	if err != nil {
		return errs.E(op, errs.Database, err)
	}

	if rowsAffected != 1 {
		return errs.E(op, errs.Database, fmt.Sprintf("user rows affected should be 1, actual: %d", rowsAffected))
	}

	return nil

}

// FindUserByID finds a User in the datastore given their User ID
func FindUserByID(ctx context.Context, dbtx datastore.DBTX, id uuid.UUID) (*diygoapi.User, error) {
	const op errs.Op = "service/FindUserByID"

	dbUser, err := datastore.New(dbtx).FindUserByID(ctx, id)
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	var ulp []datastore.UsersLangPref
	ulp, err = datastore.New(dbtx).FindUserLanguagePreferencesByUserID(ctx, id)
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	var langPrefs []language.Tag
	for _, p := range ulp {
		tag := language.Make(p.LanguageTag)
		langPrefs = append(langPrefs, tag)
	}

	u := &diygoapi.User{
		ID:                  dbUser.UserID,
		ExternalID:          secure.MustParseIdentifier(dbUser.UserExtlID),
		NamePrefix:          dbUser.NamePrefix.String,
		FirstName:           dbUser.FirstName,
		MiddleName:          dbUser.MiddleName.String,
		LastName:            dbUser.LastName,
		FullName:            "", // TODO - add FullName to users table (and structs)
		NameSuffix:          dbUser.NameSuffix.String,
		Nickname:            dbUser.Nickname.String,
		Gender:              "", // TODO - add Gender to db (full list)
		Email:               dbUser.Email.String,
		CompanyName:         dbUser.CompanyName.String,
		CompanyDepartment:   dbUser.CompanyDept.String,
		JobTitle:            dbUser.JobTitle.String,
		BirthDate:           dbUser.BirthDate.Time,
		LanguagePreferences: langPrefs,
		HostedDomain:        "", // TODO - add a bunch of fields to db
		PictureURL:          "",
		ProfileLink:         "",
		Source:              "",
	}

	return u, nil
}

type attachOrgAssociationParams struct {
	Org   *diygoapi.Org
	User  *diygoapi.User
	Audit diygoapi.Audit
}

// attachOrgAssociation associates an Org with a User in the database.
func attachOrgAssociation(ctx context.Context, tx pgx.Tx, params attachOrgAssociationParams) error {
	const op errs.Op = "service/attachOrgAssociation"

	createUsersOrgParams := datastore.CreateUsersOrgParams{
		UsersOrgID:      uuid.New(),
		OrgID:           params.Org.ID,
		UserID:          params.User.ID,
		CreateAppID:     params.Audit.App.ID,
		CreateUserID:    diygoapi.NewNullUUID(params.Audit.User.ID),
		CreateTimestamp: params.Audit.Moment,
		UpdateAppID:     params.Audit.App.ID,
		UpdateUserID:    diygoapi.NewNullUUID(params.Audit.User.ID),
		UpdateTimestamp: params.Audit.Moment,
	}

	// create database record using datastore
	rowsAffected, err := datastore.New(tx).CreateUsersOrg(ctx, createUsersOrgParams)
	if err != nil {
		return errs.E(op, errs.Database, err)
	}

	// should only impact exactly one record
	if rowsAffected != 1 {
		return errs.E(op, errs.Database, fmt.Sprintf("CreateUsersOrg() should insert 1 row, actual: %d", rowsAffected))
	}

	return nil
}
