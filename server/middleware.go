package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/justinas/alice"
	"github.com/rs/zerolog/hlog"
	"golang.org/x/oauth2"

	"github.com/gilcrest/diy-go-api/domain/app"
	"github.com/gilcrest/diy-go-api/domain/audit"
	"github.com/gilcrest/diy-go-api/domain/auth"
	"github.com/gilcrest/diy-go-api/domain/errs"
	"github.com/gilcrest/diy-go-api/domain/user"
	"github.com/gilcrest/diy-go-api/service"
)

const (
	// App ID header key
	appIDHeaderKey string = "X-APP-ID"
	// API key header key
	apiKeyHeaderKey string = "X-API-KEY"
	// Authorization provider header key
	authProviderHeaderKey string = "X-AUTH-PROVIDER"
	// Default Realm used as part of the WWW-Authenticate response
	// header when returning a 401 Unauthorized response
	defaultRealm string = "diy-go-api"
)

// jsonContentTypeResponseHandler middleware is used to add the
// application/json Content-Type Header for responses
func (s *Server) jsonContentTypeResponseHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(contentTypeHeaderKey, appJSONContentTypeHeaderVal)
			h.ServeHTTP(w, r) // call original
		})
}

// appHandler middleware is used to parse the request app id and api key
// from the X-APP-ID and X-API-KEY headers, retrieve and validate
// their veracity, retrieve the App details from the datastore and
// finally set the App to the request context.
func (s *Server) appHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lgr := *hlog.FromRequest(r)

		// retrieve the context from the http.Request
		ctx := r.Context()

		var (
			appExtlID string
			err       error
		)
		appExtlID, err = xHeader(defaultRealm, r.Header, appIDHeaderKey)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		var apiKey string
		apiKey, err = xHeader(defaultRealm, r.Header, apiKeyHeaderKey)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		var a app.App
		a, err = s.MiddlewareService.FindAppByAPIKey(ctx, defaultRealm, appExtlID, apiKey)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		// add access token to context
		ctx = app.CtxWithApp(ctx, a)

		// call original, adding access token to request context
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// userHandler middleware is used to parse the request authorization
// provider and authorization headers (X-AUTH-PROVIDER + Authorization respectively),
// retrieve and validate their veracity, retrieve the User details from
// the Oauth2 provider as well as the datastore and finally set the User
// to the request context.
func (s *Server) userHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lgr := *hlog.FromRequest(r)

		// retrieve the context from the http.Request
		ctx := r.Context()

		u, err := newUser(ctx, s.MiddlewareService, r, true)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		// add User to context
		ctx = user.CtxWithUser(ctx, u)

		// call original, adding User to request context
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// newUserHandler middleware is used to parse the request authorization
// provider and authorization headers (X-AUTH-PROVIDER + Authorization respectively),
// retrieve and validate their veracity, retrieve the User details from
// the Oauth2 provider and finally set the User to the request context.
func (s *Server) newUserHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lgr := *hlog.FromRequest(r)

		// retrieve the context from the http.Request
		ctx := r.Context()

		u, err := newUser(ctx, s.MiddlewareService, r, false)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		// add User to context
		ctx = user.CtxWithUser(ctx, u)

		// call original, adding User to request context
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func newUser(ctx context.Context, s MiddlewareService, r *http.Request, retrieveFromDB bool) (user.User, error) {

	var (
		a   app.App
		err error
	)
	a, err = app.FromRequest(r)
	if err != nil {
		return user.User{}, err
	}

	var providerVal string
	providerVal, err = xHeader(defaultRealm, r.Header, authProviderHeaderKey)
	if err != nil {
		return user.User{}, err
	}
	provider := auth.ParseProvider(providerVal)

	var token oauth2.Token
	token, err = authHeader(defaultRealm, r.Header)
	if err != nil {
		return user.User{}, err
	}

	params := service.FindUserParams{
		Realm:          defaultRealm,
		App:            a,
		Provider:       provider,
		Token:          token,
		RetrieveFromDB: retrieveFromDB,
	}

	var u user.User
	u, err = s.FindUserByOauth2Token(ctx, params)
	if err != nil {
		return user.User{}, err
	}

	return u, nil
}

// authorizeUserHandler middleware is used authorize a User for a request path and http method
func (s *Server) authorizeUserHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lgr := *hlog.FromRequest(r)

		// retrieve user from request context
		adt, err := audit.FromRequest(r)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		// authorize user can access the path/method
		err = s.MiddlewareService.Authorize(lgr, r, adt)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		h.ServeHTTP(w, r) // call original
	})
}

// LoggerChain returns a middleware chain (via alice.Chain)
// initialized with all the standard middleware handlers for logging. The logger
// will be added to the request context for subsequent use with pre-populated
// fields, including the request method, url, status, size, duration, remote IP,
// user agent, referer. A unique Request ID is also added to the logger, context
// and response headers.
func (s *Server) loggerChain() alice.Chain {
	ac := alice.New(hlog.NewHandler(s.Logger),
		hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
			hlog.FromRequest(r).Info().
				Str("method", r.Method).
				Stringer("url", r.URL).
				Int("status", status).
				Int("size", size).
				Dur("duration", duration).
				Msg("request logged")
		}),
		hlog.RemoteAddrHandler("remote_ip"),
		hlog.UserAgentHandler("user_agent"),
		hlog.RefererHandler("referer"),
		hlog.RequestIDHandler("request_id", "Request-Id"),
	)

	return ac
}

// xHeader parses and returns the header value given the key. It is
// used to validate various header values as part of authentication
func xHeader(realm string, header http.Header, key string) (v string, err error) {
	// Pull the header value from the Header map given the key
	headerValue, ok := header[http.CanonicalHeaderKey(key)]
	if !ok {
		return "", errs.E(errs.Unauthenticated, errs.Realm(realm), fmt.Sprintf("unauthenticated: no %s header sent", key))

	}

	// too many values sent - should only be one value
	if len(headerValue) > 1 {
		return "", errs.E(errs.Unauthenticated, errs.Realm(realm), fmt.Sprintf("%s header value > 1", key))
	}

	// retrieve header value from map
	v = headerValue[0]

	// remove all leading/trailing white space
	v = strings.TrimSpace(v)

	// should not be empty
	if v == "" {
		return "", errs.E(errs.Unauthenticated, errs.Realm(realm), fmt.Sprintf("unauthenticated: %s header value not found", key))
	}

	return
}

// authHeader parses/validates the Authorization header and returns an Oauth2 token
func authHeader(realm string, header http.Header) (oauth2.Token, error) {
	// Pull the token from the Authorization header by retrieving the
	// value from the Header map with "Authorization" as the key
	//
	// format: Authorization: Bearer
	headerValue, ok := header["Authorization"]
	if !ok {
		return oauth2.Token{}, errs.E(errs.Unauthenticated, errs.Realm(realm), "unauthenticated: no Authorization header sent")
	}

	// too many values sent - spec allows for only one token
	if len(headerValue) > 1 {
		return oauth2.Token{}, errs.E(errs.Unauthenticated, errs.Realm(realm), "header value > 1")
	}

	// retrieve token from map
	token := headerValue[0]

	// Oauth2 should have "Bearer " as the prefix as the authentication scheme
	hasBearer := strings.HasPrefix(token, auth.BearerTokenType+" ")
	if !hasBearer {
		return oauth2.Token{}, errs.E(errs.Unauthenticated, errs.Realm(realm), "unauthenticated: Bearer authentication scheme not found")
	}

	// remove "Bearer " authentication scheme from header value
	token = strings.TrimPrefix(token, auth.BearerTokenType+" ")

	// remove all leading/trailing white space
	token = strings.TrimSpace(token)

	// token should not be empty
	if token == "" {
		return oauth2.Token{}, errs.E(errs.Unauthenticated, errs.Realm(realm), "unauthenticated: Authorization header sent with Bearer scheme, but no token found")
	}

	return oauth2.Token{AccessToken: token, TokenType: auth.BearerTokenType}, nil
}
