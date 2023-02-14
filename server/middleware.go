package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/justinas/alice"
	"github.com/rs/zerolog/hlog"

	"github.com/gilcrest/diygoapi"
	"github.com/gilcrest/diygoapi/errs"
)

const (
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

		a, err := s.AuthenticationServicer.FindAppByAPIKey(r, defaultRealm)
		if err != nil {
			var e *errs.Error
			if errors.As(err, &e) {
				if e.Kind != errs.NotExist {
					errs.HTTPErrorResponse(w, lgr, err)
					return
				}
				// using app authentication is optional
				// (app will be determined based on oauth2 token if not found here)
				// if errs.NotExist then call original, do not add anything to context
				h.ServeHTTP(w, r)
				return
			}
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

		auth, err := s.AuthenticationServicer.FindExistingAuth(r, defaultRealm)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		ctx := diygoapi.NewContextWithUser(r.Context(), auth.User)

		ctx, err = s.AuthenticationServicer.DetermineAppContext(ctx, auth, defaultRealm)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		// call original, with new context
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// authorizeUserHandler middleware is used to authorize a User for a
// request path and http method
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

		params, err := s.AuthenticationServicer.NewAuthenticationParams(r, defaultRealm)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		ctx = diygoapi.NewContextWithAuthParams(ctx, params)

		// call original, with new context
		h.ServeHTTP(w, r.WithContext(ctx))
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
