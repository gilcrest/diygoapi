package saaswhip_test

import (
	"github.com/gilcrest/saaswhip"
	"testing"

	qt "github.com/frankban/quicktest"
)

func TestNewProvider(t *testing.T) {
	t.Run("google", func(t *testing.T) {
		c := qt.New(t)
		p := saaswhip.ParseProvider("GoOgLe")
		c.Assert(p, qt.Equals, saaswhip.Google)
	})
	t.Run("unknown", func(t *testing.T) {
		c := qt.New(t)
		p := saaswhip.ParseProvider("anything else!")
		c.Assert(p, qt.Equals, saaswhip.UnknownProvider)
	})
}

func TestProvider_String(t *testing.T) {
	t.Run("google", func(t *testing.T) {
		c := qt.New(t)
		p := saaswhip.ParseProvider("GoOgLe")
		provider := p.String()
		c.Assert(provider, qt.Equals, "google")
	})
	t.Run("unknown", func(t *testing.T) {
		c := qt.New(t)
		p := saaswhip.ParseProvider("anything else")
		provider := p.String()
		c.Assert(provider, qt.Equals, "unknown_provider")
	})
}
