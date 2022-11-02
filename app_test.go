package diy_test

import (
	"fmt"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/uuid"

	"github.com/gilcrest/diy-go-api"
	"github.com/gilcrest/diy-go-api/secure"
)

func TestApp_AddKey(t *testing.T) {
	t.Run("add valid key", func(t *testing.T) {
		c := qt.New(t)

		o := &diy.Org{
			ID:          uuid.New(),
			ExternalID:  secure.NewID(),
			Name:        "app test",
			Description: "test app",
			Kind:        &diy.OrgKind{},
		}

		a := diy.App{
			ID:          uuid.New(),
			ExternalID:  secure.NewID(),
			Org:         o,
			Name:        "",
			Description: "",
			APIKeys:     nil,
		}
		var (
			ek  *[32]byte
			err error
		)
		ek, err = secure.NewEncryptionKey()
		c.Assert(err, qt.IsNil)

		var key diy.APIKey
		key, err = diy.NewAPIKey(secure.RandomGenerator{}, ek, time.Now().Add(time.Hour*100))
		c.Assert(err, qt.IsNil)

		err = a.AddKey(key)
		c.Assert(err, qt.IsNil)
		c.Assert(len(a.APIKeys), qt.Equals, 1)
	})
	t.Run("add expired key", func(t *testing.T) {
		c := qt.New(t)

		o := &diy.Org{
			ID:          uuid.New(),
			ExternalID:  secure.NewID(),
			Name:        "app test",
			Description: "test app",
			Kind:        &diy.OrgKind{},
		}

		a := diy.App{
			ID:          uuid.New(),
			ExternalID:  secure.NewID(),
			Org:         o,
			Name:        "",
			Description: "",
			APIKeys:     nil,
		}
		var (
			ek  *[32]byte
			err error
		)
		ek, err = secure.NewEncryptionKey()
		c.Assert(err, qt.IsNil)

		var key diy.APIKey
		key, err = diy.NewAPIKey(secure.RandomGenerator{}, ek, time.Now().Add(time.Hour*-100))
		c.Assert(err, qt.IsNil)

		err = a.AddKey(key)
		c.Assert(err, qt.Not(qt.IsNil))
	})
}

func TestApp_ValidateKey(t *testing.T) {
	t.Run("valid key", func(t *testing.T) {
		c := qt.New(t)

		o := &diy.Org{
			ID:          uuid.New(),
			ExternalID:  secure.NewID(),
			Name:        "app test",
			Description: "test app",
			Kind:        &diy.OrgKind{},
		}

		a := diy.App{
			ID:          uuid.New(),
			ExternalID:  secure.NewID(),
			Org:         o,
			Name:        "",
			Description: "",
			APIKeys:     nil,
		}
		var (
			ek  *[32]byte
			err error
		)
		ek, err = secure.NewEncryptionKey()
		c.Assert(err, qt.IsNil)

		var key diy.APIKey
		key, err = diy.NewAPIKey(secure.RandomGenerator{}, ek, time.Now().Add(time.Hour*100))
		c.Assert(err, qt.IsNil)

		err = a.AddKey(key)
		c.Assert(err, qt.IsNil)

		err = a.ValidateKey("deep in the realm", key.Key())
		c.Assert(err, qt.IsNil)
	})
	t.Run("key does not match", func(t *testing.T) {
		c := qt.New(t)

		o := &diy.Org{
			ID:          uuid.New(),
			ExternalID:  secure.NewID(),
			Name:        "app test",
			Description: "test app",
			Kind:        &diy.OrgKind{},
		}

		a := diy.App{
			ID:          uuid.New(),
			ExternalID:  secure.NewID(),
			Org:         o,
			Name:        "",
			Description: "",
			APIKeys:     nil,
		}

		err := a.ValidateKey("deep in the realm", "badkey")
		c.Assert(err, qt.ErrorMatches, "Key does not match any keys for the App")
	})
	t.Run("key matches but invalid", func(t *testing.T) {
		c := qt.New(t)

		o := &diy.Org{
			ID:          uuid.New(),
			ExternalID:  secure.NewID(),
			Name:        "app test",
			Description: "test app",
			Kind:        &diy.OrgKind{},
		}

		a := diy.App{
			ID:          uuid.New(),
			ExternalID:  secure.NewID(),
			Org:         o,
			Name:        "",
			Description: "",
			APIKeys:     nil,
		}
		var (
			ek  *[32]byte
			err error
		)
		ek, err = secure.NewEncryptionKey()
		c.Assert(err, qt.IsNil)

		var key diy.APIKey
		key, err = diy.NewAPIKey(secure.RandomGenerator{}, ek, time.Now().Add(time.Hour*-100))
		c.Assert(err, qt.IsNil)

		a.APIKeys = append(a.APIKeys, key)

		err = a.ValidateKey("deep in the realm", key.Key())
		c.Assert(err, qt.ErrorMatches, fmt.Sprintf("Key Deactivation %s is before current time .*", key.DeactivationDate().String()))
	})
}
