package datastore

import (
	"context"
	"database/sql"

	"github.com/gilcrest/go-api-basic/domain/errs"
)

// MockDS is a mock implementation for a database
type MockDS struct {
}

// BeginTx is a wrapper for sql.DB.BeginTx in order to expose from
// the Datastore interface
func (db *MockDS) BeginTx(ctx context.Context) error {
	const op errs.Op = "datastore/MockDS.BeginTx"

	return nil
}

// Tx exposes the Tx stored in the struct in order to be exposed from
// the Datastore interface.
func (db *MockDS) Tx() (*sql.Tx, error) {
	const op errs.Op = "datastore/MockDS.Tx"

	return nil, nil
}

// RollbackTx is a wrapper for sql.Tx.Rollback in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (db *MockDS) RollbackTx(err error) error {
	const op errs.Op = "datastore/MockDS.RollbackTx"

	return errs.E(op, errs.Database, err)
}

// CommitTx is a wrapper for sql.Tx.Commit in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (db *MockDS) CommitTx() error {
	const op errs.Op = "datastore/MockDS.CommitTx"

	return nil
}
