package authgateway

import (
	"context"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"

	googleoauth "google.golang.org/api/oauth2/v2"

	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/user"
	"github.com/gilcrest/go-api-basic/domain/user/usertest"
)

func Test_newUser(t *testing.T) {
	c := qt.New(t)

	type args struct {
		userinfo *googleoauth.Userinfo
	}

	ui := &googleoauth.Userinfo{
		Email:      "otto.maddox@helpinghandacceptanceco.com",
		FamilyName: "Maddox",
		GivenName:  "Otto",
		Name:       "Otto Maddox",
		Hd:         "helpinghand.com",
		Link:       "google.com/ottoprofile",
		Picture:    "google.com/picture",
	}

	u := user.User{
		Email:        "otto.maddox@helpinghandacceptanceco.com",
		LastName:     "Maddox",
		FirstName:    "Otto",
		FullName:     "Otto Maddox",
		HostedDomain: "helpinghand.com",
		PictureURL:   "google.com/picture",
		ProfileLink:  "google.com/ottoprofile",
	}

	tests := []struct {
		name string
		args args
		want user.User
	}{
		{"typical", args{ui}, u},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newUser(tt.args.userinfo)
			c.Assert(got, qt.Equals, tt.want)
		})
	}
}

func TestGoogleToken2User_User(t *testing.T) {
	c := qt.New(t)

	// set environment variable NO_INT to skip integration
	// dependent tests
	if os.Getenv("SKIP_INT") == "true" {
		t.Skip("skipping integration test")
	}

	type args struct {
		ctx   context.Context
		token auth.AccessToken
	}
	ctx := context.Background()

	// use the Google oauth2 playground https://developers.google.com/oauthplayground/
	// to get a valid Access token to test this function
	token, ok := os.LookupEnv("GOOGLE_ACCESS_TOKEN")
	if !ok {
		t.Fatalf("GOOGLE_ACCESS_TOKEN environment variable not properly set\nSet environment variable SKIP_INT = true to skip integration tests")
	}

	at := auth.AccessToken{
		Token:     token,
		TokenType: auth.BearerTokenType,
	}
	u := usertest.NewUser(t)
	bt := auth.AccessToken{
		Token:     "badToken",
		TokenType: auth.BearerTokenType,
	}

	zeroUser := user.User{}

	tests := []struct {
		name    string
		args    args
		want    user.User
		wantErr bool
	}{
		{"typical", args{ctx, at}, u, false},
		{"bad token", args{ctx, bt}, zeroUser, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gat := GoogleAccessTokenConverter{}
			got, err := gat.Convert(tt.args.ctx, tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, nil expected", err)
				return
			}
			if got != zeroUser {
				c.Assert(got, qt.Equals, tt.want)
				return
			}
			c.Assert(err, qt.Not(qt.IsNil))
		})
	}
}
