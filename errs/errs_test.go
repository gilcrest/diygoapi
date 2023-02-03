// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errs

import (
	"errors"
	"testing"
)

func TestNoArgs(t *testing.T) {
	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("E() did not panic")
		}
	}()
	_ = E()
}

type kindTest struct {
	err  error
	kind Kind
	want bool
}

func TestKind(t *testing.T) {

	var kindTests = []kindTest{
		// Non-Error errors.
		{nil, NotExist, false},
		{errors.New("not an *Error"), NotExist, false},
		// Basic comparisons.
		{E(NotExist), NotExist, true},
		{E(Exist), NotExist, false},
		{E("no kind"), NotExist, false},
		{E("no kind"), Other, false},
		// Nested *Error values.
		{E("Nesting", E(NotExist)), NotExist, true},
		{E("Nesting", E(Exist)), NotExist, false},
		{E("Nesting", E("no kind")), NotExist, false},
		{E("Nesting", E("no kind")), Other, false},
	}

	for _, test := range kindTests {
		got := KindIs(test.kind, test.err)
		if got != test.want {
			t.Errorf("Is(%q, %q)=%t; want %t", test.kind, test.err, got, test.want)
		}
	}
}
