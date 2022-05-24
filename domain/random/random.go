// Package random has helper functions to create random strings or bytes
package random

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/gilcrest/diy-go-api/domain/errs"
)

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
// Taken from https://stackoverflow.com/questions/35781197/generating-a-random-fixed-length-byte-array-in-go
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, errs.E(errs.Internal, err)
	}

	return b, nil
}

// StringGenerator generates random strings
type StringGenerator struct{}

// CryptoString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue. This should be used
// when there are concerns about security and need something
// cryptographically secure.
func (sg StringGenerator) CryptoString(n int) (string, error) {
	b, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), err
}
