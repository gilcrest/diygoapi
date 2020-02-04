package datastore

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/gilcrest/errs"

	_ "github.com/lib/pq" // pq driver calls for blank identifier
)

// Datastorer is an interface for working with the Database
type Datastorer interface {
	DB() *sql.DB
	BeginTx(context.Context) (*sql.Tx, error)
	RollbackTx(*sql.Tx, error) error
	CommitTx(*sql.Tx) error
}

// DSName defines the name for the Datastore
type Name int

const (
	// LocalDatastore represents the local PostgreSQL db
	LocalDatastore Name = iota
	// GCPCPDatastore represents a local connection to a GCP Cloud
	// SQL db through the Google Cloud Proxy
	GCPCPDatastore
	// GCPDatastore represents a true GCP connection to a GCP
	// Cloud SQL db
	GCPDatastore
	// MockedDatastore represents a Mocked Database
	MockedDatastore
)

func (n Name) String() string {
	switch n {
	case LocalDatastore:
		return "Local"
	case GCPCPDatastore:
		return "Google Cloud SQL through the Google Cloud Proxy"
	case GCPDatastore:
		return "Google Cloud SQL"
	case MockedDatastore:
		return "Mock"
	}
	return "unknown_datastore_name"
}

func dbEnv(n Name) (map[string]string, error) {
	const op errs.Op = "datastore/dbEnv"

	// Constants for the local PostgreSQL Database connection
	const (
		localDBHost     string = "PG_APP_HOST"
		localDBPort     string = "PG_APP_PORT"
		localDBName     string = "PG_APP_DBNAME"
		localDBUser     string = "PG_APP_USERNAME"
		localDBPassword string = "PG_APP_PASSWORD"
	)

	// Constants for the local PostgreSQL Google Cloud Proxy Database
	// connection
	const (
		gcpCPDBHost     string = "PG_GCP_CP_HOST"
		gcpCPDBPort     string = "PG_GCP_CP_PORT"
		gcpCPDBName     string = "PG_GCP_CP_DBNAME"
		gcpCPDBUser     string = "PG_GCP_CP_USERNAME"
		gcpCPDBPassword string = "PG_GCP_CP_PASSWORD"
	)

	// Constants for the GCP Cloud SQL Connection
	const (
		gcpDBHost     string = "PG_GCP_HOST"
		gcpDBPort     string = "PG_GCP_PORT"
		gcpDBName     string = "PG_GCP_DBNAME"
		gcpDBUser     string = "PG_GCP_USERNAME"
		gcpDBPassword string = "PG_GCP_PASSWORD"
	)

	var (
		ok       bool
		dbName   string
		user     string
		password string
		host     string
		port     string
	)

	switch n {
	case LocalDatastore:
		host, ok = os.LookupEnv(localDBHost)
		if !ok {
			return nil, errs.E(op, fmt.Sprintf("No environment variable found for %s", localDBHost))
		}
		port, ok = os.LookupEnv(localDBPort)
		if !ok {
			return nil, errs.E(op, fmt.Sprintf("No environment variable found for %s", localDBPort))
		}
		dbName, ok = os.LookupEnv(localDBName)
		if !ok {
			return nil, errs.E(op, fmt.Sprintf("No environment variable found for %s", localDBName))
		}
		user, ok = os.LookupEnv(localDBUser)
		if !ok {
			return nil, errs.E(op, fmt.Sprintf("No environment variable found for %s", localDBUser))
		}
		password, ok = os.LookupEnv(localDBPassword)
		if !ok {
			return nil, errs.E(op, fmt.Sprintf("No environment variable found for %s", localDBPassword))
		}
	case GCPCPDatastore:
		host, ok = os.LookupEnv(gcpCPDBHost)
		if !ok {
			return nil, errs.E(op, fmt.Sprintf("No environment variable found for %s", gcpCPDBHost))
		}
		port, ok = os.LookupEnv(gcpCPDBPort)
		if !ok {
			return nil, errs.E(op, fmt.Sprintf("No environment variable found for %s", gcpCPDBPort))
		}
		dbName, ok = os.LookupEnv(gcpCPDBName)
		if !ok {
			return nil, errs.E(op, fmt.Sprintf("No environment variable found for %s", gcpCPDBName))
		}
		user, ok = os.LookupEnv(gcpCPDBUser)
		if !ok {
			return nil, errs.E(op, fmt.Sprintf("No environment variable found for %s", gcpCPDBUser))
		}
		password, ok = os.LookupEnv(gcpCPDBPassword)
		if !ok {
			return nil, errs.E(op, fmt.Sprintf("No environment variable found for %s", gcpCPDBPassword))
		}
	case GCPDatastore:
		host, ok = os.LookupEnv(gcpDBHost)
		if !ok {
			return nil, errs.E(op, fmt.Sprintf("No environment variable found for %s", gcpDBHost))
		}
		port, ok = os.LookupEnv(gcpDBPort)
		if !ok {
			return nil, errs.E(op, fmt.Sprintf("No environment variable found for %s", gcpDBPort))
		}
		dbName, ok = os.LookupEnv(gcpDBName)
		if !ok {
			return nil, errs.E(op, fmt.Sprintf("No environment variable found for %s", gcpDBName))
		}
		user, ok = os.LookupEnv(gcpDBUser)
		if !ok {
			return nil, errs.E(op, fmt.Sprintf("No environment variable found for %s", gcpDBUser))
		}
		password, ok = os.LookupEnv(gcpDBPassword)
		if !ok {
			return nil, errs.E(op, fmt.Sprintf("No environment variable found for %s", gcpDBPassword))
		}
	default:
		return nil, errs.E(op, "Unrecognized DSName")
	}

	dbEnvMap := map[string]string{
		"dbname":   dbName,
		"user":     user,
		"password": password,
		"host":     host,
		"port":     port}

	return dbEnvMap, nil
}

func NewDatastore(db *sql.DB) *Datastore {
	return &Datastore{db: db}
}

// Datastore is a concrete implementation for a database
type Datastore struct {
	db *sql.DB
}

func (ds *Datastore) DB() *sql.DB {
	return ds.db
}

// BeginTx is a wrapper for sql.DB.BeginTx in order to expose from
// the Datastore interface
func (ds *Datastore) BeginTx(ctx context.Context) (*sql.Tx, error) {
	const op errs.Op = "datastore/Datastore.BeginTx"

	if ds.db == nil {
		return nil, errs.E(op, errs.Database, "DB cannot be nil")
	}

	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	return tx, nil
}

// RollbackTx is a wrapper for sql.Tx.Rollback in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (ds *Datastore) RollbackTx(tx *sql.Tx, err error) error {
	const op errs.Op = "datastore/Datastore.RollbackTx"

	if tx == nil {
		return errs.E(op, errs.Database, "tx cannot be nil")
	}

	// Attempt to rollback the transaction
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		return errs.E(op, errs.Database, err)
	}

	// If rollback was successful, send back original error
	return errs.E(op, errs.Database, err)
}

// CommitTx is a wrapper for sql.Tx.Commit in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (ds *Datastore) CommitTx(tx *sql.Tx) error {
	const op errs.Op = "datastore/Datastore.CommitTx"

	if err := tx.Commit(); err != nil {
		return errs.E(op, errs.Database, err)
	}

	return nil
}
