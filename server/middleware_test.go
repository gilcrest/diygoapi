package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/gilcrest/diygoapi"
	"github.com/gilcrest/diygoapi/logger"
)

type mockAuthenticationService struct{}

func (s mockAuthenticationService) SelfRegister(ctx context.Context, params *diygoapi.AuthenticationParams) (ur *diygoapi.UserResponse, err error) {
	//TODO implement me
	panic("implement me")
}

func (s mockAuthenticationService) FindExistingAuth(r *http.Request, realm string) (diygoapi.Auth, error) {
	//TODO implement me
	panic("implement me")
}

func (s mockAuthenticationService) DetermineAppContext(ctx context.Context, auth diygoapi.Auth, realm string) (context.Context, error) {
	//TODO implement me
	panic("implement me")
}

func (s mockAuthenticationService) FindAppByAPIKey(r *http.Request, realm string) (*diygoapi.App, error) {
	return &diygoapi.App{
		ID:          uuid.UUID{},
		ExternalID:  []byte("so random"),
		Org:         &diygoapi.Org{},
		Name:        "",
		Description: "",
		APIKeys:     nil,
	}, nil
}

func (s mockAuthenticationService) AuthenticationParamExchange(ctx context.Context, params *diygoapi.AuthenticationParams) (*diygoapi.ProviderInfo, error) {
	//TODO implement me
	panic("implement me")
}

func (s mockAuthenticationService) NewAuthenticationParams(r *http.Request, realm string) (*diygoapi.AuthenticationParams, error) {
	//TODO implement me
	panic("implement me")
}

func (s mockAuthenticationService) FindAuth(ctx context.Context, params diygoapi.AuthenticationParams) (diygoapi.Auth, error) {
	panic("implement me")
}

func (s mockAuthenticationService) FindAppByProviderClientID(ctx context.Context, realm string, auth diygoapi.Auth) (a *diygoapi.App, err error) {
	panic("implement me")
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

// TODO - currently using mock - should use database test to actually query db. Requires quite a bit of data setup, but is appropriate and will get to this.
func TestServer_appHandler(t *testing.T) {
	t.Run("typical - mock database", func(t *testing.T) {
		c := qt.New(t)

		req, err := http.NewRequest("GET", "/ping", nil)
		if err != nil {
			t.Fatalf("http.NewRequest() error = %v", err)
		}
		req.Header.Add(diygoapi.AppIDHeaderKey, "test_app_extl_id")
		req.Header.Add(diygoapi.ApiKeyHeaderKey, "test_app_api_key")

		testAppHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var a *diygoapi.App
			a, err = diygoapi.AppFromRequest(r)
			if err != nil {
				t.Fatal("app.FromRequest() error", err)
			}
			wantApp := &diygoapi.App{
				ID:          uuid.UUID{},
				ExternalID:  []byte("so random"),
				Org:         &diygoapi.Org{},
				Name:        "",
				Description: "",
				APIKeys:     nil,
			}
			c.Assert(a, qt.DeepEquals, wantApp)
		})

		rr := httptest.NewRecorder()

		lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)

		s := New(NewMuxRouter(), NewDriver(), lgr)
		s.AuthenticationServicer = mockAuthenticationService{}

		handlers := s.appHandler(testAppHandler)
		handlers.ServeHTTP(rr, req)

		// If there is any issues with the Access Token, the body
		// should be empty and the status code should be 401
		c.Assert(rr.Code, qt.Equals, http.StatusOK)
	})
}
