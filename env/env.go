// Package env has a type to store common environment related items
// sql db, logger, etc. as well as a constructor-like function to instantiate it
package env

import (
	"github.com/rs/zerolog"
)

// Env type stores common environment related items
type Env struct {
	DS      *Datastore
	Logger  zerolog.Logger
	LogOpts *HTTPLogOpts
}

// NewEnv constructs Env type to be passed around to functions
func NewEnv() (*Env, error) {

	// setup logger
	logger := newLogger()
	// if err != nil {
	// 	return nil, err
	// }

	// open db connection pools
	ds, err := NewDatastore()
	if err != nil {
		return nil, err
	}

	// get logMap with initialized values
	lopts, err := newHTTPLogOpts()
	if err != nil {
		return nil, err
	}

	environment := &Env{Logger: logger, DS: ds, LogOpts: lopts}

	return environment, nil
}
