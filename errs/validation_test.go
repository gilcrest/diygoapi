package errs

import (
	"errors"
	"testing"
)

func TestMissingField(t *testing.T) {

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

func TestInputUnwanted(t *testing.T) {

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

func TestMissingField_Error(t *testing.T) {
	tests := []struct {
		name string
		e    MissingField
		want string
	}{
		{"standard", MissingField("some_field"), "some_field is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInputUnwanted_Error(t *testing.T) {
	tests := []struct {
		name string
		e    InputUnwanted
		want string
	}{
		{"standard", InputUnwanted("some_field"), "some_field has a value, but should be nil"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
