package user

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	qt "github.com/frankban/quicktest"
)

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
				Email:        tt.fields.Email,
				LastName:     tt.fields.LastName,
				FirstName:    tt.fields.FirstName,
				FullName:     tt.fields.FullName,
				HostedDomain: tt.fields.HostedDomain,
				PictureURL:   tt.fields.PictureURL,
				ProfileLink:  tt.fields.ProfileLink,
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

	otto := User{
		Email:        "otto.maddox@helpinghandacceptanceco.com",
		LastName:     "Maddox",
		FirstName:    "Otto",
		FullName:     "Otto Maddox",
		HostedDomain: "",
		PictureURL:   "",
		ProfileLink:  "",
	}

	invalidOtto := User{
		Email:        "otto.maddox@helpinghandacceptanceco.com",
		LastName:     "",
		FirstName:    "Otto",
		FullName:     "Otto Maddox",
		HostedDomain: "",
		PictureURL:   "",
		ProfileLink:  "",
	}

	ctx := context.Background()
	ctx = CtxWithUser(ctx, otto)
	r = r.WithContext(ctx)

	noUserRequest, err := http.NewRequest(http.MethodGet, "/api/v1/movies", nil)
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
