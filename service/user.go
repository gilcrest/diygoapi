package service

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"golang.org/x/oauth2"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/datastore/personstore"
	"github.com/gilcrest/go-api-basic/datastore/userstore"
	"github.com/gilcrest/go-api-basic/domain/app"
	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/org"
	"github.com/gilcrest/go-api-basic/domain/person"
	"github.com/gilcrest/go-api-basic/domain/secure"
	"github.com/gilcrest/go-api-basic/domain/user"
	"github.com/gilcrest/go-api-basic/gateway/authgateway"
)

// GoogleOauth2TokenConverter converts an oauth2.Token to an authgateway.Userinfo struct
type GoogleOauth2TokenConverter interface {
	Convert(ctx context.Context, realm string, token oauth2.Token) (authgateway.ProviderUserInfo, error)
}

// FindUserParams is parameters for finding a User
type FindUserParams struct {
	Realm          string
	App            app.App
	Provider       auth.Provider
	Token          oauth2.Token
	RetrieveFromDB bool
}

// FindUserService represents a service for managing User retrieval
// from a Provider and/or the database.
type FindUserService struct {
	GoogleOauth2TokenConverter GoogleOauth2TokenConverter
	Datastorer                 Datastorer
}

// FindUserByOauth2Token retrieves a users' identity from a Provider
// and then retrieves the associated registered user from the datastore
func (s FindUserService) FindUserByOauth2Token(ctx context.Context, params FindUserParams) (user.User, error) {
	var (
		uInfo authgateway.ProviderUserInfo
		err   error
	)

	if params.Provider == auth.Invalid {
		return user.User{}, errs.E(errs.Unauthenticated, errs.Realm(params.Realm), "provider not recognized")
	}

	if params.Provider == auth.Apple {
		return user.User{}, errs.E(errs.Unauthenticated, errs.Realm(params.Realm), "apple authentication not yet implemented")
	}

	if params.Provider == auth.Google {
		uInfo, err = s.GoogleOauth2TokenConverter.Convert(ctx, params.Realm, params.Token)
		if err != nil {
			return user.User{}, err
		}
	}

	findUserByUsernameParams := userstore.FindUserByUsernameParams{
		Username: uInfo.Username,
		OrgID:    params.App.Org.ID,
	}

	if params.RetrieveFromDB {
		var findUserByUsernameRow userstore.FindUserByUsernameRow
		findUserByUsernameRow, err = userstore.New(s.Datastorer.Pool()).FindUserByUsername(ctx, findUserByUsernameParams)
		if err != nil {
			if err == pgx.ErrNoRows {
				return user.User{}, errs.E(errs.Unauthenticated, errs.Realm(params.Realm), "No user registered in database")
			}
			return user.User{}, errs.E(errs.Unauthenticated, errs.Realm(params.Realm), err)
		}

		return hydrateUserFromDB(findUserByUsernameRow), nil
	}

	return hydrateUserFromProviderUserInfo(params, uInfo), nil
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

func hydrateUserFromDB(row userstore.FindUserByUsernameRow) user.User {
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

	return u
}

// RegisterUserService represents a service for managing new User
// registration.
type RegisterUserService struct {
	Datastorer Datastorer
}

// SelfRegister is used to register a User with an Organization. This is "self registration" as opposed to one user
// registering another user.
func (s RegisterUserService) SelfRegister(ctx context.Context, adt audit.Audit) error {
	var err error

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return err
	}

	err = createUserDB(ctx, s.Datastorer, tx, adt.User, adt)
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

// createUserDB creates a user in the database given a domain user.User and audit.Audit
// If it is a self registration, u and adt.User will be the same
func createUserDB(ctx context.Context, ds Datastorer, tx pgx.Tx, u user.User, adt audit.Audit) error {
	var err error

	createPersonParams := personstore.CreatePersonParams{
		PersonID:        u.Profile.Person.ID,
		OrgID:           u.Profile.Person.Org.ID,
		CreateAppID:     adt.App.ID,
		CreateUserID:    datastore.NewNullUUID(adt.User.ID),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    datastore.NewNullUUID(adt.User.ID),
		UpdateTimestamp: adt.Moment,
	}

	// create Person db record
	_, err = personstore.New(tx).CreatePerson(ctx, createPersonParams)
	if err != nil {
		return errs.E(errs.Database, ds.RollbackTx(ctx, tx, err))
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
		CreateUserID:    datastore.NewNullUUID(adt.User.ID),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    datastore.NewNullUUID(adt.User.ID),
		UpdateTimestamp: adt.Moment,
	}

	_, err = personstore.New(tx).CreatePersonProfile(ctx, createPersonProfileParams)
	if err != nil {
		return errs.E(errs.Database, ds.RollbackTx(ctx, tx, err))
	}

	createUserParams := userstore.CreateUserParams{
		UserID:          u.ID,
		Username:        u.Username,
		OrgID:           u.Org.ID,
		PersonProfileID: u.Profile.ID,
		CreateAppID:     adt.App.ID,
		CreateUserID:    datastore.NewNullUUID(adt.User.ID),
		CreateTimestamp: adt.Moment,
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    datastore.NewNullUUID(adt.User.ID),
		UpdateTimestamp: adt.Moment,
	}

	_, err = userstore.New(tx).CreateUser(ctx, createUserParams)
	if err != nil {
		return errs.E(errs.Database, ds.RollbackTx(ctx, tx, err))
	}

	return nil

}
