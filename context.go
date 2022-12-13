package diygoapi

import (
	"context"
	"net/http"
	"time"

	"github.com/gilcrest/diygoapi/errs"
)

type contextKey string

const (
	appContextKey        = contextKey("app")
	contextKeyUser       = contextKey("user")
	authParamsContextKey = contextKey("authParams")
)

// NewContextWithApp returns a new context with the given App
func NewContextWithApp(ctx context.Context, a *App) context.Context {
	return context.WithValue(ctx, appContextKey, a)
}

// AppFromRequest is a helper function which returns the App from the
// request context.
func AppFromRequest(r *http.Request) (*App, error) {
	return AppFromContext(r.Context())
}

// AppFromContext returns the App from the given context
func AppFromContext(ctx context.Context) (*App, error) {
	a, ok := ctx.Value(appContextKey).(*App)
	if !ok {
		return a, errs.E(errs.NotExist, "App not set to context")
	}
	return a, nil
}

// NewContextWithUser returns a new context with the given User
func NewContextWithUser(ctx context.Context, u *User) context.Context {
	return context.WithValue(ctx, contextKeyUser, u)
}

// UserFromRequest returns the User from the request context
func UserFromRequest(r *http.Request) (u *User, err error) {
	var ok bool
	u, ok = r.Context().Value(contextKeyUser).(*User)
	if !ok {
		return nil, errs.E(errs.Internal, "User not set properly to context")
	}
	if err = u.Validate(); err != nil {
		return nil, err
	}
	return u, nil
}

// AuditFromRequest is a convenience function that sets up an Audit
// struct from the App and User set to the request context.
// The moment is also set to time.Now
func AuditFromRequest(r *http.Request) (adt Audit, err error) {
	var a *App
	a, err = AppFromRequest(r)
	if err != nil {
		return Audit{}, err
	}

	var u *User
	u, err = UserFromRequest(r)
	if err != nil {
		return Audit{}, err
	}

	adt.App = a
	adt.User = u
	adt.Moment = time.Now()

	return adt, nil
}

// NewContextWithAuthParams returns a new context with the given AuthenticationParams
func NewContextWithAuthParams(ctx context.Context, ap *AuthenticationParams) context.Context {
	return context.WithValue(ctx, authParamsContextKey, ap)
}

// AuthParamsFromContext returns the AuthenticationParams from the given context
func AuthParamsFromContext(ctx context.Context) (*AuthenticationParams, error) {
	a, ok := ctx.Value(authParamsContextKey).(*AuthenticationParams)
	if !ok {
		return a, errs.E(errs.NotExist, "Authentication Params not set to context")
	}
	return a, nil
}
