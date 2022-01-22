package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/uuid"

	"github.com/gilcrest/go-api-basic/domain/org"
	"github.com/gilcrest/go-api-basic/domain/person"
)

// TODO - these tests were built before I had the concept of Profiles, Orgs, etc. - need updating
func TestUser_IsValid(t *testing.T) {
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
		name   string
		fields fields
		want   bool
	}{
		{"typical", otto, true},
		{"no email", noEmail, false},
		{"no last name", noLastName, false},
		{"no first name", noFirstName, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := User{
				ID:       uuid.New(),
				Username: tt.fields.Email,
				Org:      org.Org{},
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
			if got := u.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromRequest(t *testing.T) {
	c := qt.New(t)

	type args struct {
		r *http.Request
	}

	r := httptest.NewRequest(http.MethodGet, "/api/v1/movies", nil)

	otto := User{}
	otto.Username = "otto.maddox@helpinghandacceptanceco.com"
	otto.Profile.LastName = "Maddox"
	otto.Profile.FirstName = "Otto"
	otto.Profile.FullName = "Otto Maddox"

	invalidOtto := User{}
	invalidOtto.Username = "otto.maddox@helpinghandacceptanceco.com"
	invalidOtto.Profile.LastName = ""
	invalidOtto.Profile.FirstName = "Otto"
	invalidOtto.Profile.FullName = "Otto Maddox"

	ctx := context.Background()
	ctx = CtxWithUser(ctx, otto)
	r = r.WithContext(ctx)

	noUserRequest, err := http.NewRequest(http.MethodGet, "/api/v1/ping", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}

	invalidUserRequest := httptest.NewRequest(http.MethodGet, "/api/v1/movies", nil)

	ctx2 := context.Background()
	ctx2 = CtxWithUser(ctx2, invalidOtto)
	invalidUserRequest = invalidUserRequest.WithContext(ctx2)

	tests := []struct {
		name    string
		args    args
		want    User
		wantErr bool
	}{
		{"typical", args{r: r}, otto, false},
		{"no User added to Request context", args{r: noUserRequest}, User{}, true},
		{"user added but invalid", args{r: invalidUserRequest}, invalidOtto, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromRequest(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			c.Assert(got, qt.DeepEquals, tt.want)
		})
	}
}
