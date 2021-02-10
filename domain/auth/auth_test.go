package auth

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/gilcrest/go-api-basic/domain/user"

	"golang.org/x/oauth2"
)

func TestAccessToken_NewGoogleOauth2Token(t *testing.T) {
	type fields struct {
		Token     string
		TokenType string
	}

	gtoken := &oauth2.Token{
		AccessToken: "abcdef123",
		TokenType:   "Bearer",
	}

	tests := []struct {
		name   string
		fields fields
		want   *oauth2.Token
	}{
		{"typical", fields{Token: "abcdef123", TokenType: "Bearer"}, gtoken},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			at := AccessToken{
				Token:     tt.fields.Token,
				TokenType: tt.fields.TokenType,
			}
			if got := at.NewGoogleOauth2Token(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGoogleOauth2Token() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultAuthorizer_Authorize(t *testing.T) {
	type args struct {
		ctx context.Context
		sub *user.User
		obj string
		act string
	}

	ctx := context.Background()
	u := &user.User{Email: "gilcrest@gmail.com"}
	invalidUser := &user.User{Email: "badactor@gmail.com"}
	obj := "/api/v1/movies"
	act := http.MethodGet

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"typical", args{ctx, u, obj, act}, false},
		{"typical", args{ctx, invalidUser, obj, act}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := DefaultAuthorizer{}
			if err := a.Authorize(tt.args.ctx, tt.args.sub, tt.args.obj, tt.args.act); (err != nil) != tt.wantErr {
				t.Errorf("Authorize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetAccessToken2Context(t *testing.T) {
	type args struct {
		ctx       context.Context
		token     string
		tokenType string
	}
	ctx := context.Background()
	token := "abcdef123"
	bearer := "Bearer"

	at := AccessToken{
		Token:     token,
		TokenType: bearer,
	}

	wantCtx := context.WithValue(ctx, contextKeyAccessToken, at)

	tests := []struct {
		name string
		args args
		want context.Context
	}{
		{"typical", args{ctx, token, bearer}, wantCtx},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SetAccessToken2Context(tt.args.ctx, tt.args.token, tt.args.tokenType); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetAccessToken2Context() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromRequest(t *testing.T) {
	type args struct {
		r *http.Request
	}
	token := "abcdef123"
	bearer := "Bearer"

	r, err := http.NewRequest(http.MethodGet, "/api/v1/movies", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	at := AccessToken{
		Token:     token,
		TokenType: bearer,
	}

	ctx := context.Background()
	ctx = SetAccessToken2Context(ctx, token, bearer)
	r = r.WithContext(ctx)

	noAccessTokenRequest, err := http.NewRequest(http.MethodGet, "/api/v1/movies", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}

	noTokenRequest, err := http.NewRequest(http.MethodGet, "/api/v1/movies", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	ctx2 := context.Background()
	ctx2 = SetAccessToken2Context(ctx2, "", bearer)
	noTokenRequest = noTokenRequest.WithContext(ctx2)
	at2 := AccessToken{
		Token:     "",
		TokenType: bearer,
	}

	tests := []struct {
		name    string
		args    args
		want    AccessToken
		wantErr bool
	}{
		{"typical", args{r: r}, at, false},
		{"no AccessToken", args{r: noAccessTokenRequest}, AccessToken{}, true},
		{"no token", args{r: noTokenRequest}, at2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromRequest(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}
