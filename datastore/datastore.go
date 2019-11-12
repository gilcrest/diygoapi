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
	Tx() (*sql.Tx, error)
	RollbackTx(error) error
	CommitTx() error
}

// DS is a concrete implementation for a database
type DS struct {
	DB *sql.DB
	tx *sql.Tx
}

// BeginTx is a wrapper for sql.DB.BeginTx in order to expose from
// the Datastore interface
func (db *DS) BeginTx(ctx context.Context) error {
	const op errs.Op = "movie/Movie.createDB"

	tx, err := db.DB.BeginTx(ctx, nil)
	if err != nil {
		return errs.E(op, errs.Database, err)
	}

	db.tx = tx

	return nil
}

// Tx exposes the Tx stored in the struct in order to be exposed from
// the Datastore interface.
func (db *DS) Tx() (*sql.Tx, error) {
	const op errs.Op = "movie/Movie.createDB"

	if db.tx == nil {
		return nil, errs.E(op, "DB Transaction has not been started")
	}
	return db.tx, nil
}

// RollbackTx is a wrapper for sql.Tx.Rollback in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (db *DS) RollbackTx(err error) error {
	const op errs.Op = "movie/Movie.createDB"

	// Attempt to rollback the transaction
	if rollbackErr := db.tx.Rollback(); rollbackErr != nil {
		return errs.E(op, errs.Database, err)
	}
	// If rollback was successful, error should be an errs.Error
	// If so, surface error Kind, Code and Param to keep in tact
	var e errs.Error
	if errors.As(err, e) {
		return errs.E(e.Kind, e.Code, e.Param, err)
	}
	// Should not actually fall to here, but including as
	// good practice
	return errs.E(op, errs.Database, err)
}

// CommitTx is a wrapper for sql.Tx.Commit in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (db *DS) CommitTx() error {
	const op errs.Op = "movie/Movie.createDB"

	if err := db.tx.Commit(); err != nil {
		return errs.E(op, errs.Database, err)
	}

	return nil
}

// ProvideDS provides either a DS struct, which has a concrete
// implementation of a database or a MockDS struct which is a mocked
// DB implementation
func ProvideDS(db *sql.DB) Datastore {
	return &DS{DB: db}
}
