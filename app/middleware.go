package app

import (
	"net/http"
	"strings"
	"time"

	"github.com/justinas/alice"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/hlog"

	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/user"
)

// JSONContentTypeResponseHandler middleware is used to add the
// application/json Content-Type Header for responses
func (s *Server) jsonContentTypeResponseHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(contentTypeHeaderKey, appJSONContentTypeHeaderVal)
			h.ServeHTTP(w, r) // call original
		})
}

// defaultRealmHandler middleware is used to set a default Realm to
// the request context
func (s *Server) defaultRealmHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// retrieve the context from the http.Request
		ctx := r.Context()

		// add realm to context
		ctx = auth.CtxWithRealm(ctx, auth.DefaultRealm)

		// call original, adding realm to request context
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AccessTokenHandler middleware is used to pull the Bearer token
// from the Authorization header and set it to the request context
// as an auth.AccessToken
func (s *Server) accessTokenHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lgr := *hlog.FromRequest(r)

		// The realm must be set to the request context in order to
		// properly send the WWW-Authenticate error in case of unauthorized
		// access attempts
		realm, ok := auth.RealmFromRequest(r)
		if !ok {
			errs.HTTPErrorResponse(w, lgr, errs.E(errs.Internal, "Realm not set properly to context"))
			return
		}
		if realm == "" {
			errs.HTTPErrorResponse(w, lgr, errs.E(errs.Internal, "Realm empty in context"))
			return
		}

		token, err := authHeader(realm, r.Header)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		// retrieve the context from the http.Request
		ctx := r.Context()

		// add access token to context
		ctx = auth.CtxWithAccessToken(ctx, auth.NewAccessToken(token, auth.BearerTokenType))

		// call original, adding access token to request context
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// authHeader parses and validates the Authorization header
func authHeader(realm auth.WWWAuthenticateRealm, header http.Header) (token string, err error) {
	// Pull the token from the Authorization header by retrieving the
	// value from the Header map with "Authorization" as the key
	//
	// format: Authorization: Bearer
	headerValue, ok := header["Authorization"]
	if !ok {
		return "", errs.NewUnauthenticatedError(string(realm), errors.New("unauthenticated: no Authorization header sent"))
	}

	// too many values sent - spec allows for only one token
	if len(headerValue) > 1 {
		return "", errs.NewUnauthenticatedError(string(realm), errors.New("header value > 1"))
	}

	// retrieve token from map
	token = headerValue[0]

	// Oauth2 should have "Bearer " as the prefix as the authentication scheme
	hasBearer := strings.HasPrefix(token, auth.BearerTokenType+" ")
	if !hasBearer {
		return "", errs.NewUnauthenticatedError(string(realm), errors.New("unauthenticated: Bearer authentication scheme not found"))
	}

	// remove "Bearer " authentication scheme from header value
	token = strings.TrimPrefix(token, auth.BearerTokenType+" ")

	// remove all leading/trailing white space
	token = strings.TrimSpace(token)

	// token should not be empty
	if token == "" {
		return "", errs.NewUnauthenticatedError(string(realm), errors.New("unauthenticated: Authorization header sent with Bearer scheme, but no token found"))
	}

	return
}

// ConvertAccessTokenHandler middleware is used to convert an
// AccessToken to a User and store the User to the request context
func (s *Server) convertAccessTokenHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lgr := *hlog.FromRequest(r)

		// retrieve access token from Context
		accessToken, ok := auth.AccessTokenFromRequest(r)
		if !ok {
			errs.HTTPErrorResponse(w, lgr, errs.E("Access Token not set properly to context"))
		}
		if accessToken.Token == "" {
			errs.HTTPErrorResponse(w, lgr, errs.E("Access Token empty in context"))
		}

		// convert access token to User
		u, err := s.AccessTokenConverter.Convert(r.Context(), accessToken)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		// add User to context
		ctx := user.CtxWithUser(r.Context(), u)

		// call original, adding User to request context
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AuthorizeUserHandler middleware is used authorize a User for a request path and http method
func (s *Server) authorizeUserHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lgr := *hlog.FromRequest(r)

		// retrieve user from request context
		u, err := user.FromRequest(r)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		// authorize user can access the path/method
		err = s.Authorizer.Authorize(lgr, u, r.URL.Path, r.Method)
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
	ac := alice.New(hlog.NewHandler(s.logger),
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

// CtxWithUserChain chains handlers together to set the Realm, Access
// Token and User to the Context
func (s *Server) ctxWithUserChain() alice.Chain {
	ac := alice.New(
		s.defaultRealmHandler,
		s.accessTokenHandler,
		s.convertAccessTokenHandler,
	)

	return ac
}
