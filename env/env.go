// Package env has a type to store common environment related items
// sql db, logger, etc. as well as a constructor-like function to instantiate it
package env

import (
	"fmt"

	"github.com/gilcrest/go-API-template/db"
	"github.com/rs/zerolog"
)

// Env type stores common environment related items
type Env struct {
	DS      *db.Datastore
	Logger  zerolog.Logger
	LogOpts *HTTPLogOpts
}

// NewEnv constructs Env type to be passed around to functions
func NewEnv(lvl zerolog.Level) (*Env, error) {

	// setup logger
	logger := newLogger(lvl)
	// if err != nil {
	// 	return nil, err
	// }

	// open db connection pools
	ds, err := db.NewDatastore()
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

// LogErr logs the Operation (client.Lookup, etc.) as well
// as the error string and returns an error
func (e *Env) LogErr(op string, s string) error {
	err := fmt.Errorf("%s: %s", op, s)
	e.Logger.Error().Err(err).Msg("")
	return err
}
