package handler

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	qt "github.com/frankban/quicktest"

	"github.com/gilcrest/go-api-basic/domain/auth"
)

func TestJSONContentTypeResponseHandler(t *testing.T) {
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

	mw := Middleware{}

	handlers := mw.JSONContentTypeResponseHandler(testJSONContentTypeResponseHandler)
	handlers.ServeHTTP(rr, req)
}

func TestAccessTokenHandler(t *testing.T) {
	t.Run("typical", func(t *testing.T) {
		c := qt.New(t)

		req, err := http.NewRequest("GET", "/ping", nil)
		if err != nil {
			t.Fatalf("http.NewRequest() error = %v", err)
		}
		req.Header.Add("Authorization", auth.BearerTokenType+" abcdef123")

		testAccessTokenHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := auth.AccessTokenFromRequest(r)
			if !ok {
				t.Fatal("auth.AccessTokenFromRequest() !ok")
			}
			wantToken := auth.AccessToken{
				Token:     "abcdef123",
				TokenType: auth.BearerTokenType,
			}
			c.Assert(token, qt.Equals, wantToken)
		})

		rr := httptest.NewRecorder()

		mw := Middleware{}

		handlers := mw.DefaultRealmHandler(mw.AccessTokenHandler(testAccessTokenHandler))
		handlers.ServeHTTP(rr, req)

		// If there is any issues with the Access Token, the body
		// should be empty and the status code should be 401
		c.Assert(rr.Code, qt.Equals, http.StatusOK)
	})

	// Authorization header is not added at all
	t.Run("no auth header", func(t *testing.T) {
		c := qt.New(t)

		req, err := http.NewRequest("GET", "/ping", nil)
		if err != nil {
			t.Fatalf("http.NewRequest() error = %v", err)
		}

		testAccessTokenHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c.Fatal("handler should not make it here")
		})

		rr := httptest.NewRecorder()

		mw := Middleware{}

		handlers := mw.DefaultRealmHandler(mw.AccessTokenHandler(testAccessTokenHandler))
		handlers.ServeHTTP(rr, req)

		body, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Fatalf("ioutil.ReadAll() error = %v", err)
		}

		// If there is any issues with the Access Token, the status
		// code should be 401, and the body should be empty
		c.Assert(rr.Code, qt.Equals, http.StatusUnauthorized)
		c.Assert(string(body), qt.Equals, "")
	})
}

func Test_authHeader(t *testing.T) {
	c := qt.New(t)

	const realm auth.WWWAuthenticateRealm = "DeepInTheRealm"
	const reqHeader string = "Authorization"

	type args struct {
		realm  auth.WWWAuthenticateRealm
		header http.Header
	}

	hdr := http.Header{}
	hdr.Add(reqHeader, "Bearer booyah")

	emptyHdr := http.Header{}
	emptyHdrErr := errs.NewUnauthenticatedError(string(realm), errors.New("unauthenticated: no Authorization header sent"))

	noBearer := http.Header{}
	noBearer.Add(reqHeader, "xyz")
	noBearerErr := errs.NewUnauthenticatedError(string(realm), errors.New("unauthenticated: Bearer authentication scheme not found"))

	hdrSpacesBearer := http.Header{}
	hdrSpacesBearer.Add("Authorization", "Bearer  ")
	spacesHdrErr := errs.NewUnauthenticatedError(string(realm), errors.New("unauthenticated: Authorization header sent with Bearer scheme, but no token found"))

	tests := []struct {
		name      string
		args      args
		wantToken string
		wantErr   error
	}{
		{"typical", args{realm: realm, header: hdr}, "booyah", nil},
		{"no authorization header", args{realm: realm, header: emptyHdr}, "", emptyHdrErr},
		{"no bearer scheme", args{realm: realm, header: noBearer}, "", noBearerErr},
		{"spaces as token", args{realm: realm, header: hdrSpacesBearer}, "", spacesHdrErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotToken, err := authHeader(tt.args.realm, tt.args.header)
			if (err != nil) && (tt.wantErr == nil) {
				t.Errorf("authHeader() error = %v, nil expected", err)
				return
			}
			c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.MatchUnauthenticated)), tt.wantErr)
			c.Assert(gotToken, qt.Equals, tt.wantToken)
		})
	}
}
