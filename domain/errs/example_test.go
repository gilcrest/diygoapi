// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package errs_test

import (
	"errors"
	"fmt"

	"github.com/gilcrest/go-api-basic/domain/errs"
)

func ExampleError() {
	path := errs.PathName("jane@doe.com/file")
	user := errs.UserName("joe@blow.com")
	// Single error.
	e1 := errs.E(errs.Op("Get"), path, errs.IO, "network unreachable")
	fmt.Println("\nSimple error:")
	fmt.Println(e1)
	// Nested error.
	fmt.Println("\nNested error:")
	e2 := errs.E(errs.Op("Read"), path, user, errs.Other, e1)
	fmt.Println(e2)
	// Output:
	//
	// Simple error:
	// Get: jane@doe.com/file: I/O_error|: network unreachable
	//
	// Nested error:
	// Read: jane@doe.com/file, user joe@blow.com: I/O_error] Get|: network unreachable
}
func ExampleMatch() {
	path := errs.PathName("jane@doe.com/file")
	user := errs.UserName("joe@blow.com")
	err := errors.New("network unreachable")
	// Construct an error, one we pretend to have received from a test.
	got := errs.E(errs.Op("Get"), path, user, errs.IO, err)
	// Now construct a reference error, which might not have all
	// the fields of the error from the test.
	expect := errs.E(user, errs.IO, err)
	fmt.Println("Match:", errs.Match(expect, got))
	// Now one that's incorrect - wrong Kind.
	got = errs.E(errs.Op("Get"), path, user, errs.Permission, err)
	fmt.Println("Mismatch:", errs.Match(expect, got))
	// Output:
	//
	// Match: true
	// Mismatch: false
}

func ExampleE() {
	err := layer4()

	fmt.Println(err)
	// Output:
	//
	// errors/layer4: input_validation_error] errors/layer3] errors/layer2] errors/layer1|: Actual error message
}

func ExampleRE() {
	err := layer4()

	fmt.Println(errs.RE(err))
	// Output:
	//
	// Actual error message
}

func ExampleAs() {
	err := layer4()

	var errsErr *errs.Error
	if errors.As(err, &errsErr) {
		fmt.Println("Error Kind:", errsErr.Kind)
	}

	// Output:
	//
	// Error Kind: input_validation_error
}

func layer4() error {
	const op errs.Op = "errors/layer4"
	err := layer3()
	return errs.E(op, errs.Validation, err)
}

func layer3() error {
	const op errs.Op = "errors/layer3"
	err := layer2()
	return errs.E(op, errs.Validation, err)
}

func layer2() error {
	const op errs.Op = "errors/layer2"
	err := layer1()
	return errs.E(op, errs.Validation, err)
}

func layer1() error {
	const op errs.Op = "errors/layer1"
	return errs.E(op, errs.Validation, "Actual error message")
}

// One alternative is to return custom error values, called sentinel errors.
// These kind of errors can be found in the Go standard library (sql.ErrNoRows, io.EOF, etc).
// They are useful in that they indicate if a certain kind of error has happened
// (like a database query returning nothing), but they cannot provide any additional context,
// so sentinel errors are not a very flexible tool.

// On the other hand, they are easy to handle, since they’re based on a simple value equality:

// 1	err := db.QueryRow("SELECT * FROM users WHERE id = ?", userID)
// 2	if err == sql.ErrNoRows {
// 3		// handle record not found error
// 4	} else if err != nil {
// 5		// something else went wrong
// 6	}

// ErrFoo is for testing Sentinel Errors with errors.Is
var SentinelErr = errors.New("foo error")

func ExampleSentinelErr() {
	err := SentinelErr

	got := errors.Is(err, SentinelErr)

	fmt.Println("Is ErrFoo:", got)
	// Output:
	//
	// Is ErrFoo: true

}
