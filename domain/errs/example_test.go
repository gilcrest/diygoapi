// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package errs_test

import (
	"fmt"
	"net/http/httptest"
	"os"

	"github.com/pkg/errors"

	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/domain/errs"
)

func ExampleError() {
	user := errs.UserName("joe@blow.com")
	// Single error.
	e1 := errs.E(errs.IO, "network unreachable")
	fmt.Println("\nSimple error:")
	fmt.Println(e1)
	// Nested error.
	fmt.Println("\nNested error:")
	e2 := errs.E(user, errs.Other, e1)
	fmt.Println(e2)
	// Output:
	//
	// Simple error:
	// network unreachable
	//
	// Nested error:
	// network unreachable
}
func ExampleMatch() {
	user := errs.UserName("joe@blow.com")
	err := errors.New("network unreachable")
	// Construct an error, one we pretend to have received from a test.
	got := errs.E(user, errs.IO, err)
	// Now construct a reference error, which might not have all
	// the fields of the error from the test.
	expect := errs.E(user, errs.IO, err)
	fmt.Println("Match:", errs.Match(expect, got))
	// Now one that's incorrect - wrong Kind.
	got = errs.E(user, errs.Permission, err)
	fmt.Println("Mismatch:", errs.Match(expect, got))
	// Output:
	//
	// Match: true
	// Mismatch: false
}

func ExampleHTTPErrorResponse() {

	w := httptest.NewRecorder()
	logger := setupLogger()

	err := layer4()
	errs.HTTPErrorResponse(w, logger, err)

	fmt.Println(w.Body)
	// Output:
	//
	// {"level":"error","error":"errors/layer4: input_validation_error] errors/layer3] errors/layer2] errors/layer1|: Actual error message","HTTPStatusCode":400,"Kind":"input_validation_error","Parameter":"testParam","Code":"0212","message":"Response Error Sent"}
	// {"error":{"kind":"input_validation_error","code":"0212","param":"testParam","message":"Actual error message"}}
}

//func ExampleAs() {
//	err := layer4()
//
//	var errsErr *errs.Error
//	if errors.As(err, &errsErr) {
//		fmt.Println("Error Kind:", errsErr.Kind)
//	}
//
//	// Output:
//	//
//	// Error Kind: input_validation_error
//}

func ExampleE() {
	err := layer4()
	if err != nil {
		switch t := errors.Cause(err).(type) {
		case *errs.Error:
			fmt.Printf("%+v\n", t.Err)
		default:
			fmt.Println("WTF!")
		}
	}

	// Output:
	//
	// Actual error message
}

func layer4() error {
	err := layer3()
	return err
}

func layer3() error {
	err := layer2()
	return err
}

func layer2() error {
	err := layer1()
	return err
}

func layer1() error {
	return errs.E(errs.Validation, errs.Parameter("testParam"), errs.Code("0212"), errors.New("Actual error message"))
}

func setupLogger() zerolog.Logger {
	zerolog.TimeFieldFormat = ""

	// set logging level based on input
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	// start a new logger with Stdout as the target
	return zerolog.New(os.Stdout).With().Logger()
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

// func ExampleSentinelErr() {

// 	// ErrFoo is for testing Sentinel Errors with errors.Is
// 	var sentinelErr = errors.New("foo error")

// 	err := sentinelErr

// 	got := errors.Is(err, sentinelErr)

// 	fmt.Println("Is ErrFoo:", got)
// 	// Output:
// 	//
// 	// Is ErrFoo: true
// }
