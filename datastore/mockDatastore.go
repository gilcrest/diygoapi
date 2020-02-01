package datastore

import (
	"context"
	"database/sql"

	"github.com/gilcrest/errs"
)

// NewMockDatastore returns a nil sql.DB
func NewMockDatastore(_ Name) *sql.DB {
	return nil
}

// MockDatastore is a mock implementation for a database
type MockDatastore struct {
}

// BeginTx is a wrapper for sql.DB.BeginTx in order to expose from
// the Datastore interface
func (db *MockDatastore) BeginTx(_ context.Context) error {
	return nil
}

// Tx exposes the Tx stored in the struct in order to be exposed from
// the Datastore interface.
func (db *MockDatastore) Tx() (*sql.Tx, error) {
	return nil, nil
}

// RollbackTx is a wrapper for sql.Tx.Rollback in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (db *MockDatastore) RollbackTx(err error) error {
	const op errs.Op = "datastore/MockDS.RollbackTx"

	return errs.E(op, errs.Database, err)
}

// CommitTx is a wrapper for sql.Tx.Commit in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (db *MockDatastore) CommitTx() error {
	return nil
}
