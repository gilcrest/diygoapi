package handler

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/zerolog/hlog"

	"github.com/pkg/errors"

	"github.com/gilcrest/go-api-basic/domain/errs"
)

type contextKey string

const contextKeyAccessToken = contextKey("access-token")

// accessToken gets the access token from the request
func accessToken(r *http.Request) (string, error) {

	// retrieve the context from the http.Request
	ctx := r.Context()

	tokenStr, ok := ctx.Value(contextKeyAccessToken).(string)
	if !ok {
		return "", errs.E(errs.Unauthenticated, errors.New("Access Token not set properly to context"))
	}
	if tokenStr == "" {
		return "", errs.E(errs.Unauthenticated, errors.New("Access Token empty in context"))
	}
	return tokenStr, nil
}

// AccessTokenHandler middleware is used to set the Google access
// token to the context
func (ah *AppHandler) AccessTokenHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			logger := *hlog.FromRequest(r)

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
				token = strings.TrimPrefix(token, "Bearer ")
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
				errs.HTTPErrorResponse(w, logger, errs.E(errs.Unauthenticated, errors.New("Unauthenticated - empty Bearer token")))
				return
			}

			// add access token to context
			ctx = context.WithValue(ctx, contextKeyAccessToken, token)

			// call original, adding access token to request context
			h.ServeHTTP(w, r.WithContext(ctx))
		})
}
