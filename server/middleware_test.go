package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/oauth2"

	"github.com/gilcrest/diygoapi"
	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/logger"
)

type mockAuthenticationService struct{}

func (mockAuthenticationService) SelfRegister(ctx context.Context, params diygoapi.AuthenticationParams) (diygoapi.Auth, error) {
	panic("implement me")
}

func (mockAuthenticationService) FindAuth(ctx context.Context, params diygoapi.AuthenticationParams) (diygoapi.Auth, error) {
	panic("implement me")
}

func (mockAuthenticationService) FindAppByProviderClientID(ctx context.Context, realm string, auth diygoapi.Auth) (a *diygoapi.App, err error) {
	panic("implement me")
}

func (mockAuthenticationService) FindAppByAPIKey(ctx context.Context, realm, appExtlID, key string) (*diygoapi.App, error) {
	return &diygoapi.App{
		ID:          uuid.UUID{},
		ExternalID:  []byte("so random"),
		Org:         &diygoapi.Org{},
		Name:        "",
		Description: "",
		APIKeys:     nil,
	}, nil
}

func TestJSONContentTypeResponseHandler(t *testing.T) {

	s := Server{}

	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}

	testJSONContentTypeResponseHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Fatalf("Content-Type %s is invalid", contentType)
		}
	})

	rr := httptest.NewRecorder()

	handlers := s.jsonContentTypeResponseHandler(testJSONContentTypeResponseHandler)
	handlers.ServeHTTP(rr, req)
}

// TODO - currently using mock - should use database test to actually query db. Requires quite a bit of data setup, but is appropriate and will get to this.
func TestServer_appHandler(t *testing.T) {
	t.Run("typical - mock database", func(t *testing.T) {
		c := qt.New(t)

		req, err := http.NewRequest("GET", "/ping", nil)
		if err != nil {
			t.Fatalf("http.NewRequest() error = %v", err)
		}
		req.Header.Add(appIDHeaderKey, "test_app_extl_id")
		req.Header.Add(apiKeyHeaderKey, "test_app_api_key")

		testAppHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var a *diygoapi.App
			a, err = diygoapi.AppFromRequest(r)
			if err != nil {
				t.Fatal("app.FromRequest() error", err)
			}
			wantApp := &diygoapi.App{
				ID:          uuid.UUID{},
				ExternalID:  []byte("so random"),
				Org:         &diygoapi.Org{},
				Name:        "",
				Description: "",
				APIKeys:     nil,
			}
			c.Assert(a, qt.DeepEquals, wantApp)
		})

		rr := httptest.NewRecorder()

		lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)

		s := New(NewMuxRouter(), NewDriver(), lgr)
		s.AuthenticationServicer = mockAuthenticationService{}

		handlers := s.appHandler(testAppHandler)
		handlers.ServeHTTP(rr, req)

		// If there is any issues with the Access Token, the body
		// should be empty and the status code should be 401
		c.Assert(rr.Code, qt.Equals, http.StatusOK)
	})
}

func Test_parseAppHeader(t *testing.T) {
	t.Run("x-app-id", func(t *testing.T) {
		c := qt.New(t)
		hdr := http.Header{}
		hdr.Add(appIDHeaderKey, "appIdHeaderFakeText")

		appID, err := parseAppHeader(defaultRealm, hdr, appIDHeaderKey)
		c.Assert(err, qt.IsNil)
		c.Assert(appID, qt.Equals, "appIdHeaderFakeText")
	})
	t.Run("no header error", func(t *testing.T) {
		c := qt.New(t)
		hdr := http.Header{}

		_, err := parseAppHeader(defaultRealm, hdr, appIDHeaderKey)
		c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), errs.E(errs.NotExist, errs.Realm(defaultRealm), fmt.Sprintf("no %s header sent", appIDHeaderKey)))
	})
	t.Run("too many values error", func(t *testing.T) {
		c := qt.New(t)
		hdr := http.Header{}
		hdr.Add(appIDHeaderKey, "value1")
		hdr.Add(appIDHeaderKey, "value2")

		_, err := parseAppHeader(defaultRealm, hdr, appIDHeaderKey)
		c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), errs.E(errs.Unauthenticated, errs.Realm(defaultRealm), fmt.Sprintf("%s header value > 1", appIDHeaderKey)))
	})
	t.Run("empty value error", func(t *testing.T) {
		c := qt.New(t)
		hdr := http.Header{}
		hdr.Add(appIDHeaderKey, "")

		_, err := parseAppHeader(defaultRealm, hdr, appIDHeaderKey)
		c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), errs.E(errs.Unauthenticated, errs.Realm(defaultRealm), fmt.Sprintf("unauthenticated: %s header value not found", appIDHeaderKey)))
	})
}

func Test_parseAuthorizationHeader(t *testing.T) {
	c := qt.New(t)

	const reqHeader string = "Authorization"

	type args struct {
		realm  string
		header http.Header
	}

	hdr := http.Header{}
	hdr.Add(reqHeader, "Bearer foobarbbq")

	emptyHdr := http.Header{}
	emptyHdrErr := errs.E(errs.Unauthenticated, errs.Realm(defaultRealm), "unauthenticated: no Authorization header sent")

	tooManyValues := http.Header{}
	tooManyValues.Add(reqHeader, "value1")
	tooManyValues.Add(reqHeader, "value2")
	tooManyValuesErr := errs.E(errs.Unauthenticated, errs.Realm(defaultRealm), "header value > 1")

	noBearer := http.Header{}
	noBearer.Add(reqHeader, "xyz")
	noBearerErr := errs.E(errs.Unauthenticated, errs.Realm(defaultRealm), "unauthenticated: Bearer authentication scheme not found")

	hdrSpacesBearer := http.Header{}
	hdrSpacesBearer.Add("Authorization", "Bearer  ")
	spacesHdrErr := errs.E(errs.Unauthenticated, errs.Realm(defaultRealm), "unauthenticated: Authorization header sent with Bearer scheme, but no token found")

	tests := []struct {
		name      string
		args      args
		wantToken oauth2.Token
		wantErr   error
	}{
		{"typical", args{realm: defaultRealm, header: hdr}, oauth2.Token{AccessToken: "foobarbbq", TokenType: diygoapi.BearerTokenType}, nil},
		{"no authorization header error", args{realm: defaultRealm, header: emptyHdr}, oauth2.Token{}, emptyHdrErr},
		{"too many values error", args{realm: defaultRealm, header: tooManyValues}, oauth2.Token{}, tooManyValuesErr},
		{"no bearer scheme error", args{realm: defaultRealm, header: noBearer}, oauth2.Token{}, noBearerErr},
		{"spaces as token error", args{realm: defaultRealm, header: hdrSpacesBearer}, oauth2.Token{}, spacesHdrErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := parseAuthorizationHeader(tt.args.realm, tt.args.header)
			if (err != nil) && (tt.wantErr == nil) {
				t.Errorf("authHeader() error = %v, nil expected", err)
				return
			}
			var gotToken oauth2.Token
			if token != nil {
				gotToken = *token
			}
			c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), tt.wantErr)
			c.Assert(gotToken, qt.Equals, tt.wantToken)
		})
	}
}
