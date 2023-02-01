package secure

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io"

	"github.com/gilcrest/diygoapi/errs"
)

const defaultIDByteLength int = 12

// Identifier is a random, cryptographically generated sequence of characters used to refer to something
type Identifier []byte

// NewIdentifier creates a new random Identifier of n bytes or
// returns an error.
func NewIdentifier(n int) (Identifier, error) {
	const op errs.Op = "secure/NewIdentifier"

	id, err := RandomGenerator{}.RandomBytes(n)
	if err != nil {
		return Identifier{}, errs.E(op, err)
	}

	return id, nil
}

// NewID is like NewIdentifier, but panics if the Identifier
// cannot be initialized
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
	const op errs.Op = "secure/ParseIdentifier"

	b, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return Identifier{}, errs.E(op, errs.Internal, err)
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

// NewEncryptionKey generates a random 256-bit key. It will return an
// error if the system's secure random number generator fails to
// function correctly, in which case the caller should not continue.
// Taken from https://github.com/gtank/cryptopasta/blob/master/encrypt.go
func NewEncryptionKey() (*[32]byte, error) {
	const op errs.Op = "secure/NewEncryptionKey"

	key := [32]byte{}
	_, err := io.ReadFull(rand.Reader, key[:])
	if err != nil {
		return nil, errs.E(op, errs.Internal, err)
	}
	return &key, nil
}

// ParseEncryptionKey decodes the string representation of an encryption key
// and returns its bytes
func ParseEncryptionKey(s string) (*[32]byte, error) {
	const op errs.Op = "secure/ParseEncryptionKey"

	// get hex encoded encryption key from cloud secret
	key, err := hex.DecodeString(s)
	if err != nil {
		return nil, errs.E(op, errs.Internal, err)
	}
	if len(key) != 32 {
		return nil, errs.E(op, errs.Internal, "Encryption key byte length must be exactly 32 bytes")
	}
	// loop through each byte and add it to the 32 byte encryption key array (ek)
	ek := [32]byte{}
	for i, bite := range key {
		ek[i] = bite
	}
	return &ek, nil
}

// Encrypt encrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Output takes the
// form nonce|ciphertext|tag where '|' indicates concatenation.
// Taken from https://github.com/gtank/cryptopasta/blob/master/encrypt.go
func Encrypt(plaintext []byte, key *[32]byte) (ciphertext []byte, err error) {
	const op errs.Op = "secure/Encrypt"

	var block cipher.Block
	block, err = aes.NewCipher(key[:])
	if err != nil {
		return nil, errs.E(op, errs.Internal, err)
	}

	var gcm cipher.AEAD
	gcm, err = cipher.NewGCM(block)
	if err != nil {
		return nil, errs.E(op, errs.Internal, err)
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return nil, errs.E(op, errs.Internal, err)
	}

	return gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts data using 256-bit AES-GCM.  This both hides the content of
// the data and provides a check that it hasn't been altered. Expects input
// form nonce|ciphertext|tag where '|' indicates concatenation.
// Taken from https://github.com/gtank/cryptopasta/blob/master/encrypt.go
func Decrypt(ciphertext []byte, key *[32]byte) (plaintext []byte, err error) {
	const op errs.Op = "secure/Decrypt"

	var block cipher.Block
	block, err = aes.NewCipher(key[:])
	if err != nil {
		return nil, errs.E(op, errs.Internal, err)
	}

	var gcm cipher.AEAD
	gcm, err = cipher.NewGCM(block)
	if err != nil {
		return nil, errs.E(op, errs.Internal, err)
	}

	if len(ciphertext) < gcm.NonceSize() {
		return nil, errs.E(op, errs.Internal, "malformed ciphertext")
	}

	plaintext, err = gcm.Open(nil,
		ciphertext[:gcm.NonceSize()],
		ciphertext[gcm.NonceSize():],
		nil,
	)
	if err != nil {
		return nil, errs.E(errs.Internal, err)
	}

	return plaintext, nil
}
