// Package datastore is used to interact with a datastore. It has
// functions to help set up a sql.DB as well as helpers for working
// with the sql.DB once it's initialized.
package datastore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/gilcrest/diy-go-api/domain/errs"
)

const (
	// DBHostEnv is the database host environment variable name
	DBHostEnv string = "DB_HOST"
	// DBPortEnv is the database port environment variable name
	DBPortEnv string = "DB_PORT"
	// DBNameEnv is the database name environment variable name
	DBNameEnv string = "DB_NAME"
	// DBUserEnv is the database user environment variable name
	DBUserEnv string = "DB_USER"
	// DBPasswordEnv is the database user password environment variable name
	DBPasswordEnv string = "DB_PASSWORD"
	// DBSearchPathEnv is the database search path environment variable name
	DBSearchPathEnv string = "DB_SEARCH_PATH"
)

// PostgreSQLDSN is a PostgreSQL datasource name
type PostgreSQLDSN struct {
	Host       string
	Port       int
	DBName     string
	SearchPath string
	User       string
	Password   string
}

// ConnectionURI returns a formatted PostgreSQL datasource "Keyword/Value Connection String"
// The general form for a connection URI is:
// postgresql://[userspec@][hostspec][/dbname][?paramspec]
// where userspec is
//     user[:password]
// and hostspec is:
//     [host][:port][,...]
// and paramspec is:
//     name=value[&...]
// The URI scheme designator can be either postgresql:// or postgres://.
// Each of the remaining URI parts is optional.
// The following examples illustrate valid URI syntax:
//    postgresql://
//    postgresql://localhost
//    postgresql://localhost:5433
//    postgresql://localhost/mydb
//    postgresql://user@localhost
//    postgresql://user:secret@localhost
//    postgresql://other@localhost/otherdb?connect_timeout=10&application_name=myapp
//    postgresql://host1:123,host2:456/somedb?target_session_attrs=any&application_name=myapp
func (dsn PostgreSQLDSN) ConnectionURI() string {

	const uriSchemeDesignator string = "postgresql"

	var h string
	h = dsn.Host
	if dsn.Port != 0 {
		h += ":" + strconv.Itoa(dsn.Port)
	}

	u := url.URL{
		Scheme: uriSchemeDesignator,
		User:   url.User(dsn.User),
		Host:   h,
		Path:   dsn.DBName,
	}

	if dsn.SearchPath != "" {
		q := u.Query()
		q.Set("options", fmt.Sprintf("-csearch_path=%s", dsn.SearchPath))
		u.RawQuery = q.Encode()
	}

	return u.String()
}

// KeywordValueConnectionString returns a formatted PostgreSQL datasource "Keyword/Value Connection String"
func (dsn PostgreSQLDSN) KeywordValueConnectionString() string {

	var s string

	// if db connection does not have a password (should only be for local testing and preferably never),
	// the password parameter must be removed from the string, otherwise the connection will fail.
	switch dsn.Password {
	case "":
		s = fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=disable", dsn.Host, dsn.Port, dsn.DBName, dsn.User)
	default:
		s = fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable", dsn.Host, dsn.Port, dsn.DBName, dsn.User, dsn.Password)
	}

	// if search path needs to be explicitly set, will be added to the end of the datasource string
	switch dsn.SearchPath {
	case "":
		return s
	default:
		return s + " " + fmt.Sprintf("search_path=%s", dsn.SearchPath)
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
		if err != nil {
			return errs.E(errs.Database, errs.Code("nil_tx"), fmt.Sprintf("RollbackTx() error = tx cannot be nil: Original error = %s", err.Error()))
		}
		return errs.E(errs.Database, errs.Code("nil_tx"), fmt.Sprintf("RollbackTx() error = tx cannot be nil: Original error is nil"))
	}

	// Attempt to roll back the transaction
	if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
		// if the transaction has already been closed, the transaction
		// has already been committed or rolled back. In this case, there
		// is nothing to do and is not considered a new error. Send back
		// the original error (err)
		if errors.Is(rollbackErr, pgx.ErrTxClosed) {
			return err
		}
		// any other error should be reported, it should be a *pgconn.PgError type
		var pgErr *pgconn.PgError
		if errors.As(rollbackErr, &pgErr) {
			return errs.E(errs.Database, errs.Code("rollback_err"), fmt.Sprintf("PG Error Code: %s, PG Error Message: %s, RollbackTx() error = %v: Original error = %s", pgErr.Code, pgErr.Message, rollbackErr, err.Error()))
		}
		// in case it is somehow not a &pgconn.PgError type
		return errs.E(errs.Database, errs.Code("rollback_err"), fmt.Sprintf("RollbackTx() error = %v: Original error = %s", rollbackErr, err.Error()))
	}

	// If rollback was successful, send back original error
	return err
}

// CommitTx is a wrapper for sql.Tx.Commit in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (ds Datastore) CommitTx(ctx context.Context, tx pgx.Tx) error {
	if tx == nil {
		return errs.E(errs.Database, errs.Code("nil_tx"), "CommitTx() error = tx cannot be nil")
	}

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

// NewNullTime returns a null if t is the zero value for time.Time,
// otherwise it returns the time which was input
func NewNullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{}
	}
	return sql.NullTime{
		Time:  t,
		Valid: true,
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

// NewNullInt32 returns a null if i == 0, otherwise it returns
// the int32 which was input.
func NewNullInt32(i int32) sql.NullInt32 {
	if i == 0 {
		return sql.NullInt32{}
	}
	return sql.NullInt32{
		Int32: i,
		Valid: true,
	}
}
