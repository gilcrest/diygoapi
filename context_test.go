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

		const op1 errs.Op = "diygoapi/UserFromContext"
		const op2 errs.Op = "diygoapi/UserFromRequest"
		ctxErr := errs.E(op1, errs.NotExist, "User not set properly to context")
		wantErr := errs.E(op2, ctxErr)

		u, err := UserFromRequest(r)
		c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), wantErr)
		c.Assert(u, qt.IsNil)
	})
}
