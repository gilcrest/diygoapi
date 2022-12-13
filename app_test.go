package diygoapi_test

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/uuid"

	"github.com/gilcrest/diygoapi"
	"github.com/gilcrest/diygoapi/secure"
)

func TestApp_AddKey(t *testing.T) {
	t.Run("add valid key", func(t *testing.T) {
		c := qt.New(t)

		o := &diygoapi.Org{
			ID:          uuid.New(),
			ExternalID:  secure.NewID(),
			Name:        "app test",
			Description: "test app",
			Kind:        &diygoapi.OrgKind{},
		}

		a := diygoapi.App{
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

		var key diygoapi.APIKey
		key, err = diygoapi.NewAPIKey(secure.RandomGenerator{}, ek, time.Now().Add(time.Hour*100))
		c.Assert(err, qt.IsNil)

		err = a.AddKey(key)
		c.Assert(err, qt.IsNil)
		c.Assert(len(a.APIKeys), qt.Equals, 1)
	})
	t.Run("add expired key", func(t *testing.T) {
		c := qt.New(t)

		o := &diygoapi.Org{
			ID:          uuid.New(),
			ExternalID:  secure.NewID(),
			Name:        "app test",
			Description: "test app",
			Kind:        &diygoapi.OrgKind{},
		}

		a := diygoapi.App{
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

		var key diygoapi.APIKey
		key, err = diygoapi.NewAPIKey(secure.RandomGenerator{}, ek, time.Now().Add(time.Hour*-100))
		c.Assert(err, qt.IsNil)

		err = a.AddKey(key)
		c.Assert(err, qt.Not(qt.IsNil))
	})
}

func TestApp_ValidateKey(t *testing.T) {
	t.Run("valid key", func(t *testing.T) {
		c := qt.New(t)

		o := &diygoapi.Org{
			ID:          uuid.New(),
			ExternalID:  secure.NewID(),
			Name:        "app test",
			Description: "test app",
			Kind:        &diygoapi.OrgKind{},
		}

		a := diygoapi.App{
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

		var key diygoapi.APIKey
		key, err = diygoapi.NewAPIKey(secure.RandomGenerator{}, ek, time.Now().Add(time.Hour*100))
		c.Assert(err, qt.IsNil)

		err = a.AddKey(key)
		c.Assert(err, qt.IsNil)

		err = a.ValidateKey("deep in the realm", key.Key())
		c.Assert(err, qt.IsNil)
	})
	t.Run("key does not match", func(t *testing.T) {
		c := qt.New(t)

		o := &diygoapi.Org{
			ID:          uuid.New(),
			ExternalID:  secure.NewID(),
			Name:        "app test",
			Description: "test app",
			Kind:        &diygoapi.OrgKind{},
		}

		a := diygoapi.App{
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

		o := &diygoapi.Org{
			ID:          uuid.New(),
			ExternalID:  secure.NewID(),
			Name:        "app test",
			Description: "test app",
			Kind:        &diygoapi.OrgKind{},
		}

		a := diygoapi.App{
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

		var key diygoapi.APIKey
		key, err = diygoapi.NewAPIKey(secure.RandomGenerator{}, ek, time.Now().Add(time.Hour*-100))
		c.Assert(err, qt.IsNil)

		a.APIKeys = append(a.APIKeys, key)

		err = a.ValidateKey("deep in the realm", key.Key())
		c.Assert(err, qt.ErrorMatches, fmt.Sprintf("Key Deactivation %s is before current time .*", key.DeactivationDate().String()))
	})
}

func TestNewAPIKey(t *testing.T) {
	t.Run("key byte length", func(t *testing.T) {
		c := qt.New(t)
		var (
			ek  *[32]byte
			err error
		)
		ek, err = secure.NewEncryptionKey()
		c.Assert(err, qt.IsNil)

		var key diygoapi.APIKey
		key, err = diygoapi.NewAPIKey(secure.RandomGenerator{}, ek, time.Date(2999, 12, 31, 0, 0, 0, 0, time.UTC))
		c.Assert(err, qt.IsNil)

		// decode base64
		var keyBytes []byte
		keyBytes, err = base64.URLEncoding.DecodeString(key.Key())
		c.Assert(err, qt.IsNil)

		c.Assert(len(keyBytes), qt.Equals, 16, qt.Commentf("assure key byte length is always 16 (128-bit)"))
	})
	t.Run("decrypt key", func(t *testing.T) {
		c := qt.New(t)
		var (
			ek  *[32]byte
			err error
		)
		ek, err = secure.NewEncryptionKey()
		c.Assert(err, qt.IsNil)

		var key diygoapi.APIKey
		key, err = diygoapi.NewAPIKey(secure.RandomGenerator{}, ek, time.Date(2999, 12, 31, 0, 0, 0, 0, time.UTC))
		c.Assert(err, qt.IsNil)

		// Ciphertext method returns the bytes as a hex encoded string.
		// decode to get the bytes
		var cb []byte
		cb, err = hex.DecodeString(key.Ciphertext())
		c.Assert(err, qt.IsNil)

		// decrypt the encrypted key
		var apiKey []byte
		apiKey, err = secure.Decrypt(cb, ek)
		c.Assert(err, qt.IsNil)

		c.Assert(string(apiKey), qt.Equals, key.Key(), qt.Commentf("ensure decrypted key matches key string"))
	})
}
