package handler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog/hlog"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/pkg/errors"

	"github.com/gilcrest/go-api-basic/domain/user"

	qt "github.com/frankban/quicktest"

	"github.com/gilcrest/go-api-basic/domain/auth"
)

func TestJSONContentTypeHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/ping", nil)
	if err != nil {
		t.Fatalf("http.NewRequest() error = %v", err)
	}

	testJSONContentTypeHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Fatalf("Content-Type %s is invalid", contentType)
		}
	})

	rr := httptest.NewRecorder()

	handlers := JSONContentTypeHandler(testJSONContentTypeHandler)
	handlers.ServeHTTP(rr, req)
}

func TestAccessTokenHandler(t *testing.T) {
	t.Run("typical", func(t *testing.T) {
		c := qt.New(t)

		req, err := http.NewRequest("GET", "/ping", nil)
		if err != nil {
			t.Fatalf("http.NewRequest() error = %v", err)
		}
		req.Header.Add("Authorization", "Bearer abcdef123")

		testAccessTokenHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := auth.FromRequest(r)
			if err != nil {
				t.Fatalf("auth.FromRequest() error = %v", err)
			}
			wantToken := auth.AccessToken{
				Token:     "abcdef123",
				TokenType: "Bearer",
			}
			t.Log(token)
			c.Assert(token, qt.Equals, wantToken)
		})

		rr := httptest.NewRecorder()

		handlers := AccessTokenHandler(testAccessTokenHandler)
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

		handlers := AccessTokenHandler(testAccessTokenHandler)
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

func TestNewStandardResponse(t *testing.T) {
	t.Run("no request id", func(t *testing.T) {
		c := qt.New(t)

		req, err := http.NewRequest("GET", "/ping", nil)
		if err != nil {
			t.Fatalf("http.NewRequest() error = %v", err)
		}
		u := user.User{Email: "otto.maddox711@gmail.com"}
		_, err = NewStandardResponse(req, u)
		wantErr := errs.E(errors.New("request ID not properly set to request context"))

		c.Assert(errs.Match(err, wantErr), qt.Equals, true)

	})

	t.Run("typical", func(t *testing.T) {
		c := qt.New(t)

		req, err := http.NewRequest("GET", "/ping", nil)
		if err != nil {
			t.Fatalf("http.NewRequest() error = %v", err)
		}
		rr := httptest.NewRecorder()

		var requestID string
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rID, ok := hlog.IDFromRequest(r)
			if !ok {
				t.Fatal("Request ID not set to request context")
			}
			requestID = rID.String()
			req = r
		})

		h := hlog.RequestIDHandler("request_id", "Request-Id")(testHandler)
		h.ServeHTTP(rr, req)

		u := user.User{Email: "otto.maddox711@gmail.com"}
		sr, err := NewStandardResponse(req, u)
		if err != nil {
			t.Fatalf("NewStandardResponse() error = %v", err)
		}

		wantStandardResponse := &StandardResponse{Path: "/ping", RequestID: requestID, Data: u}

		c.Assert(sr, qt.DeepEquals, wantStandardResponse)

	})
}

func TestDecoderErr(t *testing.T) {
	t.Run("typical", func(t *testing.T) {
		c := qt.New(t)

		type testBody struct {
			Director string `json:"director"`
			Writer   string `json:"writer"`
		}

		requestBody := []byte(`{
				"director": "Alex Cox",
				"writer": "Alex Cox"
			}`)

		r, err := http.NewRequest(http.MethodPost, "/fake", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Fatalf("http.NewRequest() error = %v", err)
		}

		// Decode JSON HTTP request body into a Decoder type
		// and unmarshal that into the testBody struct. DecoderErr
		// wraps errors from Decode when body is nil, json is malformed
		// or any other error
		wantBody := new(testBody)
		err = DecoderErr(json.NewDecoder(r.Body).Decode(&wantBody))
		defer r.Body.Close()
		c.Assert(err, qt.IsNil)
	})

	t.Run("malformed JSON", func(t *testing.T) {
		c := qt.New(t)

		type testBody struct {
			Director string `json:"director"`
			Writer   string `json:"writer"`
		}

		// removed trailing curly bracket
		requestBody := []byte(`{
				"director": "Alex Cox",
				"writer": "Alex Cox"`)

		r, err := http.NewRequest(http.MethodPost, "/fake", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Fatalf("http.NewRequest() error = %v", err)
		}

		// Decode JSON HTTP request body into a Decoder type
		// and unmarshal that into the testBody struct. DecoderErr
		// wraps errors from Decode when body is nil, JSON is malformed
		// or any other error
		wantBody := new(testBody)
		err = DecoderErr(json.NewDecoder(r.Body).Decode(&wantBody))
		defer r.Body.Close()

		wantErr := errs.E(errs.InvalidRequest, errors.New("Malformed JSON"))
		c.Assert(errs.Match(err, wantErr), qt.IsTrue)
	})

	t.Run("empty request body", func(t *testing.T) {
		c := qt.New(t)

		type testBody struct {
			Director string `json:"director"`
			Writer   string `json:"writer"`
		}

		// empty body
		requestBody := []byte("")

		r, err := http.NewRequest(http.MethodPost, "/fake", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Fatalf("http.NewRequest() error = %v", err)
		}

		// Decode JSON HTTP request body into a Decoder type
		// and unmarshal that into the testBody struct. DecoderErr
		// wraps errors from Decode when body is nil, JSON is malformed
		// or any other error
		wantBody := new(testBody)
		err = DecoderErr(json.NewDecoder(r.Body).Decode(&wantBody))
		defer r.Body.Close()

		wantErr := errs.E(errs.InvalidRequest, errors.New("Request Body cannot be empty"))
		c.Assert(errs.Match(err, wantErr), qt.IsTrue)
	})

	t.Run("invalid request body", func(t *testing.T) {
		c := qt.New(t)

		type testBody struct {
			Director string `json:"director"`
			Writer   string `json:"writer"`
		}

		// has unknown field
		requestBody := []byte(`{
				"director": "Alex Cox",
				"writer": "Alex Cox",
                "unknown_field": "I should fail"
			}`)

		r, err := http.NewRequest(http.MethodPost, "/fake", bytes.NewBuffer(requestBody))
		if err != nil {
			t.Fatalf("http.NewRequest() error = %v", err)
		}

		// force an error with DisallowUnknownFields
		wantBody := new(testBody)
		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err = DecoderErr(decoder.Decode(&wantBody))
		defer r.Body.Close()

		// check to make sure I have an error
		c.Assert(err != nil, qt.Equals, true)
	})
}
