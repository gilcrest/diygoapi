package secure_test

import (
	"encoding/hex"
	"testing"

	qt "github.com/frankban/quicktest"

	"github.com/gilcrest/diygoapi/secure"
)

func TestNewEncryptionKey(t *testing.T) {
	t.Run("new key", func(t *testing.T) {
		c := qt.New(t)
		keyBytes, err := secure.NewEncryptionKey()
		c.Logf("Random Key Ciphertext:\t[%s]\n", hex.EncodeToString(keyBytes[:]))
		c.Assert(err, qt.IsNil)
		c.Assert(keyBytes, qt.Not(qt.IsNil))
		c.Assert(len(keyBytes), qt.Equals, 32)
	})
}

func TestParseEncryptionKey(t *testing.T) {
	t.Run("parse key (typical)", func(t *testing.T) {
		c := qt.New(t)

		keyBytes, err := secure.ParseEncryptionKey("f2c100b5661c3b6dc80ba64c499ed7b51482e557e99eeda6126ecc37f2b0381d")
		c.Assert(err, qt.IsNil)
		c.Assert(keyBytes, qt.Not(qt.IsNil))
		c.Assert(len(keyBytes), qt.Equals, 32)
	})
}
