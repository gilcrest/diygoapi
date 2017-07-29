// Package env has a type to store common environment related items
// sql db, logger, etc. as well as a constructor-like function to instantiate it
package env

import (
	"database/sql"

	"github.com/gilcrest/lp/pkg/config/db"

	"go.uber.org/zap"
)

// Env type stores common environment related items
type Env struct {
	Db     *sql.DB
	Logger *zap.Logger
}

// Init Constructs Env type to be passed around to functions
func Init() (*Env, error) {

	logger, _ := zap.NewProduction()

	// returns an open database handle of 0 or more underlying connections
	// func NewDB() (*sql.DB, error)
	sqldb, err := db.NewDB()

	if err != nil {
		return nil, err
	}

	environment := &Env{Db: sqldb, Logger: logger}

	return environment, nil

}
