// Package errs is a modified copy of the upspin.io/errors package.
// Originally, I used quite a bit of the upspin.io/errors package,
// but have moved to only use a very small amount of it. Even still,
// I think it's appropriate to leave the license information in...
//
// Copyright 2016 The Upspin Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Package errs defines the error handling used by all Upspin software.
package errs

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"runtime"
	"sort"

	pkgerrors "github.com/pkg/errors"
)

// Error is the type that implements the error interface.
// It contains a number of fields, each of different type.
// An Error value may leave some values unset.
type Error struct {
	// Op is the operation being performed, usually the name of the method
	// being invoked.
	Op Op
	// User is the name of the user attempting the operation.
	User UserName
	// Kind is the class of error, such as permission failure,
	// or "Other" if its class is unknown or irrelevant.
	Kind Kind
	// Param represents the parameter related to the error.
	Param Parameter
	// Code is a human-readable, short representation of the error
	Code Code
	// Realm is a description of a protected area, used in the WWW-Authenticate header.
	Realm Realm
	// The underlying error that triggered this one, if any.
	Err error
}

func (e *Error) isZero() bool {
	return e.User == "" && e.Kind == 0 && e.Param == "" && e.Code == "" && e.Err == nil
}

// Unwrap method allows for unwrapping errors using errors.As
func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) Error() string {
	return e.Err.Error()
}

// OpStack returns the op stack information for an error
func OpStack(err error) []string {
	type o struct {
		Op    string
		Order int
	}

	e := err
	i := 0
	var os []o

	// loop through all wrapped errors and add to struct
	// order will be from top to bottom of stack
	for errors.Unwrap(e) != nil {
		var errsError *Error
		if errors.As(e, &errsError) {
			if errsError.Op != "" {
				op := o{Op: string(errsError.Op), Order: i}
				os = append(os, op)
			}
		}
		e = errors.Unwrap(e)
		i++
	}

	// reverse the order of the stack (bottom to top)
	sort.Slice(os, func(i, j int) bool { return os[i].Order > os[j].Order })

	// pull out just the stack info, now in reversed order
	var ops []string
	for _, op := range os {
		ops = append(ops, op.Op)
	}

	return ops
}

// TopError recursively unwraps all errors and retrieves the topmost error
func TopError(err error) error {
	currentErr := err
	for errors.Unwrap(currentErr) != nil {
		currentErr = errors.Unwrap(currentErr)
	}

	return currentErr
}

// Op describes an operation, usually as the package and method,
// such as "key/server.Lookup".
type Op string

// UserName is a string representing a user
type UserName string

// Kind defines the kind of error this is, mostly for use by systems
// such as FUSE that must act differently depending on the error.
type Kind uint8

// Parameter represents the parameter related to the error.
type Parameter string

// Code is a human-readable, short representation of the error
type Code string

// Realm is a description of a protected area, used in the WWW-Authenticate header.
// Realm should be set when error Kind is Unauthenticated. If left unset, Realm
// will be set to the default set by the Default method
type Realm string

// Kinds of errors.
//
// The values of the error kinds are common between both
// clients and servers. Do not reorder this list or remove
// any items since that will change their values.
// New items must be added only to the end.
const (
	Other          Kind = iota // Unclassified error. This value is not printed in the error message.
	Invalid                    // Invalid operation for this type of item.
	IO                         // External I/O error such as network failure.
	Exist                      // Item already exists.
	NotExist                   // Item does not exist.
	Private                    // Information withheld.
	Internal                   // Internal error or inconsistency.
	BrokenLink                 // Link target does not exist.
	Database                   // Error from database.
	Validation                 // Input validation error.
	Unanticipated              // Unanticipated error.
	InvalidRequest             // Invalid Request
	// Unauthenticated is used when a request lacks valid authentication credentials.
	//
	// For Unauthenticated errors, the response body will be empty.
	// The error is logged and http.StatusUnauthorized (401) is sent.
	Unauthenticated // Unauthenticated Request
	// Unauthorized is used when a user is authenticated, but is not authorized
	// to access the resource.
	//
	// For Unauthorized errors, the response body should be empty.
	// The error is logged and http.StatusForbidden (403) is sent.
	Unauthorized
)

func (k Kind) String() string {
	switch k {
	case Other:
		return "other error"
	case Invalid:
		return "invalid operation"
	case IO:
		return "I/O error"
	case Exist:
		return "item already exists"
	case NotExist:
		return "item does not exist"
	case BrokenLink:
		return "link target does not exist"
	case Private:
		return "information withheld"
	case Internal:
		return "internal error"
	case Database:
		return "database error"
	case Validation:
		return "input validation error"
	case Unanticipated:
		return "unanticipated error"
	case InvalidRequest:
		return "invalid request error"
	case Unauthenticated:
		return "unauthenticated request"
	case Unauthorized:
		return "unauthorized request"
	}
	return "unknown error kind"
}

// E builds an error value from its arguments.
// There must be at least one argument or E panics.
// The type of each argument determines its meaning.
// If more than one argument of a given type is presented,
// only the last one is recorded.
//
// The types are:
//
//	UserName
//		The username of the user attempting the operation.
//	string
//		Treated as an error message and assigned to the
//		Err field after a call to errors.New.
//	errors.Kind
//		The class of error, such as permission failure.
//	error
//		The underlying error that triggered this one.
//
// If the error is printed, only those items that have been
// set to non-zero values will appear in the result.
//
// If Kind is not specified or Other, we set it to the Kind of
// the underlying error.
func E(args ...interface{}) error {
	type stackTracer interface {
		StackTrace() pkgerrors.StackTrace
	}

	if len(args) == 0 {
		panic("call to errors.E with no arguments")
	}
	e := &Error{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case Op:
			e.Op = arg
		case UserName:
			e.User = arg
		case Kind:
			e.Kind = arg
		case string:
			if zerolog.ErrorStackMarshaler != nil {
				e.Err = pkgerrors.New(arg)
			} else {
				e.Err = Str(arg)
			}
		case *Error:
			// Make a copy
			errorCopy := *arg
			e.Err = &errorCopy
		case error:
			if zerolog.ErrorStackMarshaler != nil {
				// if the error implements stackTracer, then it is
				// a pkg/errors error type and does not need to have
				// the stack added
				_, ok := arg.(stackTracer)
				if ok {
					e.Err = arg
				} else {
					e.Err = pkgerrors.New(arg.Error())
				}
			} else {
				e.Err = arg
			}
		case Code:
			e.Code = arg
		case Parameter:
			e.Param = arg
		case Realm:
			e.Realm = arg
		default:
			_, file, line, _ := runtime.Caller(1)
			return fmt.Errorf("errors.E: bad call from %s:%d: %v, unknown type %T, value %v in error call", file, line, args, arg, arg)
		}
	}

	prev, ok := e.Err.(*Error)
	if !ok {
		return e
	}

	// If this error has Kind unset or Other, pull up the inner one.
	if e.Kind == Other {
		e.Kind = prev.Kind
		prev.Kind = Other
	}

	if prev.Code == e.Code {
		prev.Code = ""
	}
	// If this error has Code == "", pull up the inner one.
	if e.Code == "" {
		e.Code = prev.Code
		prev.Code = ""
	}

	if prev.Param == e.Param {
		prev.Param = ""
	}
	// If this error has Param == "", pull up the inner one.
	if e.Param == "" {
		e.Param = prev.Param
		prev.Param = ""
	}

	if prev.Realm == e.Realm {
		prev.Realm = ""
	}
	// If this error has WWWAuthenticateRealm == "", pull up the inner one.
	if e.Realm == "" {
		e.Realm = prev.Realm
		prev.Realm = ""
	}

	return e
}

// Str returns an error that formats as the given text. It is intended to
// be used as the error-typed argument to the E function.
func Str(text string) error {
	return &errorString{text}
}

// errorString is a trivial implementation of error.
type errorString struct {
	s string
}

func (e *errorString) Error() string {
	return e.s
}

// Match compares its two error arguments. It can be used to check
// for expected errors in tests. Both arguments must have underlying
// type *Error or Match will return false. Otherwise, it returns true
// if every non-zero element of the first error is equal to the
// corresponding element of the second.
// If the Err field is a *Error, Match recurs on that field;
// otherwise it compares the strings returned by the Error methods.
// Elements that are in the second argument but not present in
// the first are ignored.
//
// For example,
//
//		Match(errors.E(upspin.UserName("joe@schmoe.com"), errors.Permission), err)
//	 tests whether err is an Error with Kind=Permission and User=joe@schmoe.com.
func Match(err1, err2 error) bool {
	e1, ok := err1.(*Error)
	if !ok {
		return false
	}
	var e2 *Error
	e2, ok = err2.(*Error)
	if !ok {
		return false
	}
	if e1.User != "" && e2.User != e1.User {
		return false
	}
	if e1.Kind != Other && e2.Kind != e1.Kind {
		return false
	}
	if e1.Param != "" && e2.Param != e1.Param {
		return false
	}
	if e1.Code != "" && e2.Code != e1.Code {
		return false
	}
	if e1.Err != nil {
		if _, k := e1.Err.(*Error); k {
			return Match(e1.Err, e2.Err)
		}
		if e2.Err == nil || e2.Err.Error() != e1.Err.Error() {
			return false
		}
	}
	return true
}

// KindIs reports whether err is an *Error of the given Kind.
// If err is nil then KindIs returns false.
func KindIs(kind Kind, err error) bool {
	var e *Error
	if errors.As(err, &e) {
		if e.Kind != Other {
			return e.Kind == kind
		}
		if e.Err != nil {
			return KindIs(kind, e.Err)
		}
	}
	return false
}
