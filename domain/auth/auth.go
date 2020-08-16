package auth

import (
	"context"
	"fmt"

	"github.com/gilcrest/errs"
	"github.com/gilcrest/go-api-basic/domain/user"
)

func AuthorizeUser(ctx context.Context, u *user.User) error {
	const op errs.Op = "domain/auth/AuthorizeUser"

	m := make(map[string]bool)

	m["gilcrest@gmail.com"] = true

	authorized, exists := m[u.Email]

	// "In summary, a 401 Unauthorized response should be used for missing or
	// bad authentication, and a 403 Forbidden response should be used afterwards,
	// when the user is authenticated but isn’t authorized to perform the
	// requested operation on the given resource."
	// If the user has gotten here, they have gotten through authentication
	// but do have the right access, this they are Unauthorized
	switch {
	case exists && !authorized:
		// User exists, but is not authorized
		return errs.E(op, errs.Unauthorized, fmt.Sprintf("user email %s exists in auth map, but not allowed", u.Email))
	case !exists:
		// User does not exist in authorization list
		return errs.E(op, errs.Unauthorized, fmt.Sprintf("user email %s does not exist in auth map", u.Email))
	}

	return nil
}
