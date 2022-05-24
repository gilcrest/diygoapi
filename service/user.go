package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"

	"github.com/gilcrest/diy-go-api/datastore/personstore"
	"github.com/gilcrest/diy-go-api/datastore/userstore"
	"github.com/gilcrest/diy-go-api/domain/audit"
	"github.com/gilcrest/diy-go-api/domain/errs"
	"github.com/gilcrest/diy-go-api/domain/org"
	"github.com/gilcrest/diy-go-api/domain/person"
	"github.com/gilcrest/diy-go-api/domain/secure"
	"github.com/gilcrest/diy-go-api/domain/user"
	"github.com/gilcrest/diy-go-api/gateway/authgateway"
)

// RegisterUserService represents a service for managing new User
// registration.
type RegisterUserService struct {
	Datastorer Datastorer
}

// SelfRegister is used to register a User with an Organization. This is "self registration" as opposed to one user
// registering another user.
func (s RegisterUserService) SelfRegister(ctx context.Context, adt audit.Audit) (err error) {
	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	err = createUserTx(ctx, tx, adt.User, adt)
	if err != nil {
		return err
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return err
	}

	return nil
}

// createUserTx creates a user in the database given a domain user.User and audit.Audit
// If it is a self registration, u and adt.User will be the same
func createUserTx(ctx context.Context, tx pgx.Tx, u user.User, adt audit.Audit) error {
	var err error

	createPersonParams := personstore.CreatePersonParams{
		PersonID:        u.Profile.Person.ID,
		OrgID:           u.Profile.Person.Org.ID,
		CreateAppID:     adt.App.ID,
		CreateUserID:    adt.User.NullUUID(),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    adt.User.NullUUID(),
		UpdateTimestamp: adt.Moment,
	}

	// create Person db record
	var rowsAffected int64
	rowsAffected, err = personstore.New(tx).CreatePerson(ctx, createPersonParams)
	if err != nil {
		return errs.E(errs.Database, err)
	}

	if rowsAffected != 1 {
		return errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	// create Person Profile db record
	createPersonProfileParams := personstore.CreatePersonProfileParams{
		PersonProfileID: u.Profile.ID,
		PersonID:        u.Profile.Person.ID,
		NamePrefix:      sql.NullString{},
		FirstName:       u.Profile.FirstName,
		MiddleName:      sql.NullString{},
		LastName:        u.Profile.LastName,
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
		CreateAppID:     adt.App.ID,
		CreateUserID:    adt.User.NullUUID(),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    adt.User.NullUUID(),
		UpdateTimestamp: adt.Moment,
	}

	var personProfileRowsAffected int64
	personProfileRowsAffected, err = personstore.New(tx).CreatePersonProfile(ctx, createPersonProfileParams)
	if err != nil {
		return errs.E(errs.Database, err)
	}

	if personProfileRowsAffected != 1 {
		return errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	createUserParams := userstore.CreateUserParams{
		UserID:          u.ID,
		UserExtlID:      u.ExternalID.String(),
		Username:        u.Username,
		OrgID:           u.Org.ID,
		PersonProfileID: u.Profile.ID,
		CreateAppID:     adt.App.ID,
		CreateUserID:    adt.User.NullUUID(),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    adt.User.NullUUID(),
		UpdateTimestamp: adt.Moment,
	}

	var userRowsAffected int64
	userRowsAffected, err = userstore.New(tx).CreateUser(ctx, createUserParams)
	if err != nil {
		return errs.E(errs.Database, err)
	}

	if userRowsAffected != 1 {
		return errs.E(errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	return nil

}

func hydrateUserFromProviderUserInfo(params FindUserParams, pui authgateway.ProviderUserInfo) user.User {

	p := person.Person{
		ID:  uuid.New(),
		Org: params.App.Org,
	}

	pfl := person.Profile{
		ID:                uuid.UUID{},
		Person:            p,
		NamePrefix:        "",
		FirstName:         pui.GivenName,
		MiddleName:        "",
		LastName:          pui.FamilyName,
		FullName:          pui.Name,
		NameSuffix:        "",
		Nickname:          "",
		CompanyName:       "",
		CompanyDepartment: "",
		JobTitle:          "",
		BirthDate:         time.Time{},
		LanguageID:        uuid.UUID{},
		HostedDomain:      pui.Hd,
		PictureURL:        pui.Picture,
		ProfileLink:       pui.Link,
		ProfileSource:     params.Provider.String(),
	}

	u := user.User{
		ID:       uuid.New(),
		Username: pui.Username,
		Org:      params.App.Org,
		Profile:  pfl,
	}

	return u
}

// findUserByID finds a user given its ID
func findUserByID(ctx context.Context, dbtx DBTX, id uuid.UUID) (user.User, error) {
	row, err := userstore.New(dbtx).FindUserByID(ctx, id)
	if err != nil {
		return user.User{}, errs.E(errs.Database, err)
	}
	u := user.User{}
	u.ID = row.UserID
	u.Username = row.Username
	o := org.Org{
		ID:          row.OrgID,
		ExternalID:  secure.MustParseIdentifier(row.OrgExtlID),
		Name:        row.OrgName,
		Description: row.OrgDescription,
	}
	p := person.Person{
		ID:  row.PersonID,
		Org: o,
	}
	pp := person.Profile{
		ID:                row.PersonProfileID,
		Person:            p,
		NamePrefix:        row.NamePrefix.String,
		FirstName:         row.FirstName,
		MiddleName:        row.MiddleName.String,
		LastName:          row.LastName,
		NameSuffix:        row.NameSuffix.String,
		Nickname:          row.Nickname.String,
		CompanyName:       row.CompanyName.String,
		CompanyDepartment: row.CompanyDept.String,
		JobTitle:          row.JobTitle.String,
		BirthDate:         time.Time{},
		LanguageID:        row.LanguageID.UUID,
		HostedDomain:      "",
		PictureURL:        "",
		ProfileLink:       "",
		ProfileSource:     "",
	}
	u.Org = o
	u.Profile = pp

	return u, nil
}

func hydrateUserFromUsernameRow(row userstore.FindUserByUsernameRow) user.User {
	u := user.User{}
	u.ID = row.UserID
	u.ExternalID = secure.MustParseIdentifier(row.UserExtlID)
	u.Username = row.Username

	o := org.Org{
		ID:          row.OrgID,
		ExternalID:  secure.MustParseIdentifier(row.OrgExtlID),
		Name:        row.OrgName,
		Description: row.OrgDescription,
	}

	p := person.Person{
		ID:  row.PersonID,
		Org: o,
	}

	pp := person.Profile{
		ID:                row.PersonProfileID,
		Person:            p,
		NamePrefix:        row.NamePrefix.String,
		FirstName:         row.FirstName,
		MiddleName:        row.MiddleName.String,
		LastName:          row.LastName,
		NameSuffix:        row.NameSuffix.String,
		Nickname:          row.Nickname.String,
		CompanyName:       row.CompanyName.String,
		CompanyDepartment: row.CompanyDept.String,
		JobTitle:          row.JobTitle.String,
		BirthDate:         time.Time{},
		LanguageID:        row.LanguageID.UUID,
		HostedDomain:      "",
		PictureURL:        "",
		ProfileLink:       "",
		ProfileSource:     "",
	}
	u.Org = o
	u.Profile = pp

	return u
}

func hydrateUserFromExternalIDRow(row userstore.FindUserByExternalIDRow) user.User {
	u := user.User{}
	u.ID = row.UserID
	u.ExternalID = secure.MustParseIdentifier(row.UserExtlID)
	u.Username = row.Username

	o := org.Org{
		ID:          row.OrgID,
		ExternalID:  secure.MustParseIdentifier(row.OrgExtlID),
		Name:        row.OrgName,
		Description: row.OrgDescription,
	}

	p := person.Person{
		ID:  row.PersonID,
		Org: o,
	}

	pp := person.Profile{
		ID:                row.PersonProfileID,
		Person:            p,
		NamePrefix:        row.NamePrefix.String,
		FirstName:         row.FirstName,
		MiddleName:        row.MiddleName.String,
		LastName:          row.LastName,
		NameSuffix:        row.NameSuffix.String,
		Nickname:          row.Nickname.String,
		CompanyName:       row.CompanyName.String,
		CompanyDepartment: row.CompanyDept.String,
		JobTitle:          row.JobTitle.String,
		BirthDate:         time.Time{},
		LanguageID:        row.LanguageID.UUID,
		HostedDomain:      "",
		PictureURL:        "",
		ProfileLink:       "",
		ProfileSource:     "",
	}
	u.Org = o
	u.Profile = pp

	return u
}
