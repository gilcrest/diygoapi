package datastore

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/gilcrest/go-api-basic/domain/errs"

	_ "github.com/lib/pq" // pq driver calls for blank identifier
	"github.com/pkg/errors"
)

// Datastorer is an interface for working with the Database
type Datastorer interface {
	// DB returns a sql.DB
	DB() *sql.DB
	// BeginTx starts a sql.Tx using the input context
	BeginTx(context.Context) (*sql.Tx, error)
	// RollbackTx rolls back the input sql.Tx
	RollbackTx(*sql.Tx, error) error
	// CommitTx commits the Tx
	CommitTx(*sql.Tx) error
}

// Name defines the name for the Datastore
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

// PGDatasourceName is a Postgres datasource name
type PGDatasourceName struct {
	Host     string
	Port     int
	DBName   string
	User     string
	Password string
}

func (dsn PGDatasourceName) String() string {
	// Craft string for database connection
	switch dsn.Password {
	case "":
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=disable", dsn.Host, dsn.Port, dsn.DBName, dsn.User)
	default:
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable", dsn.Host, dsn.Port, dsn.DBName, dsn.User, dsn.Password)
	}
}

func NewPGDatasourceName(n Name) (PGDatasourceName, error) {
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

	var pgds PGDatasourceName

	switch n {
	case LocalDatastore:
		host, ok = os.LookupEnv(localDBHost)
		if !ok {
			return pgds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", localDBHost)))
		}
		port, ok = os.LookupEnv(localDBPort)
		if !ok {
			return pgds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", localDBPort)))
		}
		dbName, ok = os.LookupEnv(localDBName)
		if !ok {
			return pgds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", localDBName)))
		}
		user, ok = os.LookupEnv(localDBUser)
		if !ok {
			return pgds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", localDBUser)))
		}
		password, ok = os.LookupEnv(localDBPassword)
		if !ok {
			return pgds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", localDBPassword)))
		}
	case GCPCPDatastore:
		host, ok = os.LookupEnv(gcpCPDBHost)
		if !ok {
			return pgds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", gcpCPDBHost)))
		}
		port, ok = os.LookupEnv(gcpCPDBPort)
		if !ok {
			return pgds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", gcpCPDBPort)))
		}
		dbName, ok = os.LookupEnv(gcpCPDBName)
		if !ok {
			return pgds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", gcpCPDBName)))
		}
		user, ok = os.LookupEnv(gcpCPDBUser)
		if !ok {
			return pgds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", gcpCPDBUser)))
		}
		password, ok = os.LookupEnv(gcpCPDBPassword)
		if !ok {
			return pgds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", gcpCPDBPassword)))
		}
	case GCPDatastore:
		host, ok = os.LookupEnv(gcpDBHost)
		if !ok {
			return pgds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", gcpDBHost)))
		}
		port, ok = os.LookupEnv(gcpDBPort)
		if !ok {
			return pgds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", gcpDBPort)))
		}
		dbName, ok = os.LookupEnv(gcpDBName)
		if !ok {
			return pgds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", gcpDBName)))
		}
		user, ok = os.LookupEnv(gcpDBUser)
		if !ok {
			return pgds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", gcpDBUser)))
		}
		password, ok = os.LookupEnv(gcpDBPassword)
		if !ok {
			return pgds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", gcpDBPassword)))
		}
	default:
		return pgds, errs.E(errors.New("Unrecognized Name"))
	}

	dbPort, err := strconv.Atoi(port)
	if err != nil {
		return pgds, errs.E(errors.New(fmt.Sprintf("Unable to convert db port %s to int", port)))
	}

	pgds = PGDatasourceName{
		DBName:   dbName,
		User:     user,
		Password: password,
		Host:     host,
		Port:     dbPort,
	}

	return pgds, nil
}

func NewDatastore(db *sql.DB) *Datastore {
	return &Datastore{db: db}
}

// Datastore is a concrete implementation for a sql database
type Datastore struct {
	db *sql.DB
}

func (ds *Datastore) DB() *sql.DB {
	return ds.db
}

// BeginTx is a wrapper for sql.DB.BeginTx in order to expose from
// the Datastore interface
func (ds *Datastore) BeginTx(ctx context.Context) (*sql.Tx, error) {
	if ds.db == nil {
		return nil, errs.E(errs.Database, errors.New("DB cannot be nil"))
	}

	tx, err := ds.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errs.E(errs.Database, err)
	}

	return tx, nil
}

// RollbackTx is a wrapper for sql.Tx.Rollback in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (ds *Datastore) RollbackTx(tx *sql.Tx, err error) error {
	if tx == nil {
		return errs.E(errs.Database, errors.New("tx cannot be nil"))
	}

	// Attempt to rollback the transaction
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		return errs.E(errs.Database, rollbackErr)
	}

	// If rollback was successful, send back original error
	return errs.E(errs.Database, err)
}

// CommitTx is a wrapper for sql.Tx.Commit in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (ds *Datastore) CommitTx(tx *sql.Tx) error {
	if err := tx.Commit(); err != nil {
		return errs.E(errs.Database, err)
	}

	return nil
}

// NewNullString returns a null if s is empty, otherwise it returns
// the string which was input
func NewNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

// NewNullInt64 returns a null if i == 0, otherwise it returns
// the int64 which was input
func NewNullInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{
		Int64: i,
		Valid: true,
	}
}
