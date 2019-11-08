package errs

import "fmt"

// MissingField is an error type that can be used when
// validating input fields that do not have a value, but should
type MissingField string

func (e MissingField) Error() string {
	return string(e) + " is required"
}

// InputUnwanted is an error type that can be used when
// validating input fields that have a value, but should should not
type InputUnwanted string

func (e InputUnwanted) Error() string {
	return string(e) + " has a value, but should be nil"
}

// BazError is a temp error until I figure this out
type BazError struct {
	Reason string
	Inner  error
}

func (e BazError) Error() string {
	if e.Inner != nil {
		return fmt.Sprintf("baz error: %s: %v", e.Reason, e.Inner)
	}
	return fmt.Sprintf("baz error: %s", e.Reason)
}

// Unwrap does some unwrapping shit
func (e BazError) Unwrap() error {
	return e.Inner
}
