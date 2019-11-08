package errs

import (
	"errors"
	"fmt"
	"testing"
)

func TestMissingField_Error(t *testing.T) {

	mf := MissingField("foo")

	tests := []struct {
		name string
		e    MissingField
		want bool
	}{
		{"Test 1", mf, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mfe MissingField
			if got := errors.As(tt.e, &mfe); got != tt.want {
				t.Errorf("Errors.As = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInputUnwanted_Error(t *testing.T) {

	iuFoo := InputUnwanted("foo")
	iuBar := InputUnwanted("bar")
	err := E("not a InputUnwanted Error")

	tests := []struct {
		name string
		e    error
		want bool
	}{
		{"Positive Test", iuFoo, true},
		{"Positive Test 2", iuBar, true},
		{"Negative Test", err, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var iue InputUnwanted
			if got := errors.As(tt.e, &iue); got != tt.want {
				t.Errorf("Errors.As = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBaz(t *testing.T) {
	tests := []struct {
		name          string
		expectedError string
	}{
		{"E Test", "errors/layer4: input_validation_error:\n\terrors/layer3:\n\terrors/layer2:\n\terrors/layer1|: Actual error message"},
		// {"RE Test", "Actual error message"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := bazLayer3()

			var baz BazError
			if errors.As(err, &baz) {
				fmt.Println(baz)
			}

			// if baz.Error() != tt.expectedError {
			// 	t.Errorf("Invalid Error Message: got %q; want %q", baz.Error(), tt.expectedError)
			// }
		})
	}
}

// func lyr4() error {
// 	const op Op = "errors/layer4"
// 	err := layer3()
// 	return E(op, Validation, err)
// }

func bazLayer3() error {
	const op Op = "baz/bazLayer3"
	err := bazLayer2()
	if err != nil {
		return BazError{Reason: "lyr3 reason", Inner: err}
	}
	return errors.New("Invalid Error")
}

func bazLayer2() error {
	const op Op = "baz/bazLayer2"
	err := bazLayer1()
	if err != nil {
		return BazError{Reason: "lyr2 reason", Inner: err}
	}
	return errors.New("Invalid Error")
}

func bazLayer1() error {
	const op Op = "baz/bazLayer1"
	return BazError{Reason: "Actual error message"}
}
