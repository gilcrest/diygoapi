package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/gilcrest/diygoapi/errs"
)

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
		err = decoderErr(json.NewDecoder(r.Body).Decode(&wantBody))
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
		err = decoderErr(json.NewDecoder(r.Body).Decode(&wantBody))
		defer r.Body.Close()

		wantErr := errs.E(errs.InvalidRequest, errors.New("malformed JSON"))
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
		err = decoderErr(json.NewDecoder(r.Body).Decode(&wantBody))
		defer r.Body.Close()

		wantErr := errs.E(errs.InvalidRequest, errors.New("request body cannot be empty"))
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
		err = decoderErr(decoder.Decode(&wantBody))
		defer r.Body.Close()

		// check to make sure I have an error
		c.Assert(err != nil, qt.Equals, true)
	})
}
