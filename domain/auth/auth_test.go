package auth_test

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/gilcrest/diy-go-api/domain/auth"
)

func TestNewProvider(t *testing.T) {
	t.Run("google", func(t *testing.T) {
		c := qt.New(t)
		p := auth.ParseProvider("GoOgLe")
		c.Assert(p, qt.Equals, auth.Google)
	})
	t.Run("apple", func(t *testing.T) {
		c := qt.New(t)
		p := auth.ParseProvider("ApPlE")
		c.Assert(p, qt.Equals, auth.Apple)
	})
	t.Run("invalid", func(t *testing.T) {
		c := qt.New(t)
		p := auth.ParseProvider("anything else!")
		c.Assert(p, qt.Equals, auth.Invalid)
	})
}

func TestProvider_String(t *testing.T) {
	t.Run("google", func(t *testing.T) {
		c := qt.New(t)
		p := auth.ParseProvider("GoOgLe")
		provider := p.String()
		c.Assert(provider, qt.Equals, "google")
	})
	t.Run("apple", func(t *testing.T) {
		c := qt.New(t)
		p := auth.ParseProvider("APPLe")
		provider := p.String()
		c.Assert(provider, qt.Equals, "apple")
	})
	t.Run("invalid", func(t *testing.T) {
		c := qt.New(t)
		p := auth.ParseProvider("anything else")
		provider := p.String()
		c.Assert(provider, qt.Equals, "invalid_provider")
	})
}
