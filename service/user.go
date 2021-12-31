package service

import (
	"context"
	"time"

	"github.com/gilcrest/go-api-basic/domain/secure"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"
	"golang.org/x/oauth2"

	"github.com/gilcrest/go-api-basic/datastore/userstore"
	"github.com/gilcrest/go-api-basic/domain/app"
	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/org"
	"github.com/gilcrest/go-api-basic/domain/person"
	"github.com/gilcrest/go-api-basic/domain/user"
	"github.com/gilcrest/go-api-basic/gateway/authgateway"
)

// GoogleOauth2TokenConverter converts an Oauth2 token to a google Userinfo struct
type GoogleOauth2TokenConverter interface {
	Convert(ctx context.Context, realm string, token oauth2.Token) (authgateway.Userinfo, error)
}

// FindUserParams is the parameters for the FindUser function
type FindUserParams struct {
	Realm    string
	App      app.App
	Provider auth.Provider
	Token    oauth2.Token
}

// FindUserService retrieves a User from the Database
type FindUserService struct {
	GoogleOauth2TokenConverter GoogleOauth2TokenConverter
	Datastorer                 Datastorer
}

// FindUserByOauth2Token retrieves a users' identity from a Provider
// and then retrieves the associated registered user from the datastore
func (fus FindUserService) FindUserByOauth2Token(ctx context.Context, params FindUserParams) (user.User, error) {
	var (
		emptyUser user.User
		uInfo     authgateway.Userinfo
		err       error
	)

	if params.Provider == auth.Invalid {
		return emptyUser, errs.E(errs.Unauthenticated, errs.Realm(params.Realm), "Provider not recognized")
	}

	if params.Provider == auth.Google {
		uInfo, err = fus.GoogleOauth2TokenConverter.Convert(ctx, params.Realm, params.Token)
		if err != nil {
			return emptyUser, err
		}
	}

	findUserByUsernameParams := userstore.FindUserByUsernameParams{
		Username: uInfo.Username,
		OrgID:    params.App.Org.ID,
	}

	findUserByUsernameRow, err := userstore.New(fus.Datastorer.Pool()).FindUserByUsername(ctx, findUserByUsernameParams)
	if err != nil {
		if err == pgx.ErrNoRows {
			return emptyUser, errs.E(errs.Unauthenticated, errs.Realm(params.Realm), "No user registered in database")
		}
		return emptyUser, errs.E(errs.Unauthenticated, errs.Realm(params.Realm), err)
	}

	return hydrateUserFromDB(findUserByUsernameRow), nil
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
