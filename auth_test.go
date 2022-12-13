package diygoapi_test

import (
	"github.com/gilcrest/diygoapi"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestNewProvider(t *testing.T) {
	t.Run("google", func(t *testing.T) {
		c := qt.New(t)
		p := diygoapi.ParseProvider("GoOgLe")
		c.Assert(p, qt.Equals, diygoapi.Google)
	})
	t.Run("unknown", func(t *testing.T) {
		c := qt.New(t)
		p := diygoapi.ParseProvider("anything else!")
		c.Assert(p, qt.Equals, diygoapi.UnknownProvider)
	})
}

func TestProvider_String(t *testing.T) {
	t.Run("google", func(t *testing.T) {
		c := qt.New(t)
		p := diygoapi.ParseProvider("GoOgLe")
		provider := p.String()
		c.Assert(provider, qt.Equals, "google")
	})
	t.Run("unknown", func(t *testing.T) {
		c := qt.New(t)
		p := diygoapi.ParseProvider("anything else")
		provider := p.String()
		c.Assert(provider, qt.Equals, "unknown_provider")
	})
}
