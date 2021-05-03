package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/auth/authtest"
	"github.com/gilcrest/go-api-basic/domain/logger"

	qt "github.com/frankban/quicktest"
	"github.com/gorilla/mux"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func TestDefaultLoggerHandlers_ReadLogger(t *testing.T) {
	t.Run("typical", func(t *testing.T) {
		// initialize quickest checker
		c := qt.New(t)

		// initialize a zerolog Logger
		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)

		// set global logging level to Info
		zerolog.SetGlobalLevel(zerolog.InfoLevel)

		// set Error stack trace to true
		logger.WriteErrorStackGlobal(true)

		// initialize mockAccessTokenConverter
		mockAccessTokenConverter := authtest.NewMockAccessTokenConverter(t)

		// use default authorizer
		defaultAuthorizer := auth.DefaultAuthorizer{}

		// initialize DefaultMovieHandlers
		dlh := DefaultLoggerHandlers{
			AccessTokenConverter: mockAccessTokenConverter,
			Authorizer:           defaultAuthorizer,
		}

		// setup path
		path := pathPrefix + loggerV1PathRoot

		// form request using httptest
		req := httptest.NewRequest(http.MethodGet, path, nil)

		// add test access token
		req.Header.Add("Authorization", auth.BearerTokenType+" abc123def1")

		// create middleware to extract the request ID from
		// the request context for testing comparison
		var requestID string
		requestIDMiddleware := func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rID, ok := hlog.IDFromRequest(r)
				if !ok {
					t.Fatal("Request ID not set to request context")
				}
				requestID = rID.String()

				h.ServeHTTP(w, r)
			})
		}

		// retrieve ReadLoggerHandler HTTP handler
		readLoggerHandler := NewReadLoggerHandler(dlh)

		// initialize ResponseRecorder to use with ServeHTTP as it
		// satisfies ResponseWriter interface and records the response
		// for testing
		rr := httptest.NewRecorder()

		// initialize alice Chain to chain middleware
		ac := alice.New()

		// setup full handler chain needed for request
		h := loggerHandlerChain(lgr, ac).
			Append(AccessTokenHandler).
			Append(JSONContentTypeResponseHandler).
			Append(requestIDMiddleware).
			Then(readLoggerHandler)

		// handler needs path variable, so we need to use mux router
		router := mux.NewRouter()
		// setup the expected path and route variable
		router.Handle(pathPrefix+loggerV1PathRoot, h)
		// call the router ServeHTTP method to execute the request
		// and record the response
		router.ServeHTTP(rr, req)

		// Assert that Response Status Code equals 200 (StatusOK)
		c.Assert(rr.Code, qt.Equals, http.StatusOK)

		// Assert that the request ID has been added to the response header
		c.Assert(rr.Header().Get("Request-Id"), qt.Equals, requestID)

		// readLoggerResponse is the response struct for the logger.
		// The response struct is tucked inside the handler, so we
		// have to recreate it here
		type readLoggerResponse struct {
			LoggerMinimumLevel string `json:"logger_minimum_level,omitempty"`
			GlobalLogLevel     string `json:"global_log_level,omitempty"`
			LogErrorStack      bool   `json:"log_error_stack,omitempty"`
		}

		// setup the expected response data
		wantBody := readLoggerResponse{
			LoggerMinimumLevel: zerolog.DebugLevel.String(),
			GlobalLogLevel:     zerolog.InfoLevel.String(),
			LogErrorStack:      true,
		}

		// initialize readLoggerResponse
		gotBody := readLoggerResponse{}

		// decode the response body into gotBody
		err := DecoderErr(json.NewDecoder(rr.Result().Body).Decode(&gotBody))
		defer rr.Result().Body.Close()

		// Assert that there is no error after decoding the response body
		c.Assert(err, qt.IsNil)

		// Assert that the response body (gotBody) is as expected (wantBody).
		c.Assert(gotBody, qt.DeepEquals, wantBody)
	})
}

func TestDefaultLoggerHandlers_UpdateLogger(t *testing.T) {
	t.Run("typical", func(t *testing.T) {
		// initialize quickest checker
		c := qt.New(t)

		// initialize a zerolog Logger
		lgr := logger.NewLogger(os.Stdout, zerolog.TraceLevel, true)

		// set global logging level to Info
		zerolog.SetGlobalLevel(zerolog.InfoLevel)

		// set Error stack to true
		logger.WriteErrorStackGlobal(false)

		t.Logf("Minimum accepted log level set to %s", lgr.GetLevel().String())
		t.Logf("Initial global log level set to %s", zerolog.GlobalLevel())
		var logErrorStack bool
		if zerolog.ErrorStackMarshaler != nil {
			logErrorStack = true
		}
		t.Logf("Initial Write Error Stack global set to %t", logErrorStack)

		// initialize mockAccessTokenConverter
		mockAccessTokenConverter := authtest.NewMockAccessTokenConverter(t)

		// use default authorizer
		defaultAuthorizer := auth.DefaultAuthorizer{}

		// initialize DefaultMovieHandlers
		dlh := DefaultLoggerHandlers{
			AccessTokenConverter: mockAccessTokenConverter,
			Authorizer:           defaultAuthorizer,
		}

		// setup request body using anonymous struct
		requestBody := struct {
			GlobalLogLevel string `json:"global_log_level,omitempty"`
			LogErrorStack  string `json:"log_error_stack,omitempty"`
		}{
			GlobalLogLevel: "debug",
			LogErrorStack:  "true",
		}

		// encode request body into buffer variable
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(requestBody)
		if err != nil {
			t.Fatalf("Encode() error = %v", err)
		}

		// setup path
		path := pathPrefix + loggerV1PathRoot

		// form request using httptest
		req := httptest.NewRequest(http.MethodPost, path, &buf)

		// add test access token
		req.Header.Add("Authorization", auth.BearerTokenType+" abc123def1")

		// create middleware to extract the request ID from
		// the request context for testing comparison
		var requestID string
		requestIDMiddleware := func(h http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rID, ok := hlog.IDFromRequest(r)
				if !ok {
					t.Fatal("Request ID not set to request context")
				}
				requestID = rID.String()

				h.ServeHTTP(w, r)
			})
		}

		// retrieve ReadLoggerHandler HTTP handler
		updateLoggerHandler := NewUpdateLoggerHandler(dlh)

		// initialize ResponseRecorder to use with ServeHTTP as it
		// satisfies ResponseWriter interface and records the response
		// for testing
		rr := httptest.NewRecorder()

		// initialize alice Chain to chain middleware
		ac := alice.New()

		// setup full handler chain needed for request
		h := loggerHandlerChain(lgr, ac).
			Append(AccessTokenHandler).
			Append(JSONContentTypeResponseHandler).
			Append(requestIDMiddleware).
			Then(updateLoggerHandler)

		// handler needs path variable, so we need to use mux router
		router := mux.NewRouter()
		// setup the expected path and route variable
		router.Handle(pathPrefix+loggerV1PathRoot, h)
		// call the router ServeHTTP method to execute the request
		// and record the response
		router.ServeHTTP(rr, req)

		// Assert that Response Status Code equals 200 (StatusOK)
		c.Assert(rr.Code, qt.Equals, http.StatusOK)

		// Assert that the request ID has been added to the response header
		c.Assert(rr.Header().Get("Request-Id"), qt.Equals, requestID)

		// readLoggerResponse is the response struct for the logger.
		// The response struct is tucked inside the handler, so we
		// have to recreate it here
		type readLoggerResponse struct {
			LoggerMinimumLevel string `json:"logger_minimum_level,omitempty"`
			GlobalLogLevel     string `json:"global_log_level,omitempty"`
			LogErrorStack      bool   `json:"log_error_stack,omitempty"`
		}

		// setup the expected response data
		wantBody := readLoggerResponse{
			LoggerMinimumLevel: zerolog.TraceLevel.String(),
			GlobalLogLevel:     zerolog.DebugLevel.String(),
			LogErrorStack:      true,
		}

		// initialize readLoggerResponse
		gotBody := readLoggerResponse{}

		// decode the response body into gotBody
		err = DecoderErr(json.NewDecoder(rr.Result().Body).Decode(&gotBody))
		defer rr.Result().Body.Close()

		// Assert that there is no error after decoding the response body
		c.Assert(err, qt.IsNil)

		// Assert that the response body (gotBody) is as expected (wantBody).
		c.Assert(gotBody, qt.DeepEquals, wantBody)
	})
}
