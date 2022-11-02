package diy_test

import (
	"encoding/base64"
	"encoding/hex"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"

	"github.com/gilcrest/diy-go-api"
	"github.com/gilcrest/diy-go-api/secure"
)

func TestNewAPIKey(t *testing.T) {
	t.Run("key byte length", func(t *testing.T) {
		c := qt.New(t)
		var (
			ek  *[32]byte
			err error
		)
		ek, err = secure.NewEncryptionKey()
		c.Assert(err, qt.IsNil)

		var key diy.APIKey
		key, err = diy.NewAPIKey(secure.RandomGenerator{}, ek, time.Date(2999, 12, 31, 0, 0, 0, 0, time.UTC))
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

		var key diy.APIKey
		key, err = diy.NewAPIKey(secure.RandomGenerator{}, ek, time.Date(2999, 12, 31, 0, 0, 0, 0, time.UTC))
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
