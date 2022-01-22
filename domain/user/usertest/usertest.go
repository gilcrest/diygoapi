// Package usertest provides testing helper functions for the
// user package
package usertest

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/gilcrest/go-api-basic/domain/org"
	"github.com/gilcrest/go-api-basic/domain/person"
	"github.com/gilcrest/go-api-basic/domain/user"
)

// NewUser provides a User for testing
func NewUser(t *testing.T) user.User {
	t.Helper()

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
