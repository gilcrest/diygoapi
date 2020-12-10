package errs

import (
	"testing"

	"github.com/pkg/errors"
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
