package diygoapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"

	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/secure"
)

func TestUserFromRequest(t *testing.T) {
	t.Run("typical", func(t *testing.T) {
		c := qt.New(t)

		r := httptest.NewRequest(http.MethodGet, "/api/v1/movies", nil)

		want := &User{
			ID:                  uuid.New(),
			ExternalID:          secure.NewID(),
			NamePrefix:          "",
			FirstName:           "Otto",
			MiddleName:          "",
			LastName:            "Maddox",
			FullName:            "Otto Maddox",
			NameSuffix:          "",
			Nickname:            "",
			Email:               "otto.maddox@helpinghandacceptanceco.com",
			CompanyName:         "",
			CompanyDepartment:   "",
			JobTitle:            "",
			BirthDate:           time.Time{},
			LanguagePreferences: nil,
			HostedDomain:        "",
			PictureURL:          "",
			ProfileLink:         "",
			Source:              "",
		}

		ctx := NewContextWithUser(context.Background(), want)
		r = r.WithContext(ctx)

		got, err := UserFromRequest(r)
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.DeepEquals, want)
	})
	t.Run("no person added to Request context", func(t *testing.T) {
		c := qt.New(t)

		r := httptest.NewRequest(http.MethodGet, "/api/v1/ping", nil)

		wantErr := errs.E(errs.Internal, "User not set properly to context")

		got, err := UserFromRequest(r)
		c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), wantErr)
		c.Assert(got, qt.IsNil)
	})
	t.Run("user added but invalid", func(t *testing.T) {
		c := qt.New(t)

		r := httptest.NewRequest(http.MethodGet, "/api/v1/movies", nil)

		wantErr := errs.E(errs.Validation, "User LastName cannot be empty")

		invalidOtto := &User{
			ID:                  uuid.New(),
			ExternalID:          secure.NewID(),
			NamePrefix:          "",
			FirstName:           "Otto",
			MiddleName:          "",
			LastName:            "Maddox",
			FullName:            "Otto Maddox",
			NameSuffix:          "",
			Nickname:            "",
			Email:               "otto.maddox@helpinghandacceptanceco.com",
			CompanyName:         "",
			CompanyDepartment:   "",
			JobTitle:            "",
			BirthDate:           time.Time{},
			LanguagePreferences: nil,
			HostedDomain:        "",
			PictureURL:          "",
			ProfileLink:         "",
			Source:              "",
		}
		invalidOtto.Email = "otto.maddox@helpinghandacceptanceco.com"
		invalidOtto.LastName = ""
		invalidOtto.FirstName = "Otto"
		invalidOtto.FullName = "Otto Maddox"

		ctx := NewContextWithUser(context.Background(), invalidOtto)
		r = r.WithContext(ctx)

		got, err := UserFromRequest(r)
		c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), wantErr)
		c.Assert(got, qt.IsNil)
	})
}
