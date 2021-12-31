package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"

	"github.com/gilcrest/go-api-basic/domain/errs"
)

const defaultIDByteLength int = 12

// NewEncryptionKey generates a random 256-bit key. It will return an
// error if the system's secure random number generator fails to
// function correctly, in which case the caller should not continue.
// Taken from https://github.com/gtank/cryptopasta/blob/master/encrypt.go
func NewEncryptionKey() (*[32]byte, error) {
	key := [32]byte{}
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		return nil, errs.E(errs.Internal, err)
	}
	return &key, nil
}

// ParseEncryptionKey decodes the string representation of an encryption key
// and returns its bytes
func ParseEncryptionKey(s string) (*[32]byte, error) {
	// get hex encoded encryption key from cloud secret
	key, err := hex.DecodeString(s)
	if err != nil {
		return nil, errs.E(errs.Internal, err)
	}
	if len(key) != 32 {
		return nil, errs.E(errs.Internal, "Encryption key byte length must be exactly 32 bytes")
	}
	// loop through each byte and add it to the 32 byte encryption key array (ek)
	ek := [32]byte{}
	for i, bite := range key {
		ek[i] = bite
	}
	return &ek, nil
}

// Identifier is a way of identifying something
type Identifier []byte

// NewIdentifier creates a new random Identifier of n bytes or
// returns an error.
func NewIdentifier(n int) (Identifier, error) {
	id, err := CryptoRandomGenerator{}.RandomBytes(n)
	if err != nil {
		return Identifier{}, err
	}

	return id, nil
}

// NewID is like NewIdentifier, but panics if the string cannot be parsed.
func NewID() Identifier {
	id, err := NewIdentifier(defaultIDByteLength)
	if err != nil {
		panic(err)
	}
	return id
}

// String returns the string form of Identifier (base64 encoded
// according to RFC 4648).
func (e Identifier) String() string {
	return base64.URLEncoding.EncodeToString(e)
}

// ParseIdentifier decodes s into Identifier or returns an error.
func ParseIdentifier(s string) (Identifier, error) {
	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return Identifier{}, errs.E(errs.Internal, err)
	}

	return b, nil
}

// MustParseIdentifier is like Parse but panics if the string cannot be parsed.
func MustParseIdentifier(s string) Identifier {
	id, err := ParseIdentifier(s)
	if err != nil {
		panic(err)
	}
	return id
}

// CryptoRandomGenerator methods produce cryptographically secure random data
type CryptoRandomGenerator struct{}

// RandomBytes returns securely generated random bytes. It will return
// an error if the system's secure random number generator fails to
// function correctly, in which case the caller should not continue.
// Taken from https://stackoverflow.com/questions/35781197/generating-a-random-fixed-length-byte-array-in-go
func (g CryptoRandomGenerator) RandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, errs.E(errs.Internal, err)
	}

	return b, nil
}

// RandomString returns a URL-safe, base64 encoded, securely generated, random string.
// It will return an error if the system's secure random number generator fails to
// function correctly, in which case the caller should not continue. This should be
// used when there are concerns about security and need something cryptographically
// secure.
func (g CryptoRandomGenerator) RandomString(n int) (string, error) {
	b, err := g.RandomBytes(n)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), err
}

// Encrypt encrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Output takes the
// form nonce|ciphertext|tag where '|' indicates concatenation.
// Taken from https://github.com/gtank/cryptopasta/blob/master/encrypt.go
func Encrypt(plaintext []byte, key *[32]byte) (ciphertext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, errs.E(errs.Internal, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errs.E(errs.Internal, err)
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, errs.E(errs.Internal, err)
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Expects input
// form nonce|ciphertext|tag where '|' indicates concatenation.
// Taken from https://github.com/gtank/cryptopasta/blob/master/encrypt.go
func Decrypt(ciphertext []byte, key *[32]byte) (plaintext []byte, err error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, errs.E(errs.Internal, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errs.E(errs.Internal, err)
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errs.E(errs.Internal, "malformed ciphertext")
	}

	return gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)
}
