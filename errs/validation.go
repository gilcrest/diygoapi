package errs

// MissingField is an error type that can be used when
// validating input fields that do not have a value, but should
type MissingField string

func (e MissingField) Error() string {
	return string(e) + " is required"
}

// InputUnwanted is an error type that can be used when
// validating input fields that have a value, but should not
type InputUnwanted string

func (e InputUnwanted) Error() string {
	return string(e) + " has a value, but should be nil"
}
