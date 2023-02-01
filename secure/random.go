package secure

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/gilcrest/diygoapi/errs"
)

// RandomGenerator produces cryptographically secure random data
type RandomGenerator struct{}

// RandomBytes returns securely generated random bytes. It will return
// an error if the system's secure random number generator fails to
// function correctly, in which case the caller should not continue.
// Taken from https://stackoverflow.com/questions/35781197/generating-a-random-fixed-length-byte-array-in-go
func (g RandomGenerator) RandomBytes(n int) ([]byte, error) {
	const op errs.Op = "secure/RandomGenerator.RandomBytes"

	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, errs.E(op, errs.Internal, err)
	}

	return b, nil
}

// RandomString returns a URL-safe, base64 encoded, securely generated, random string.
// It will return an error if the system's secure random number generator fails to
// function correctly, in which case the caller should not continue. This should be
// used when there are concerns about security and need something cryptographically
// secure.
func (g RandomGenerator) RandomString(n int) (string, error) {
	const op errs.Op = "secure/RandomGenerator.RandomString"

	b, err := g.RandomBytes(n)
	if err != nil {
		return "", errs.E(op, err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
