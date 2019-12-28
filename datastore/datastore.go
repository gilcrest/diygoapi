package datastore

import (
	"context"
	"database/sql"

	"errors"

	"github.com/gilcrest/go-api-basic/domain/errs"

	_ "github.com/lib/pq" // pq driver calls for blank identifier
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
	// LocalDatastore represents the local PostgreSQL db
	LocalDatastore DSName = iota
	// LogDatastore represents http logging database
	LogDatastore
	// MockDatastore represents a Mocked Database
	MockDatastore
	// GCPCPDatastore represents a local connection to a GCP Cloud
	// SQL db via the Google Cloud Proxy
	GCPCPDatastore
)

func (n DSName) String() string {
	switch n {
	case LocalDatastore:
		return "Local"
	case LogDatastore:
		return "Logging"
	case MockDatastore:
		return "Mock"
	}
	return "unknown_datastore_name"
}

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

	// If rollback was successful, error passed in as parameter
	// should be an errs.Error type. If so, surface error Kind,
	// Code and Param to keep in tact
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

// NewDatastore provides either a DS struct, which has a concrete
// implementation of a database or a MockDS struct which is a mocked
// DB implementation
func NewDatastore(n DSName, db *sql.DB) (Datastore, error) {
	const op errs.Op = "datastore/NewDatastore"

	switch n {
	case MockDatastore:
		return &MockDS{}, nil
	case LocalDatastore:
		return &DS{DB: db}, nil
	default:
		return nil, errs.E(op, "Unknown Datastore")
	}
}
