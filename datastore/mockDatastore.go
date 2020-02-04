package datastore

import (
	"context"
	"database/sql"

	"github.com/gilcrest/errs"
)

func NewMockDatastore() *MockDatastore {
	return &MockDatastore{}
}

// MockDatastore is a mock implementation for a database
type MockDatastore struct {
}

func (mds *MockDatastore) DB() *sql.DB {
	return nil
}

// BeginTx is a wrapper for sql.DB.BeginTx in order to expose from
// the Datastore interface
func (mds *MockDatastore) BeginTx(_ context.Context) (*sql.Tx, error) {
	return nil, nil
}

// Tx exposes the Tx stored in the struct in order to be exposed from
// the Datastore interface.
func (mds *MockDatastore) Tx() (*sql.Tx, error) {
	return nil, nil
}

// RollbackTx is a wrapper for sql.Tx.Rollback in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (mds *MockDatastore) RollbackTx(_ *sql.Tx, err error) error {
	const op errs.Op = "datastore/MockDS.RollbackTx"

	return errs.E(op, errs.Database, err)
}

// CommitTx is a wrapper for sql.Tx.Commit in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (mds *MockDatastore) CommitTx(_ *sql.Tx) error {
	return nil
}
