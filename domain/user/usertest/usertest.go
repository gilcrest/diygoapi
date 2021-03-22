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

	return user.User{
		Email:        "otto.maddox711@gmail.com",
		LastName:     "Maddox",
		FirstName:    "Otto",
		FullName:     "Otto Maddox",
		HostedDomain: "",
		PictureURL:   "https://lh3.googleusercontent.com/-RYXuWxjLdxo/AAAAAAAAAAI/AAAAAAAAAAA/AMZuucmr33m3QLTjdUatlkppT3NSN5-s8g/s96-c/photo.jpg",
		ProfileLink:  "",
	}
}

// Returns an invalid user defined by the method user.IsValid()
func NewInvalidUser(t *testing.T) user.User {
	t.Helper()

	return user.User{
		Email:        "",
		LastName:     "",
		FirstName:    "",
		FullName:     "Otto Maddox",
		HostedDomain: "",
		PictureURL:   "https://lh3.googleusercontent.com/-RYXuWxjLdxo/AAAAAAAAAAI/AAAAAAAAAAA/AMZuucmr33m3QLTjdUatlkppT3NSN5-s8g/s96-c/photo.jpg",
		ProfileLink:  "",
	}
}
