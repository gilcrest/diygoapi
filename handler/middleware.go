package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/justinas/alice"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/hlog"

	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/user"
)

// Middleware are the collection of app middleware handlers
type Middleware struct {
	Logger               zerolog.Logger
	AccessTokenConverter auth.AccessTokenConverter
	Authorizer           auth.Authorizer
}

// JSONContentTypeResponseHandler middleware is used to add the
// application/json Content-Type Header for responses
func (mw Middleware) JSONContentTypeResponseHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(contentTypeHeaderKey, appJSONContentTypeHeaderVal)
			h.ServeHTTP(w, r) // call original
		})
}

// AccessTokenHandler middleware is used to pull the Bearer token
// from the Authorization header and set it to the request context
// as an auth.AccessToken
func (mw Middleware) AccessTokenHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lgr := *hlog.FromRequest(r)

		var token string

		// retrieve the context from the http.Request
		ctx := r.Context()

		// Pull the token from the Authorization header
		// by retrieving the value from the Header map with
		// "Authorization" as the key
		// format: Authorization: Bearer
		headerValue, ok := r.Header["Authorization"]
		if ok && len(headerValue) >= 1 {
			token = headerValue[0]
			token = strings.TrimPrefix(token, auth.BearerTokenType+" ")
		}

		// If the token is empty...
		if token == "" {
			// For Unauthenticated and Unauthorized errors,
			// the response body should be empty. Use logger
			// to log the error and then just send
			// http.StatusUnauthorized (401) or http.StatusForbidden (403)
			// depending on the circumstances. "In summary, a
			// 401 Unauthorized response should be used for missing or bad authentication,
			// and a 403 Forbidden response should be used afterwards, when the user is
			// authenticated but isn’t authorized to perform the requested operation on
			// the given resource."
			errs.HTTPErrorResponse(w, lgr, errs.E(errs.Unauthenticated, errors.New("Unauthenticated - empty Bearer token")))
			return
		}

		// add access token to context
		ctx = auth.CtxWithAccessToken(ctx, auth.NewAccessToken(token, auth.BearerTokenType))

		// call original, adding access token to request context
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ConvertAccessTokenHandler middleware is used to convert an
// AccessToken to a User and store the User to the request context
func (mw Middleware) ConvertAccessTokenHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lgr := *hlog.FromRequest(r)

		// retrieve access token from Context
		accessToken, err := auth.AccessTokenFromRequest(r)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		// convert access token to User
		u, err := mw.AccessTokenConverter.Convert(r.Context(), accessToken)
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
func (mw Middleware) AuthorizeUserHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lgr := *hlog.FromRequest(r)

		// retrieve user from Context
		u, err := user.FromRequest(r)
		if err != nil {
			errs.HTTPErrorResponse(w, lgr, err)
			return
		}

		// convert access token to User
		err = mw.Authorizer.Authorize(r.Context(), u, r.URL.Path, r.Method)
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
func (mw Middleware) LoggerChain() alice.Chain {
	ac := alice.New(hlog.NewHandler(mw.Logger),
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

// CtxWithUserChain chains handlers together to set the User
// to the Context
func (mw Middleware) CtxWithUserChain() alice.Chain {
	ac := alice.New(
		mw.AccessTokenHandler,
		mw.ConvertAccessTokenHandler,
	)

	return ac
}
