package datastore

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"errors"

	"github.com/gilcrest/errs"

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
	// GCPCPDatastore represents a local connection to a GCP Cloud
	// SQL db through the Google Cloud Proxy
	GCPCPDatastore
	// GCPDatastore represents a true GCP connection to a GCP
	// Cloud SQL db
	GCPDatastore
	// MockDatastore represents a Mocked Database
	MockDatastore
)

func (n DSName) String() string {
	switch n {
	case LocalDatastore:
		return "Local"
	case GCPCPDatastore:
		return "Google Cloud SQL through the Google Cloud Proxy"
	case GCPDatastore:
		return "Google Cloud SQL"
	case MockDatastore:
		return "Mock"
	}
	return "unknown_datastore_name"
}

// NewDatastore provides either a DS struct, which has a concrete
// implementation of a database or a MockDS struct which is a mocked
// DB implementation
func NewDatastore(n DSName, db *sql.DB) (Datastore, error) {
	const op errs.Op = "datastore/NewDatastore"

	switch n {
	case MockDatastore:
		return &MockDS{}, nil
	default:
		return &DS{DB: db}, nil
	}
}

func dbEnv(n DSName) (map[string]string, error) {
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
