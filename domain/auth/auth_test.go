package auth

import (
	"context"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/gilcrest/go-api-basic/domain/logger"
	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/domain/user/usertest"

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
		TokenType:   BearerTokenType,
	}

	tests := []struct {
		name   string
		fields fields
		want   *oauth2.Token
	}{
		{"typical", fields{Token: "abcdef123", TokenType: BearerTokenType}, gtoken},
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
		lgr zerolog.Logger
		sub user.User
		obj string
		act string
	}

	lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)
	u := usertest.NewUser(t)
	invalidUser := user.User{Email: "badactor@gmail.com"}
	obj := "/api/v1/movies"
	act := http.MethodGet

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"typical", args{lgr, u, obj, act}, false},
		{"invalid user", args{lgr, invalidUser, obj, act}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := DefaultAuthorizer{}
			if err := a.Authorize(tt.args.lgr, tt.args.sub, tt.args.obj, tt.args.act); (err != nil) != tt.wantErr {
				t.Errorf("Authorize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCtxWithAccessToken(t *testing.T) {
	type args struct {
		ctx   context.Context
		token AccessToken
	}
	ctx := context.Background()
	token := "abcdef123"

	at := AccessToken{
		Token:     token,
		TokenType: BearerTokenType,
	}

	wantCtx := context.WithValue(ctx, contextKeyAccessToken, at)

	tests := []struct {
		name string
		args args
		want context.Context
	}{
		{"typical", args{ctx, at}, wantCtx},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CtxWithAccessToken(tt.args.ctx, tt.args.token); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CtxWithAccessToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromRequest(t *testing.T) {
	type args struct {
		r *http.Request
	}
	token := "abcdef123"

	r, err := http.NewRequest(http.MethodGet, "/api/v1/movies", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}
	at := AccessToken{
		Token:     token,
		TokenType: BearerTokenType,
	}

	emptyAT := AccessToken{
		Token:     "",
		TokenType: BearerTokenType,
	}

	ctx := context.Background()
	ctx = CtxWithRealm(ctx, "DeepInTheRealm")
	ctx = CtxWithAccessToken(ctx, at)
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
	ctx2 = CtxWithRealm(ctx2, "DeepInTheRealm")
	ctx2 = CtxWithAccessToken(ctx2, emptyAT)
	noTokenRequest = noTokenRequest.WithContext(ctx2)
	at2 := AccessToken{
		Token:     "",
		TokenType: BearerTokenType,
	}

	tests := []struct {
		name   string
		args   args
		want   AccessToken
		wantOK bool
	}{
		{"typical", args{r: r}, at, true},
		{"no AccessToken", args{r: noAccessTokenRequest}, AccessToken{}, false},
		{"no token", args{r: noTokenRequest}, at2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := AccessTokenFromRequest(tt.args.r)
			if ok != tt.wantOK {
				t.Errorf("AccessTokenFromRequest() ok = %v, wantOK %v", ok, tt.wantOK)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AccessTokenFromRequest() got = %v, want %v", got, tt.want)
			}
		})
	}
}
