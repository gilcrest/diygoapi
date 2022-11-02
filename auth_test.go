package diy_test

import (
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/gilcrest/diy-go-api"
)

func TestNewProvider(t *testing.T) {
	t.Run("google", func(t *testing.T) {
		c := qt.New(t)
		p := diy.ParseProvider("GoOgLe")
		c.Assert(p, qt.Equals, diy.Google)
	})
	t.Run("unknown", func(t *testing.T) {
		c := qt.New(t)
		p := diy.ParseProvider("anything else!")
		c.Assert(p, qt.Equals, diy.UnknownProvider)
	})
}

func TestProvider_String(t *testing.T) {
	t.Run("google", func(t *testing.T) {
		c := qt.New(t)
		p := diy.ParseProvider("GoOgLe")
		provider := p.String()
		c.Assert(provider, qt.Equals, "google")
	})
	t.Run("unknown", func(t *testing.T) {
		c := qt.New(t)
		p := diy.ParseProvider("anything else")
		provider := p.String()
		c.Assert(provider, qt.Equals, "unknown_provider")
	})
}
