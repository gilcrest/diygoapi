package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"golang.org/x/oauth2"

	"github.com/gilcrest/go-api-basic/domain/app"
	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/org"
)

type mockFindAppService struct{}

func (m mockFindAppService) FindAppByAPIKey(ctx context.Context, realm string, appExtlID string, apiKey string) (app.App, error) {
	return app.App{
		ID:           uuid.UUID{},
		ExternalID:   []byte("so random"),
		Org:          org.Org{},
		Name:         "",
		Description:  "",
		CreateAppID:  uuid.UUID{},
		CreateUserID: uuid.UUID{},
		CreateTime:   time.Time{},
		UpdateAppID:  uuid.UUID{},
		UpdateUserID: uuid.UUID{},
		UpdateTime:   time.Time{},
		APIKeys:      nil,
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
			a, err := app.FromRequest(r)
			if err != nil {
				t.Fatal("app.FromRequest() error", err)
			}
			wantApp := app.App{
				ID:           uuid.UUID{},
				ExternalID:   []byte("so random"),
				Org:          org.Org{},
				Name:         "",
				Description:  "",
				CreateAppID:  uuid.UUID{},
				CreateUserID: uuid.UUID{},
				CreateTime:   time.Time{},
				UpdateAppID:  uuid.UUID{},
				UpdateUserID: uuid.UUID{},
				UpdateTime:   time.Time{},
				APIKeys:      nil,
			}
			c.Assert(a, qt.Equals, wantApp)
		})

		rr := httptest.NewRecorder()

		s := Server{}

		handlers := s.appHandler(testAppHandler)
		handlers.ServeHTTP(rr, req)

		// If there is any issues with the Access Token, the body
		// should be empty and the status code should be 401
		c.Assert(rr.Code, qt.Equals, http.StatusOK)
	})

	//// Authorization header is not added at all
	//t.Run("no auth header", func(t *testing.T) {
	//	c := qt.New(t)
	//
	//	req, err := http.NewRequest("GET", "/ping", nil)
	//	if err != nil {
	//		t.Fatalf("http.NewRequest() error = %v", err)
	//	}
	//
	//	testAccessTokenHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	//		c.Fatal("handler should not make it here")
	//	})
	//
	//	rr := httptest.NewRecorder()
	//
	//	s := Server{}
	//
	//	handlers := s.defaultRealmHandler(s.accessTokenHandler(testAccessTokenHandler))
	//	handlers.ServeHTTP(rr, req)
	//
	//	body, err := ioutil.ReadAll(rr.Body)
	//	if err != nil {
	//		t.Fatalf("ioutil.ReadAll() error = %v", err)
	//	}
	//
	//	// If there is any issues with the Access Token, the status
	//	// code should be 401, and the body should be empty
	//	c.Assert(rr.Code, qt.Equals, http.StatusUnauthorized)
	//	c.Assert(string(body), qt.Equals, "")
	//})
}

func TestXHeader(t *testing.T) {
	t.Run("x-app-id", func(t *testing.T) {
		c := qt.New(t)
		hdr := http.Header{}
		hdr.Add(appIDHeaderKey, "appologies")

		appID, err := xHeader(defaultRealm, hdr, appIDHeaderKey)
		c.Assert(err, qt.IsNil)
		c.Assert(appID, qt.Equals, "appologies")
	})
	t.Run("no header error", func(t *testing.T) {
		c := qt.New(t)
		hdr := http.Header{}

		_, err := xHeader(defaultRealm, hdr, appIDHeaderKey)
		c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), errs.E(errs.Unauthenticated, errs.Realm(defaultRealm), fmt.Sprintf("unauthenticated: no %s header sent", appIDHeaderKey)))
	})
	t.Run("too many values error", func(t *testing.T) {
		c := qt.New(t)
		hdr := http.Header{}
		hdr.Add(appIDHeaderKey, "value1")
		hdr.Add(appIDHeaderKey, "value2")

		_, err := xHeader(defaultRealm, hdr, appIDHeaderKey)
		c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), errs.E(errs.Unauthenticated, errs.Realm(defaultRealm), fmt.Sprintf("%s header value > 1", appIDHeaderKey)))
	})
	t.Run("empty value error", func(t *testing.T) {
		c := qt.New(t)
		hdr := http.Header{}
		hdr.Add(appIDHeaderKey, "")

		_, err := xHeader(defaultRealm, hdr, appIDHeaderKey)
		c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), errs.E(errs.Unauthenticated, errs.Realm(defaultRealm), fmt.Sprintf("unauthenticated: %s header value not found", appIDHeaderKey)))
	})
}

//func TestAccessTokenHandler(t *testing.T) {
//	t.Run("typical", func(t *testing.T) {
//		c := qt.New(t)
//
//		req, err := http.NewRequest("GET", "/ping", nil)
//		if err != nil {
//			t.Fatalf("http.NewRequest() error = %v", err)
//		}
//		req.Header.Add("Authorization", auth.BearerTokenType+" abcdef123")
//
//		testAccessTokenHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			token, ok := auth.AccessTokenFromRequest(r)
//			if !ok {
//				t.Fatal("auth.AccessTokenFromRequest() !ok")
//			}
//			wantToken := auth.AccessToken{
//				Token:     "abcdef123",
//				TokenType: auth.BearerTokenType,
//			}
//			c.Assert(token, qt.Equals, wantToken)
//		})
//
//		rr := httptest.NewRecorder()
//
//		s := Server{}
//
//		handlers := s.defaultRealmHandler(s.accessTokenHandler(testAccessTokenHandler))
//		handlers.ServeHTTP(rr, req)
//
//		// If there is any issues with the Access Token, the body
//		// should be empty and the status code should be 401
//		c.Assert(rr.Code, qt.Equals, http.StatusOK)
//	})
//
//	// Authorization header is not added at all
//	t.Run("no auth header", func(t *testing.T) {
//		c := qt.New(t)
//
//		req, err := http.NewRequest("GET", "/ping", nil)
//		if err != nil {
//			t.Fatalf("http.NewRequest() error = %v", err)
//		}
//
//		testAccessTokenHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			c.Fatal("handler should not make it here")
//		})
//
//		rr := httptest.NewRecorder()
//
//		s := Server{}
//
//		handlers := s.defaultRealmHandler(s.accessTokenHandler(testAccessTokenHandler))
//		handlers.ServeHTTP(rr, req)
//
//		body, err := ioutil.ReadAll(rr.Body)
//		if err != nil {
//			t.Fatalf("ioutil.ReadAll() error = %v", err)
//		}
//
//		// If there is any issues with the Access Token, the status
//		// code should be 401, and the body should be empty
//		c.Assert(rr.Code, qt.Equals, http.StatusUnauthorized)
//		c.Assert(string(body), qt.Equals, "")
//	})
//}

func Test_authHeader(t *testing.T) {
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
		{"typical", args{realm: defaultRealm, header: hdr}, oauth2.Token{AccessToken: "foobarbbq", TokenType: auth.BearerTokenType}, nil},
		{"no authorization header error", args{realm: defaultRealm, header: emptyHdr}, oauth2.Token{}, emptyHdrErr},
		{"too many values error", args{realm: defaultRealm, header: tooManyValues}, oauth2.Token{}, tooManyValuesErr},
		{"no bearer scheme error", args{realm: defaultRealm, header: noBearer}, oauth2.Token{}, noBearerErr},
		{"spaces as token error", args{realm: defaultRealm, header: hdrSpacesBearer}, oauth2.Token{}, spacesHdrErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := authHeader(tt.args.realm, tt.args.header)
			if (err != nil) && (tt.wantErr == nil) {
				t.Errorf("authHeader() error = %v, nil expected", err)
				return
			}
			c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), tt.wantErr)
			c.Assert(gotToken, qt.Equals, tt.wantToken)
		})
	}
}
