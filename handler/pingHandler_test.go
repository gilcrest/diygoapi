package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/datastore/pingstore"

	"github.com/rs/zerolog/hlog"

	qt "github.com/frankban/quicktest"

	"github.com/justinas/alice"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/datastore/datastoretest"
	"github.com/gilcrest/go-api-basic/domain/logger"
)

func TestDefaultPingHandler_Ping(t *testing.T) {
	t.Run("typical", func(t *testing.T) {
		if os.Getenv("NO_DB") == "true" {
			t.Skip("skipping db dependent test")
		}

		c := qt.New(t)
		var emptyBody []byte

		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)
		// initialize a sql.DB and cleanup function for it
		db, cleanup := datastoretest.NewDB(t)
		defer cleanup()
		ds := datastore.NewDefaultDatastore(db)
		pinger := pingstore.NewDefaultPinger(ds)
		dph := DefaultPingHandler{
			Pinger: pinger,
		}
		path := "/api/v1/ping"
		req := httptest.NewRequest(http.MethodGet, path, bytes.NewBuffer(emptyBody))
		rr := httptest.NewRecorder()
		var requestID string
		testMiddleware := func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rID, ok := hlog.IDFromRequest(r)
				if !ok {
					t.Fatal("Request ID not set to request context")
				}
				requestID = rID.String()

				h.ServeHTTP(w, r)
			})
		}

		pingHandler := NewPingHandler(dph)

		ac := alice.New()
		h := loggerHandlerChain(lgr, ac).
			Append(testMiddleware).
			Append(JSONContentTypeResponseHandler).
			Then(pingHandler)
		h.ServeHTTP(rr, req)

		type pingResponseData struct {
			DBUp bool `json:"db_up"`
		}
		wantBody := pingResponseData{DBUp: true}

		gotBody := pingResponseData{}
		err := DecoderErr(json.NewDecoder(rr.Result().Body).Decode(&gotBody))
		defer rr.Result().Body.Close()

		c.Assert(err, qt.IsNil)
		// Response Status Code should be 200
		c.Assert(rr.Code, qt.Equals, http.StatusOK)

		// Assert that the request ID has been added to the response header
		c.Assert(rr.Header().Get("Request-Id"), qt.Equals, requestID)

		// Assert that the response body equals the body we want
		c.Assert(gotBody, qt.DeepEquals, wantBody)
	})

	t.Run("mock", func(t *testing.T) {
		c := qt.New(t)
		var emptyBody []byte

		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)
		mp := mockPinger{}
		dph := DefaultPingHandler{
			Pinger: mp,
		}
		path := "/api/v1/ping"
		req := httptest.NewRequest(http.MethodGet, path, bytes.NewBuffer(emptyBody))
		rr := httptest.NewRecorder()
		var requestID string
		testMiddleware := func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rID, ok := hlog.IDFromRequest(r)
				if !ok {
					t.Fatal("Request ID not set to request context")
				}
				requestID = rID.String()

				h.ServeHTTP(w, r)
			})
		}

		pingHandler := NewPingHandler(dph)

		ac := alice.New()
		h := loggerHandlerChain(lgr, ac).
			Append(testMiddleware).
			Append(JSONContentTypeResponseHandler).
			Then(pingHandler)
		h.ServeHTTP(rr, req)

		type pingResponseData struct {
			DBUp bool `json:"db_up"`
		}
		wantBody := pingResponseData{DBUp: true}

		gotBody := pingResponseData{}
		err := DecoderErr(json.NewDecoder(rr.Result().Body).Decode(&gotBody))
		defer rr.Result().Body.Close()

		c.Assert(err, qt.IsNil)
		// Response Status Code should be 200
		c.Assert(rr.Code, qt.Equals, http.StatusOK)

		// Assert that the request ID has been added to the response header
		c.Assert(rr.Header().Get("Request-Id"), qt.Equals, requestID)

		// Assert that the response body equals the body we want
		c.Assert(gotBody, qt.DeepEquals, wantBody)
	})

}

type mockPinger struct{}

func (m mockPinger) PingDB(ctx context.Context) error {
	return nil
}
