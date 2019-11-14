package datastore

import (
	"context"
	"database/sql"

	"errors"

	"github.com/gilcrest/go-api-basic/domain/errs"
)

// Datastore is an interface for working with the Database
type Datastore interface {
	BeginTx(context.Context) error
	RollbackTx(error) error
	CommitTx() error
}

// DSName defines the name for the Datastore
type DSName int

const (
	// AppDatastore represents main application database
	AppDatastore DSName = iota
	// LogDatastore represents http logging database
	LogDatastore
	// MockDatastore represents a Mocked Database
	MockDatastore
)

// OS Environment variables for the App DB PostgreSQL Database
const (
	envAppDBName     = "PG_APP_DBNAME"
	envAppDBUser     = "PG_APP_USERNAME"
	envAppDBPassword = "PG_APP_PASSWORD"
	envAppDBHost     = "PG_APP_HOST"
	envAppDBPort     = "PG_APP_PORT"
)

// OS Environment variables for the Log DB PostgreSQL Database
const (
	envLogDBName     = "PG_LOG_DBNAME"
	envLogDBUser     = "PG_LOG_USERNAME"
	envLogDBPassword = "PG_LOG_PASSWORD"
	envLogDBHost     = "PG_LOG_HOST"
	envLogDBPort     = "PG_LOG_PORT"
)

// DS is a concrete implementation for a database
type DS struct {
	DB *sql.DB
	Tx *sql.Tx
}

// BeginTx is a wrapper for sql.DB.BeginTx in order to expose from
// the Datastore interface
func (db *DS) BeginTx(ctx context.Context) error {
	const op errs.Op = "datastore/DS.BeginTx"

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return errs.E(op, errs.Database, err)
	}

	db.Tx = tx

	return nil
}

// RollbackTx is a wrapper for sql.Tx.Rollback in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (db *DS) RollbackTx(err error) error {
	const op errs.Op = "datastore/DS.RollbackTx"

	// Attempt to rollback the transaction
	if rollbackErr := db.Tx.Rollback(); rollbackErr != nil {
		return errs.E(op, errs.Database, err)
	}
	// If rollback was successful, error should be an errs.Error
	// If so, surface error Kind, Code and Param to keep in tact
	var e *errs.Error
	if errors.As(err, &e) {
		return errs.E(e.Kind, e.Code, e.Param, err)
	}
	// Should not actually fall to here, but including as
	// good practice
	return errs.E(op, errs.Database, err)
}

// CommitTx is a wrapper for sql.Tx.Commit in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (db *DS) CommitTx() error {
	const op errs.Op = "datastore/DS.CommitTx"

	if err := db.Tx.Commit(); err != nil {
		return errs.E(op, errs.Database, err)
	}

	return nil
}

// ProvideDatastore provides either a DS struct, which has a concrete
// implementation of a database or a MockDS struct which is a mocked
// DB implementation
func ProvideDatastore(n DSName) (Datastore, error) {
	const op errs.Op = "datastore/ProvideDatastore"

	if n == MockDatastore {
		return &MockDS{}, nil
	}

	db, err := provideDB(n)
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	return &DS{DB: db}, nil
}
