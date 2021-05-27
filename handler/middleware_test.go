package handler

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

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

		handlers := mw.AccessTokenHandler(mw.DefaultRealmHandler(testAccessTokenHandler))
		handlers.ServeHTTP(rr, req)

		// If there is any issues with the Access Token, the body
		// should be empty and the status code should be 401
		c.Assert(rr.Code, qt.Equals, http.StatusOK)
	})

	t.Run("no token", func(t *testing.T) {
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

		handlers := mw.AccessTokenHandler(testAccessTokenHandler)
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
