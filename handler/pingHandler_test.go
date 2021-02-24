package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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

		lgr := logger.NewLogger(os.Stdout, true)
		// initialize a sql.DB and cleanup function for it
		db, cleanup := datastoretest.NewDB(t, lgr)
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

		pingHandler := ProvidePingHandler(dph)

		ac := alice.New()
		h := LoggerHandlerChain(lgr, ac).
			Append(testMiddleware).
			Append(JSONContentTypeHandler).
			Then(pingHandler)
		h.ServeHTTP(rr, req)

		type pingResponseData struct {
			DBUp bool `json:"db_up"`
		}
		type standardResponse struct {
			Path      string           `json:"path"`
			RequestID string           `json:"request_id"`
			Data      pingResponseData `json:"data"`
		}
		prd := pingResponseData{DBUp: true}
		wantBody := &standardResponse{Path: path, RequestID: requestID, Data: prd}

		gotBody := new(standardResponse)
		err := DecoderErr(json.NewDecoder(rr.Result().Body).Decode(&gotBody))
		defer rr.Result().Body.Close()

		c.Assert(err, qt.IsNil)
		c.Assert(gotBody, qt.DeepEquals, wantBody)

		// Response Status Code should be 200
		c.Assert(rr.Code, qt.Equals, http.StatusOK)
	})

	t.Run("mock", func(t *testing.T) {
		c := qt.New(t)
		var emptyBody []byte

		lgr := logger.NewLogger(os.Stdout, true)
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

		pingHandler := ProvidePingHandler(dph)

		ac := alice.New()
		h := LoggerHandlerChain(lgr, ac).
			Append(testMiddleware).
			Append(JSONContentTypeHandler).
			Then(pingHandler)
		h.ServeHTTP(rr, req)

		type pingResponseData struct {
			DBUp bool `json:"db_up"`
		}
		type standardResponse struct {
			Path      string           `json:"path"`
			RequestID string           `json:"request_id"`
			Data      pingResponseData `json:"data"`
		}
		prd := pingResponseData{DBUp: true}
		wantBody := &standardResponse{Path: path, RequestID: requestID, Data: prd}

		gotBody := new(standardResponse)
		err := DecoderErr(json.NewDecoder(rr.Result().Body).Decode(&gotBody))
		defer rr.Result().Body.Close()

		c.Assert(err, qt.IsNil)
		c.Assert(gotBody, qt.DeepEquals, wantBody)

		// Response Status Code should be 200
		c.Assert(rr.Code, qt.Equals, http.StatusOK)
	})

}

type mockPinger struct{}

func (m mockPinger) PingDB(ctx context.Context) error {
	return nil
}
