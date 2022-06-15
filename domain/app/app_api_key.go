package app

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/gilcrest/diy-go-api/domain/errs"
	"github.com/gilcrest/diy-go-api/domain/secure"
)

// APIKeyStringGenerator creates a random, 128 API key string
type APIKeyStringGenerator interface {
	RandomString(n int) (string, error)
}

// APIKey is an API key for interacting with the system
type APIKey struct {
	// key: the unencrypted API key string
	key string
	// ciphertext: the encrypted API key as []byte
	ciphertext []byte
	// deactivation: the date/time the API key is no longer usable
	deactivation time.Time
}

// NewAPIKey initializes an APIKey. It generates both a 128-bit (16 byte)
// random string as an API key and its corresponding ciphertext bytes
func NewAPIKey(g APIKeyStringGenerator, ek *[32]byte) (APIKey, error) {
	var (
		k   string
		err error
	)
	k, err = g.RandomString(18)
	if err != nil {
		return APIKey{}, err
	}

	var ct []byte
	ct, err = secure.Encrypt([]byte(k), ek)
	if err != nil {
		return APIKey{}, err
	}

	return APIKey{key: k, ciphertext: ct}, nil
}

// NewAPIKeyFromCipher initializes an APIKey
func NewAPIKeyFromCipher(ciphertext string, ek *[32]byte) (APIKey, error) {
	var (
		eak []byte
		err error
	)

	// encrypted key is stored using hex in db. Decode to get ciphertext bytes.
	eak, err = hex.DecodeString(ciphertext)
	if err != nil {
		return APIKey{}, errs.E(errs.Internal, err)
	}

	var apiKey []byte
	apiKey, err = secure.Decrypt(eak, ek)
	if err != nil {
		return APIKey{}, err
	}

	return APIKey{key: string(apiKey), ciphertext: eak}, nil
}

// Key returns the key for the API key
func (a APIKey) Key() string {
	return a.key
}

// Ciphertext returns the hex encoded text of the encrypted cipher bytes for the API key
func (a APIKey) Ciphertext() string {
	return hex.EncodeToString(a.ciphertext)
}

// DeactivationDate returns the Deactivation Date for the API key
func (a APIKey) DeactivationDate() time.Time {
	return a.deactivation
}

// SetDeactivationDate sets the deactivation date value to AppAPIkey
// TODO - try SetDeactivationDate as a candidate for generics with 1.18
func (a *APIKey) SetDeactivationDate(t time.Time) {
	a.deactivation = t
}

// SetStringAsDeactivationDate sets the deactivation date value to
// AppAPIkey given a string in RFC3339 format
func (a *APIKey) SetStringAsDeactivationDate(s string) error {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return errs.E(errs.Validation, err)
	}
	a.deactivation = t

	return nil
}

// isValid validates the API Key
func (a APIKey) isValid() error {
	if a.ciphertext == nil {
		return errs.E("ciphertext must have a value")
	}

	now := time.Now()
	if a.deactivation.Before(now) {
		return errs.E(fmt.Sprintf("Key Deactivation %s is before current time %s", a.deactivation.String(), now.String()))
	}
	return nil
}
