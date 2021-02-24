package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gilcrest/go-api-basic/domain/auth"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/justinas/alice"
	"github.com/rs/zerolog/hlog"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/datastore/datastoretest"
	"github.com/gilcrest/go-api-basic/datastore/moviestore"
	"github.com/gilcrest/go-api-basic/domain/auth/authtest"
	"github.com/gilcrest/go-api-basic/domain/logger"
)

func TestDefaultMovieHandlers_CreateMovie(t *testing.T) {
	t.Run("typical", func(t *testing.T) {
		// set environment variable NO_DB to true if you don't
		// have database connectivity and this test will be skipped
		if os.Getenv("NO_DB") == "true" {
			t.Skip("skipping db dependent test")
		}

		// initialize quickest checker
		c := qt.New(t)

		// initialize a zerolog Logger
		lgr := logger.NewLogger(os.Stdout, true)

		// initialize a sql.DB and cleanup function for it
		db, cleanup := datastoretest.NewDB(t, lgr)
		defer cleanup()

		// initialize DefaultDatastore
		ds := datastore.NewDefaultDatastore(db)

		// initialize the DefaultTransactor for the moviestore
		transactor := moviestore.NewDefaultTransactor(ds)

		// initialize the DefaultSelector for the moviestore
		selector := moviestore.NewDefaultSelector(ds)

		// initialize mockAccessTokenConverter
		mockAccessTokenConverter := authtest.NewMockAccessTokenConverter(t)

		// initialize DefaultMovieHandlers
		dmh := DefaultMovieHandlers{
			AccessTokenConverter: mockAccessTokenConverter,
			Authorizer:           authtest.NewMockAuthorizer(t),
			Transactor:           transactor,
			Selector:             selector,
		}

		// setup request body using anonymous struct
		requestBody := struct {
			Title    string `json:"title"`
			Rated    string `json:"rated"`
			Released string `json:"release_date"`
			RunTime  int    `json:"run_time"`
			Director string `json:"director"`
			Writer   string `json:"writer"`
		}{
			Title:    "Repo Man",
			Rated:    "R",
			Released: "1984-03-02T00:00:00Z",
			RunTime:  92,
			Director: "Alex Cox",
			Writer:   "Alex Cox",
		}

		// encode request body into buffer variable
		var buf bytes.Buffer
		err := json.NewEncoder(&buf).Encode(requestBody)
		if err != nil {
			t.Fatalf("Encode() error = %v", err)
		}

		// setup path
		path := pathPrefix + moviesV1PathRoot

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

		// retrieve createMovieHandler HTTP handler
		createMovieHandler := ProvideCreateMovieHandler(dmh)

		// initialize ResponseRecorder to use with ServeHTTP as it
		// satisfies ResponseWriter interface and records the response
		// for testing
		rr := httptest.NewRecorder()

		// initialize alice Chain to chain middleware
		ac := alice.New()

		// setup full handler chain needed for request
		h := LoggerHandlerChain(lgr, ac).
			Append(AccessTokenHandler).
			Append(JSONContentTypeHandler).
			Append(requestIDMiddleware).
			Then(createMovieHandler)

		// call the handler ServeHTTP method to execute the request
		// and record the response
		h.ServeHTTP(rr, req)

		// Assert that Response Status Code equals 200 (StatusOK)
		c.Assert(rr.Code, qt.Equals, http.StatusOK)

		// createMovieResponse is the response struct for a Movie
		// the response struct is tucked inside the handler, so we
		// have to recreate it here
		type createMovieResponse struct {
			ExternalID      string `json:"external_id"`
			Title           string `json:"title"`
			Rated           string `json:"rated"`
			Released        string `json:"release_date"`
			RunTime         int    `json:"run_time"`
			Director        string `json:"director"`
			Writer          string `json:"writer"`
			CreateUsername  string `json:"create_username"`
			CreateTimestamp string `json:"create_timestamp"`
			UpdateUsername  string `json:"update_username"`
			UpdateTimestamp string `json:"update_timestamp"`
		}

		// standardResponse is the standard response struct used for
		// all response bodies, the Data field is actually an
		// interface{} in the real struct (handler.StandardResponse),
		// but it's easiest to decode to JSON using a proper struct
		// as below
		type standardResponse struct {
			Path      string              `json:"path"`
			RequestID string              `json:"request_id"`
			Data      createMovieResponse `json:"data"`
		}

		// retrieve the mock User that is used for testing
		u, _ := mockAccessTokenConverter.Convert(req.Context(), authtest.NewAccessToken(t))

		// setup the expected response data
		wantBody := standardResponse{
			Path:      path,
			RequestID: requestID,
			Data: createMovieResponse{
				ExternalID:      "",
				Title:           "Repo Man",
				Rated:           "R",
				Released:        "1984-03-02T00:00:00Z",
				RunTime:         92,
				Director:        "Alex Cox",
				Writer:          "Alex Cox",
				CreateUsername:  u.Email,
				CreateTimestamp: "",
				UpdateUsername:  u.Email,
				UpdateTimestamp: "",
			},
		}

		// initialize standardResponse
		gotBody := standardResponse{}

		// decode the response body into the standardResponse (gotBody)
		err = DecoderErr(json.NewDecoder(rr.Result().Body).Decode(&gotBody))
		defer rr.Result().Body.Close()

		// Assert that there is no error after decoding the response body
		c.Assert(err, qt.IsNil)

		// quicktest uses Google's cmp library for DeepEqual comparisons. It
		// has some great options included with it. Below is an example of
		// ignoring certain fields...
		ignoreFields := cmpopts.IgnoreFields(standardResponse{},
			"Data.ExternalID", "Data.CreateTimestamp", "Data.UpdateTimestamp")

		// Assert that the response body (gotBody) is as expected (wantBody).
		// This handler is interacting with the database and getting unique
		// values, so certain fields are skipped. The other subtest using a
		// mocked database call will be used to test equality on all fields.
		c.Assert(gotBody, qt.CmpEquals(ignoreFields), wantBody)
	})

}
