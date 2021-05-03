package handler

import (
	"net/http"
	"strings"

	"github.com/justinas/alice"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/hlog"

	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/user"
)

// Middleware are a bundled set of all the application middlewares
type Middleware struct {
	JSONContentTypeResponseMw JSONContentTypeResponseMw
	AccessTokenMw             AccessTokenMw
	ConvertAccessTokenMw      ConvertAccessTokenMw
	AuthorizeUserMw           AuthorizeUserMw
}

// JSONContentTypeResponseMiddleware returns the JSON Content response middleware
func (mw Middleware) JSONContentTypeResponseMiddleware() alice.Constructor {
	return alice.Constructor(mw.JSONContentTypeResponseMw)
}

// AccessTokenMiddleware returns the access token middleware
func (mw Middleware) AccessTokenMiddleware() alice.Constructor {
	return alice.Constructor(mw.AccessTokenMw)
}

// ConvertAccessTokenMiddleware returns the Access Token converter middleware
func (mw Middleware) ConvertAccessTokenMiddleware() alice.Constructor {
	return alice.Constructor(mw.ConvertAccessTokenMw)
}

// AuthorizeUserMiddleware returns the Authorize User middleware
func (mw Middleware) AuthorizeUserMiddleware() alice.Constructor {
	return alice.Constructor(mw.AuthorizeUserMw)
}

// NewJSONContentTypeResponseMw is an initializer for JSONContentTypeResponseMw
func NewJSONContentTypeResponseMw() JSONContentTypeResponseMw {
	return JSONContentTypeResponseHandler
}

// JSONContentTypeResponseMw is a middleware to add a
// application/json Content-Type Header to responses
type JSONContentTypeResponseMw alice.Constructor

// NewAccessTokenMw is an initializer for AccessTokenMw
func NewAccessTokenMw() AccessTokenMw {
	return AccessTokenHandler
}

// AccessTokenMw is a middleware to pull the Bearer token
// from the Authorization header and set it to the request context
// as an auth.AccessToken
type AccessTokenMw alice.Constructor

// NewAuthorizeUserMw is an initializer for AuthorizeUserMw
func NewAuthorizeUserMw(authorizer auth.Authorizer) AuthorizeUserMw {
	return AuthorizeUserHandler(authorizer)
}

// AuthorizeUserMw middleware is used to authorize a User's access to
// a resource at a given request path and http method
type AuthorizeUserMw alice.Constructor

// NewConvertAccessTokenMw is an initializer for ConvertAccessTokenMw
func NewConvertAccessTokenMw(converter auth.AccessTokenConverter) ConvertAccessTokenMw {
	return ConvertAccessTokenHandler(converter)
}

// ConvertAccessTokenMw middleware is used to convert an access token
// to a User
type ConvertAccessTokenMw alice.Constructor

// JSONContentTypeResponseHandler middleware is used to add the
// application/json Content-Type Header for responses
func JSONContentTypeResponseHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add(contentTypeHeaderKey, appJSONContentTypeHeaderVal)
			h.ServeHTTP(w, r) // call original
		})
}

// AccessTokenHandler middleware is used to pull the Bearer token
// from the Authorization header and set it to the request context
// as an auth.AccessToken
func AccessTokenHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
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
func ConvertAccessTokenHandler(converter auth.AccessTokenConverter) (mw func(h http.Handler) http.Handler) {
	mw = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lgr := *hlog.FromRequest(r)

			// retrieve access token from Context
			accessToken, err := auth.AccessTokenFromRequest(r)
			if err != nil {
				errs.HTTPErrorResponse(w, lgr, err)
				return
			}

			// convert access token to User
			u, err := converter.Convert(r.Context(), accessToken)
			if err != nil {
				errs.HTTPErrorResponse(w, lgr, err)
				return
			}

			// add User to context
			ctx := user.CtxWithUser(r.Context(), u)

			h.ServeHTTP(w, r.WithContext(ctx)) // call original
		})
	}
	return
}

// AuthorizeUserHandler middleware is used authorize a User for a request path and http method
func AuthorizeUserHandler(authorizer auth.Authorizer) (mw func(h http.Handler) http.Handler) {
	mw = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lgr := *hlog.FromRequest(r)

			// retrieve user from Context
			u, err := user.FromRequest(r)
			if err != nil {
				errs.HTTPErrorResponse(w, lgr, err)
				return
			}

			// convert access token to User
			err = authorizer.Authorize(r.Context(), u, r.URL.Path, r.Method)
			if err != nil {
				errs.HTTPErrorResponse(w, lgr, err)
				return
			}

			h.ServeHTTP(w, r) // call original
		})
	}
	return
}
