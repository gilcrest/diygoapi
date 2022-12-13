package diygoapi

import (
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/secure"
)

func TestUser_Validate(t *testing.T) {
	c := qt.New(t)

	type fields struct {
		Email        string
		LastName     string
		FirstName    string
		FullName     string
		HostedDomain string
		PictureURL   string
		ProfileLink  string
	}

	otto := fields{
		Email:        "otto.maddox@helpinghandacceptanceco.com",
		LastName:     "Maddox",
		FirstName:    "Otto",
		FullName:     "Otto Maddox",
		HostedDomain: "",
		PictureURL:   "",
		ProfileLink:  "",
	}

	noLastName := fields{
		Email:        "otto.maddox@helpinghandacceptanceco.com",
		LastName:     "",
		FirstName:    "Otto",
		FullName:     "Otto Maddox",
		HostedDomain: "",
		PictureURL:   "",
		ProfileLink:  "",
	}

	noFirstName := fields{
		Email:        "otto.maddox@helpinghandacceptanceco.com",
		LastName:     "Maddox",
		FirstName:    "",
		FullName:     "Otto Maddox",
		HostedDomain: "",
		PictureURL:   "",
		ProfileLink:  "",
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr error
	}{
		{"typical", otto, nil},
		{"no last name", noLastName, errs.E(errs.Validation, "User LastName cannot be empty")},
		{"no first name", noFirstName, errs.E(errs.Validation, "User FirstName cannot be empty")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{
				ID:                  uuid.New(),
				ExternalID:          secure.NewID(),
				NamePrefix:          "",
				FirstName:           tt.fields.FirstName,
				MiddleName:          "",
				LastName:            tt.fields.LastName,
				FullName:            tt.fields.FullName,
				NameSuffix:          "",
				Nickname:            "",
				Email:               tt.fields.Email,
				CompanyName:         "",
				CompanyDepartment:   "",
				JobTitle:            "",
				BirthDate:           time.Date(2008, 1, 17, 0, 0, 0, 0, time.UTC),
				LanguagePreferences: nil,
				HostedDomain:        tt.fields.HostedDomain,
				PictureURL:          tt.fields.PictureURL,
				ProfileLink:         tt.fields.ProfileLink,
				Source:              "",
			}
			err := user.Validate()
			c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), tt.wantErr)
		})
	}
}
