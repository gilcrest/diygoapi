// Package env has a type to store common environment related items
// sql db, logger, etc. as well as a constructor-like function to instantiate it
package env

import (
	"github.com/gilcrest/go-API-template/pkg/datastore"
	"go.uber.org/zap"
)

// Env type stores common environment related items
type Env struct {
	Logger *zap.Logger
	DS     *datastore.Datastore
}

// NewEnv constructs Env type to be passed around to functions
func NewEnv() (*Env, error) {

	// setup logger
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	// open db connection pools
	ds, err := datastore.NewDatastore()
	if err != nil {
		return nil, err
	}

	environment := &Env{Logger: logger, DS: ds}

	return environment, nil

}
