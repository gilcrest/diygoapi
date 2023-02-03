// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package errs_test

import (
	"errors"
	"fmt"
	"github.com/gilcrest/diygoapi/logger"
	"github.com/rs/zerolog"
	"net/http/httptest"
	"os"

	"github.com/gilcrest/diygoapi/errs"
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
	got = errs.E(user, errs.Database, err)
	fmt.Println("Mismatch:", errs.Match(expect, got))
	// Output:
	//
	// Match: true
	// Mismatch: false
}

func ExampleHTTPErrorResponse() {

	w := httptest.NewRecorder()
	l := logger.NewWithGCPHook(os.Stdout, zerolog.DebugLevel, false)

	err := layer4()
	errs.HTTPErrorResponse(w, l, err)

	fmt.Println(w.Body)
	// Output:
	//
	// {"level":"error","error":"Actual error message","http_statuscode":400,"Kind":"input validation error","Parameter":"testParam","Code":"0212","severity":"ERROR","message":"error response sent to client"}
	// {"error":{"kind":"input validation error","code":"0212","param":"testParam","message":"Actual error message"}}
}

func ExampleE() {
	err := layer4()
	if err != nil {
		fmt.Println(err.Error())
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
