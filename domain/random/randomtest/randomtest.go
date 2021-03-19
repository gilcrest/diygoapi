// Package randomtest has test helpers for the random package
package randomtest

import "testing"

// NewMockStringGenerator is an initializer for MockStringGenerator
func NewMockStringGenerator(t *testing.T) MockStringGenerator {
	return MockStringGenerator{t: t}
}

// MockStringGenerator creates a static string for testing
type MockStringGenerator struct {
	t *testing.T
}

// CryptoString creates a static string for testing
func (g MockStringGenerator) CryptoString(n int) (string, error) {
	g.t.Helper()

	return "superRandomString", nil
}
