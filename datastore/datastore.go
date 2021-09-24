// Package datastore is used to interact with a datastore. It has
// functions to help set up a sql.DB as well as helpers for working
// with the sql.DB once it's initialized.
package datastore

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/gilcrest/go-api-basic/domain/errs"
)

// PostgreSQLDSN is a PostgreSQL datasource name
type PostgreSQLDSN struct {
	Host     string
	Port     int
	DBName   string
	User     string
	Password string
}

// NewPostgreSQLDSN is an initializer for PostgreSQLDSN
func NewPostgreSQLDSN(host, dbname, user, password string, port int) PostgreSQLDSN {
	return PostgreSQLDSN{
		DBName:   dbname,
		User:     user,
		Password: password,
		Host:     host,
		Port:     port,
	}
}

// String returns a formatted PostgreSQL datasource name. If you are
// using a local db with no password, it removes the password from the
// string, otherwise the connection will fail.
func (dsn PostgreSQLDSN) String() string {
	// Craft string for database connection
	switch dsn.Password {
	case "":
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=disable", dsn.Host, dsn.Port, dsn.DBName, dsn.User)
	default:
		return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable", dsn.Host, dsn.Port, dsn.DBName, dsn.User, dsn.Password)
	}
}

// Datastore is a concrete implementation for a sql database
type Datastore struct {
	dbpool *pgxpool.Pool
}

// NewDatastore is an initializer for the Datastore struct
func NewDatastore(dbpool *pgxpool.Pool) Datastore {
	return Datastore{dbpool: dbpool}
}

// Pool returns *pgxpool.Pool from the Datastore struct
func (ds Datastore) Pool() *pgxpool.Pool {
	return ds.dbpool
}

// BeginTx returns an acquired transaction from the db pool and
// adds app specific error handling
func (ds Datastore) BeginTx(ctx context.Context) (pgx.Tx, error) {
	if ds.dbpool == nil {
		return nil, errs.E(errs.Database, "db pool cannot be nil")
	}

	tx, err := ds.dbpool.Begin(ctx)
	if err != nil {
		return nil, errs.E(errs.Database, err)
	}

	return tx, nil
}

// RollbackTx is a wrapper for sql.Tx.Rollback in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (ds Datastore) RollbackTx(ctx context.Context, tx pgx.Tx, err error) error {
	if tx == nil {
		return errs.E(errs.Database, errs.Code("nil_tx"), fmt.Sprintf("RollbackTx() error = tx cannot be nil: Original error = %s", err.Error()))
	}

	// Attempt to rollback the transaction
	if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
		return errs.E(errs.Database, errs.Code("rollback_err"), fmt.Sprintf("RollbackTx() error = %v: Original error = %s", rollbackErr, err.Error()))
	}

	// If rollback was successful, send back original error
	return err
}

// CommitTx is a wrapper for sql.Tx.Commit in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (ds Datastore) CommitTx(ctx context.Context, tx pgx.Tx) error {
	if err := tx.Commit(ctx); err != nil {
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
// the int64 which was input.
func NewNullInt64(i int64) sql.NullInt64 {
	if i == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{
		Int64: i,
		Valid: true,
	}
}
