package service

import (
	"fmt"
	qt "github.com/frankban/quicktest"
	"github.com/gilcrest/diygoapi"
	"github.com/gilcrest/diygoapi/errs"
	"github.com/google/go-cmp/cmp"
	"golang.org/x/oauth2"
	"net/http"
	"testing"
)

func Test_parseAppHeader(t *testing.T) {
	const defaultRealm = "diy"

	t.Run("x-app-id", func(t *testing.T) {
		c := qt.New(t)
		hdr := http.Header{}
		hdr.Add(diygoapi.AppIDHeaderKey, "appIdHeaderFakeText")

		appID, err := parseAppHeader(defaultRealm, hdr, diygoapi.AppIDHeaderKey)
		c.Assert(err, qt.IsNil)
		c.Assert(appID, qt.Equals, "appIdHeaderFakeText")
	})
	t.Run("no header error", func(t *testing.T) {
		c := qt.New(t)
		hdr := http.Header{}

		_, err := parseAppHeader(defaultRealm, hdr, diygoapi.AppIDHeaderKey)
		c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), errs.E(errs.NotExist, errs.Realm(defaultRealm), fmt.Sprintf("no %s header sent", diygoapi.AppIDHeaderKey)))
	})
	t.Run("too many values error", func(t *testing.T) {
		c := qt.New(t)
		hdr := http.Header{}
		hdr.Add(diygoapi.AppIDHeaderKey, "value1")
		hdr.Add(diygoapi.AppIDHeaderKey, "value2")

		_, err := parseAppHeader(defaultRealm, hdr, diygoapi.AppIDHeaderKey)
		c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), errs.E(errs.Unauthenticated, errs.Realm(defaultRealm), fmt.Sprintf("%s header value > 1", diygoapi.AppIDHeaderKey)))
	})
	t.Run("empty value error", func(t *testing.T) {
		c := qt.New(t)
		hdr := http.Header{}
		hdr.Add(diygoapi.AppIDHeaderKey, "")

		_, err := parseAppHeader(defaultRealm, hdr, diygoapi.AppIDHeaderKey)
		c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), errs.E(errs.Unauthenticated, errs.Realm(defaultRealm), fmt.Sprintf("unauthenticated: %s header value not found", diygoapi.AppIDHeaderKey)))
	})
}

func Test_parseAuthorizationHeader(t *testing.T) {
	c := qt.New(t)

	const (
		reqHeader    = "Authorization"
		defaultRealm = "diy"
	)

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
