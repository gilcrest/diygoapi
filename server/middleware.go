package server

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/justinas/alice"
	"github.com/rs/zerolog/hlog"
	"golang.org/x/oauth2"

	"github.com/gilcrest/diygoapi"
	"github.com/gilcrest/diygoapi/errs"
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
	defaultRealm string = "diy"
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
		appExtlID, err = parseAppHeader(defaultRealm, r.Header, appIDHeaderKey)
		if err != nil {
			var e *errs.Error
			if errors.As(err, &e) {
				if e.Kind != errs.NotExist {
					errs.HTTPErrorResponse(w, lgr, err)
					return
				}
				// using app authentication is optional, if errs.NotExist
				// then call original, do not add anything to context
				h.ServeHTTP(w, r)
				return
			}
			// should never get here, but just in case...
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		var apiKey string
		apiKey, err = parseAppHeader(defaultRealm, r.Header, apiKeyHeaderKey)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		var a *diygoapi.App
		a, err = s.AuthenticationServicer.FindAppByAPIKey(ctx, defaultRealm, appExtlID, apiKey)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		// get a new context with app added
		ctx = diygoapi.NewContextWithApp(ctx, a)

		lgr.Debug().Msgf("Internal app authentication successful for: %s", a.Name)

		// call original, adding app to request context
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// authHandler middleware is used to parse the request authentication
// provider and authorization Bearer token HTTP headers (X-AUTH-PROVIDER +
// Authorization respectively) and determine authentication. Authentication
// is determined by validating the App making the request as well as the User.
//
// authHandler does the following: searches for an existing authorized User:
//
// If an auth object already exists in the datastore for the bearer token
// and the bearer token is not past its expiration date, that auth will
// be used to determine the User.
//
// If no auth object exists in the datastore for the bearer token, an attempt
// will be made to find the user's auth with the provider id and unique ID
// given by the provider (found by calling the provider API). If an auth
// object exists given these attributes, it will be updated with the
// new bearer token details.
//
// authHandler also ensures an App is authenticated:
//
// If there is no app already in the request Context, then an app must be
// found given the User's Oauth2 Provider's Client ID. This app will be set
// to the request Context. If an app has already been authenticated as part
// of an upstream middleware and set to the request Context, then the Oauth2
// Provider's Client ID for the given User is not considered.
func (s *Server) authHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lgr := *hlog.FromRequest(r)

		// retrieve the context from the http.Request
		ctx := r.Context()

		var (
			provider diygoapi.Provider
			err      error
		)
		provider, err = parseProviderHeader(defaultRealm, r.Header)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		var token *oauth2.Token
		token, err = parseAuthorizationHeader(defaultRealm, r.Header)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		params := diygoapi.AuthenticationParams{
			Realm:    defaultRealm,
			Provider: provider,
			Token:    token,
		}

		var auth diygoapi.Auth
		auth, err = s.AuthenticationServicer.FindAuth(ctx, params)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		ctx = diygoapi.NewContextWithUser(ctx, auth.User)

		_, err = diygoapi.AppFromRequest(r)
		if err != nil {
			// no app found in request, lookup app from Auth
			var a *diygoapi.App
			a, err = s.AuthenticationServicer.FindAppByProviderClientID(ctx, defaultRealm, auth)
			if err != nil {
				errs.HTTPErrorResponse(w, lgr, err)
				return
			}
			// get a new context with App from Auth added to it
			ctx = diygoapi.NewContextWithApp(ctx, a)
		}

		// call original, with new context
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// authorizeUserHandler middleware is used authorize a User for a request path and http method
func (s *Server) authorizeUserHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lgr := *hlog.FromRequest(r)

		// retrieve user from request context
		adt, err := diygoapi.AuditFromRequest(r)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		// authorize user can access the path/method
		err = s.AuthorizationServicer.Authorize(r, lgr, adt)
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

// parseAppHeader parses an app header and returns its value.
func parseAppHeader(realm string, header http.Header, key string) (v string, err error) {
	// Pull the header value from the Header map given the key
	headerValue, ok := header[http.CanonicalHeaderKey(key)]
	if !ok {
		return "", errs.E(errs.NotExist, errs.Realm(realm), fmt.Sprintf("no %s header sent", key))

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

	return v, nil
}

// parseProviderHeader parses the X-AUTH-PROVIDER header and returns its value.
func parseProviderHeader(realm string, header http.Header) (p diygoapi.Provider, err error) {
	// Pull the header value from the Header map given the key
	headerValue, ok := header[http.CanonicalHeaderKey(authProviderHeaderKey)]
	if !ok {
		return diygoapi.UnknownProvider, errs.E(errs.Unauthenticated, errs.Realm(realm), fmt.Sprintf("no %s header sent", authProviderHeaderKey))

	}

	// too many values sent - should only be one value
	if len(headerValue) > 1 {
		return diygoapi.UnknownProvider, errs.E(errs.Unauthenticated, errs.Realm(realm), fmt.Sprintf("%s header value > 1", authProviderHeaderKey))
	}

	// retrieve header value from map
	v := headerValue[0]

	// remove all leading/trailing white space
	v = strings.TrimSpace(v)

	// should not be empty
	if v == "" {
		return diygoapi.UnknownProvider, errs.E(errs.Unauthenticated, errs.Realm(realm), fmt.Sprintf("unauthenticated: %s header value not found", authProviderHeaderKey))
	}

	p = diygoapi.ParseProvider(v)

	if p == diygoapi.UnknownProvider {
		return p, errs.E(errs.Unauthenticated, errs.Realm(realm), fmt.Sprintf("unknown provider given: %s", v))
	}

	return p, nil
}

// parseAuthorizationHeader parses/validates the Authorization header and returns an Oauth2 token
func parseAuthorizationHeader(realm string, header http.Header) (*oauth2.Token, error) {
	// Pull the token from the Authorization header by retrieving the
	// value from the Header map with "Authorization" as the key
	//
	// format: Authorization: Bearer
	headerValue, ok := header["Authorization"]
	if !ok {
		return nil, errs.E(errs.Unauthenticated, errs.Realm(realm), "unauthenticated: no Authorization header sent")
	}

	// too many values sent - spec allows for only one token
	if len(headerValue) > 1 {
		return nil, errs.E(errs.Unauthenticated, errs.Realm(realm), "header value > 1")
	}

	// retrieve token from map
	token := headerValue[0]

	// Oauth2 should have "Bearer " as the prefix as the authentication scheme
	hasBearer := strings.HasPrefix(token, diygoapi.BearerTokenType+" ")
	if !hasBearer {
		return nil, errs.E(errs.Unauthenticated, errs.Realm(realm), "unauthenticated: Bearer authentication scheme not found")
	}

	// remove "Bearer " authentication scheme from header value
	token = strings.TrimPrefix(token, diygoapi.BearerTokenType+" ")

	// remove all leading/trailing white space
	token = strings.TrimSpace(token)

	// token should not be empty
	if token == "" {
		return nil, errs.E(errs.Unauthenticated, errs.Realm(realm), "unauthenticated: Authorization header sent with Bearer scheme, but no token found")
	}

	return &oauth2.Token{AccessToken: token, TokenType: diygoapi.BearerTokenType}, nil
}

// genesisAuthHandler middleware is used to parse the request authentication
// provider and authorization Bearer token HTTP headers (X-AUTH-PROVIDER +
// Authorization respectively) and determine authentication. Authentication
// is determined only by validating the User (Genesis is a one-time startup event,
// so App authentication is not possible yet).
func (s *Server) genesisAuthHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lgr := *hlog.FromRequest(r)

		// retrieve the context from the http.Request
		ctx := r.Context()

		var (
			provider diygoapi.Provider
			err      error
		)
		provider, err = parseProviderHeader(defaultRealm, r.Header)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		var token *oauth2.Token
		token, err = parseAuthorizationHeader(defaultRealm, r.Header)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		params := &diygoapi.AuthenticationParams{
			Realm:    defaultRealm,
			Provider: provider,
			Token:    token,
		}

		ctx = diygoapi.NewContextWithAuthParams(ctx, params)

		// call original, with new context
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
