package datastore

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"

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
func NewDatastore(n DSName) (Datastore, error) {
	const op errs.Op = "datastore/NewDatastore"

	if n == MockDatastore {
		return &MockDS{}, nil
	}

	db, err := newDB(n)
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	return &DS{DB: db}, nil
}

// newDB returns an open database handle of 0 or more underlying connections
func newDB(n DSName) (*sql.DB, error) {
	const op errs.Op = "datastore/newDB"

	// If we are in "mock mode", we return a nil database
	if n == MockDatastore {
		return nil, nil
	}

	// Get Database connection credentials from environment variables
	dbNme := os.Getenv(dbEnvName(n))
	dbUser := os.Getenv(dbEnvUser(n))
	dbPassword := os.Getenv(dbEnvPassword(n))
	dbHost := os.Getenv(dbEnvHost(n))
	dbPort, err := strconv.Atoi(os.Getenv(dbEnvPort(n)))
	if err != nil {
		return nil, errs.E(op, err)
	}

	// Craft string for database connection
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbNme)

	// Open the postgres database using the postgres driver (pq)
	// func Open(driverName, dataSourceName string) (*DB, error)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// Call Ping to validate the newly opened database is actually alive
	if err = db.Ping(); err != nil {
		return nil, errs.E(op, err)
	}
	return db, nil
}

func dbEnvName(n DSName) string {
	switch n {
	case AppDatastore:
		return envAppDBName
	case LogDatastore:
		return envLogDBName
	default:
		return ""
	}
}

func dbEnvUser(n DSName) string {
	switch n {
	case AppDatastore:
		return envAppDBUser
	case LogDatastore:
		return envLogDBUser
	default:
		return ""
	}
}

func dbEnvPassword(n DSName) string {
	switch n {
	case AppDatastore:
		return envAppDBPassword
	case LogDatastore:
		return envLogDBPassword
	default:
		return ""
	}
}

func dbEnvHost(n DSName) string {
	switch n {
	case AppDatastore:
		return envAppDBHost
	case LogDatastore:
		return envLogDBHost
	default:
		return ""
	}
}

func dbEnvPort(n DSName) string {
	switch n {
	case AppDatastore:
		return envAppDBPort
	case LogDatastore:
		return envLogDBPort
	default:
		return ""
	}
}

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
