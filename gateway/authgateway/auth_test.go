package authgateway

import (
	"context"
	"reflect"
	"testing"

	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/user"
	googleoauth "google.golang.org/api/oauth2/v2"
)

func Test_newUser(t *testing.T) {
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

	u := &user.User{
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
		want *user.User
	}{
		{"typical", args{ui}, u},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newUser(tt.args.userinfo); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGoogleToken2User_User(t *testing.T) {
	type args struct {
		ctx   context.Context
		token auth.AccessToken
	}
	ctx := context.Background()
	// use the Google oauth2 playground https://developers.google.com/oauthplayground/
	// to get a valid Access token to test this function
	at := auth.AccessToken{
		Token:     "ya29.A0AfH6SMBnm00fV6q1a9txFgFntz0p3eNXLGfpUFJYMcJbUwXr009nUuobSQ9lpD7DkbWFcntJkaXfqmYUpl2xM7qEwj8Qo9JhEMHwC5gT1HAEI-CipvcpsMPDCm1pSnn5XM8xLXDAt4sm2AJ2s43psCfHdmGZ",
		TokenType: "Bearer",
	}
	u := &user.User{Email: "otto.maddox711@gmail.com",
		LastName:   "Maddox",
		FirstName:  "Otto",
		FullName:   "Otto Maddox",
		PictureURL: "https://lh3.googleusercontent.com/-RYXuWxjLdxo/AAAAAAAAAAI/AAAAAAAAAAA/AMZuucmr33m3QLTjdUatlkppT3NSN5-s8g/s96-c/photo.jpg",
	}
	bt := auth.AccessToken{
		Token:     "badToken",
		TokenType: "Bearer",
	}

	tests := []struct {
		name    string
		args    args
		want    *user.User
		wantErr bool
	}{
		{"typical", args{ctx, at}, u, false},
		{"bad token", args{ctx, bt}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := GoogleToken2User{}
			got, err := c.User(tt.args.ctx, tt.args.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("User() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("User() got = %v, want %v", got, tt.want)
			}
		})
	}
}
