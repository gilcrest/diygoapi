package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/gilcrest/diy-go-api/domain/errs"
	"github.com/gilcrest/diy-go-api/domain/org"
	"github.com/gilcrest/diy-go-api/domain/person"
	"github.com/gilcrest/diy-go-api/domain/secure"
)

// TODO - these tests were built before I had the concept of Profiles, Orgs, etc. - need updating
func TestUser_IsValid(t *testing.T) {
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

	noEmail := fields{
		Email:        "",
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
		{"no email", noEmail, errs.E(errs.Validation, "username cannot be empty")},
		{"no last name", noLastName, errs.E(errs.Validation, "LastName cannot be empty")},
		{"no first name", noFirstName, errs.E(errs.Validation, "FirstName cannot be empty")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := User{
				ID:         uuid.New(),
				ExternalID: secure.NewID(),
				Username:   tt.fields.Email,
				Org:        org.Org{ID: uuid.New()},
				Profile: person.Profile{
					ID:                uuid.New(),
					Person:            person.Person{},
					NamePrefix:        "",
					FirstName:         tt.fields.FirstName,
					MiddleName:        "",
					LastName:          tt.fields.LastName,
					FullName:          tt.fields.FullName,
					NameSuffix:        "",
					Nickname:          "",
					CompanyName:       "",
					CompanyDepartment: "",
					JobTitle:          "",
					BirthDate:         time.Date(2008, 1, 17, 0, 0, 0, 0, time.UTC),
					LanguageID:        uuid.Nil,
					HostedDomain:      tt.fields.HostedDomain,
					PictureURL:        tt.fields.PictureURL,
					ProfileLink:       tt.fields.ProfileLink,
					ProfileSource:     "",
				},
			}
			err := u.IsValid()
			c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), tt.wantErr)
		})
	}
}

func TestFromRequest(t *testing.T) {
	t.Run("typical", func(t *testing.T) {
		c := qt.New(t)

		r := httptest.NewRequest(http.MethodGet, "/api/v1/movies", nil)

		want := User{}
		want.Org.ID = uuid.New()
		want.ExternalID = secure.NewID()
		want.Username = "otto.maddox@helpinghandacceptanceco.com"
		want.Profile.LastName = "Maddox"
		want.Profile.FirstName = "Otto"
		want.Profile.FullName = "Otto Maddox"

		ctx := CtxWithUser(context.Background(), want)
		r = r.WithContext(ctx)

		got, err := FromRequest(r)
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.DeepEquals, want)
	})
	t.Run("no User added to Request context", func(t *testing.T) {
		c := qt.New(t)

		r := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)

		wantErr := errs.E(errs.Internal, "User not set properly to context")
		want := User{}

		got, err := FromRequest(r)
		c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), wantErr)
		c.Assert(got, qt.DeepEquals, want)
	})
	t.Run("user added but invalid", func(t *testing.T) {
		c := qt.New(t)

		r := httptest.NewRequest(http.MethodGet, "/api/v1/movies", nil)

		wantErr := errs.E(errs.Validation, "LastName cannot be empty")

		want := User{}

		invalidOtto := User{}
		invalidOtto.Org.ID = uuid.New()
		invalidOtto.ExternalID = secure.NewID()
		invalidOtto.Username = "otto.maddox@helpinghandacceptanceco.com"
		invalidOtto.Profile.LastName = ""
		invalidOtto.Profile.FirstName = "Otto"
		invalidOtto.Profile.FullName = "Otto Maddox"

		ctx := CtxWithUser(context.Background(), invalidOtto)
		r = r.WithContext(ctx)

		got, err := FromRequest(r)
		c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), wantErr)
		c.Assert(got, qt.DeepEquals, want)
	})
}
