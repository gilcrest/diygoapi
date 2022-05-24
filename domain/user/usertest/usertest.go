// Package usertest provides testing helper functions for the
// user package
package usertest

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v4"

	"github.com/google/uuid"

	"github.com/gilcrest/diy-go-api/domain/org"
	"github.com/gilcrest/diy-go-api/domain/person"
	"github.com/gilcrest/diy-go-api/domain/user"
)

// Setup returns a User and teardown function for testing
func Setup(ctx context.Context, t *testing.T, tx pgx.Tx) (user.User, func()) {
	t.Helper()

	//userstore.New(tx).FindUserByUsername(ctx)

	return user.User{
			ID:       uuid.New(),
			Username: "otto.maddox711@gmail.com",
			Org:      org.Org{},
			Profile: person.Profile{
				ID:                uuid.New(),
				Person:            person.Person{},
				NamePrefix:        "",
				FirstName:         "Otto",
				MiddleName:        "",
				LastName:          "Maddox",
				FullName:          "Otto Maddox",
				NameSuffix:        "",
				Nickname:          "",
				CompanyName:       "",
				CompanyDepartment: "",
				JobTitle:          "",
				BirthDate:         time.Date(2008, 1, 17, 0, 0, 0, 0, time.UTC),
				LanguageID:        uuid.Nil,
				HostedDomain:      "",
				PictureURL:        "https://lh3.googleusercontent.com/-RYXuWxjLdxo/AAAAAAAAAAI/AAAAAAAAAAA/AMZuucmr33m3QLTjdUatlkppT3NSN5-s8g/s96-c/photo.jpg",
				ProfileLink:       "",
				ProfileSource:     "",
			},
		}, func() {

		}
}

// NewInvalidUser returns an invalid user defined by the method user.IsValid()
func NewInvalidUser(t *testing.T) user.User {
	t.Helper()

	return user.User{
		ID:       uuid.New(),
		Username: "otto.maddox711@gmail.com",
		Org:      org.Org{},
		Profile: person.Profile{
			ID:                uuid.New(),
			Person:            person.Person{},
			NamePrefix:        "",
			FirstName:         "",
			MiddleName:        "",
			LastName:          "",
			FullName:          "Otto Maddox",
			NameSuffix:        "",
			Nickname:          "",
			CompanyName:       "",
			CompanyDepartment: "",
			JobTitle:          "",
			BirthDate:         time.Date(2008, 1, 17, 0, 0, 0, 0, time.UTC),
			LanguageID:        uuid.Nil,
			HostedDomain:      "",
			PictureURL:        "https://lh3.googleusercontent.com/-RYXuWxjLdxo/AAAAAAAAAAI/AAAAAAAAAAA/AMZuucmr33m3QLTjdUatlkppT3NSN5-s8g/s96-c/photo.jpg",
			ProfileLink:       "",
			ProfileSource:     "",
		},
	}
}
