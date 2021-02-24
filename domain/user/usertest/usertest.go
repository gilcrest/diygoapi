// Package usertest provides testing helper functions for the
// user package
package usertest

import (
	"testing"

	"github.com/gilcrest/go-api-basic/domain/user"
)

// NewUser provides a User for testing
func NewUser(t *testing.T) user.User {
	t.Helper()

	return user.User{Email: "otto.maddox711@gmail.com",
		LastName:  "Maddox",
		FirstName: "Otto",
		FullName:  "Otto Maddox",
	}
}
