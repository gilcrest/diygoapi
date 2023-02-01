// Package sqldb is used to interact with a datastore. It has
// functions to help set up a sql.DB as well as helpers for working
// with the sql.DB once it's initialized.
package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"

	"github.com/gilcrest/diygoapi/errs"
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
//
//	user[:password]
//
// and hostspec is:
//
//	[host][:port][,...]
//
// and paramspec is:
//
//	name=value[&...]
//
// The URI scheme designator can be either postgresql:// or postgres://.
// Each of the remaining URI parts is optional.
// The following examples illustrate valid URI syntax:
//
//	postgresql://
//	postgresql://localhost
//	postgresql://localhost:5433
//	postgresql://localhost/mydb
//	postgresql://user@localhost
//	postgresql://user:secret@localhost
//	postgresql://other@localhost/otherdb?connect_timeout=10&application_name=myapp
//	postgresql://host1:123,host2:456/somedb?target_session_attrs=any&application_name=myapp
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

// NewPostgreSQLPool creates a new db pool and establishes a connection.
// In addition, returns a Close function to defer closing the pool.
func NewPostgreSQLPool(ctx context.Context, lgr zerolog.Logger, dsn PostgreSQLDSN) (pool *pgxpool.Pool, close func(), err error) {
	const op errs.Op = "sqldb/NewPostgreSQLPool"

	f := func() {}

	// Open the postgres database using the pgxpool driver (pq)
	// func Open(driverName, dataSourceName string) (*DB, error)
	pool, err = pgxpool.Connect(ctx, dsn.KeywordValueConnectionString())
	if err != nil {
		return nil, f, errs.E(op, errs.Database, err)
	}

	lgr.Info().Msgf("sql database opened for %s on port %d", dsn.Host, dsn.Port)

	return pool, func() { pool.Close() }, nil
}

// DB is a concrete implementation for a PostgreSQL database
type DB struct {
	pool *pgxpool.Pool
}

// NewDB is an initializer for the DB struct
func NewDB(pool *pgxpool.Pool) *DB {
	return &DB{pool: pool}
}

// Ping pings the DB pool.
//
// From pgx: "Ping acquires a connection from the Pool and executes
// an empty sql statement against it. If the sql returns without error,
// the database Ping is considered successful, otherwise, the error is returned."
func (db *DB) Ping(ctx context.Context) error {
	const op errs.Op = "sqldb/DB.Ping"

	err := db.pool.Ping(ctx)
	if err != nil {
		return errs.E(op, errs.Database, err)
	}
	return err
}

// ValidatePool pings the database and logs the current user and database.
// ValidatePool is used for logging db status on startup.
func (db *DB) ValidatePool(ctx context.Context, log zerolog.Logger) error {
	const op errs.Op = "sqldb/DB.ValidatePool"

	err := db.pool.Ping(ctx)
	if err != nil {
		return errs.E(op, errs.Database, err)
	}
	log.Info().Msg("sql database Ping returned successfully")

	var (
		currentDatabase string
		currentUser     string
		dbVersion       string
		searchPath      string
	)
	sqlStatement := `select current_database(), current_user, version();`
	row := db.pool.QueryRow(ctx, sqlStatement)
	err = row.Scan(&currentDatabase, &currentUser, &dbVersion)
	switch {
	case err == sql.ErrNoRows:
		return errs.E(op, errs.Database, "no rows returned")
	case err != nil:
		return errs.E(op, errs.Database, err)
	default:
		log.Info().Msgf("database version: %s", dbVersion)
		log.Info().Msgf("current database user: %s", currentUser)
		log.Info().Msgf("current database: %s", currentDatabase)
	}

	searchPathSQL := `SHOW search_path;`
	searchPathRow := db.pool.QueryRow(ctx, searchPathSQL)
	err = searchPathRow.Scan(&searchPath)
	switch {
	case err == sql.ErrNoRows:
		return errs.E(op, errs.Database, "no rows returned for search_path")
	case err != nil:
		return errs.E(op, errs.Database, err)
	default:
		log.Info().Msgf("current search_path: %s", searchPath)
	}

	return nil
}

// BeginTx returns an acquired transaction from the db pool and
// adds app specific error handling
func (db *DB) BeginTx(ctx context.Context) (pgx.Tx, error) {
	const op errs.Op = "sqldb/DB.BeginTx"

	if db.pool == nil {
		return nil, errs.E(op, errs.Database, "db pool cannot be nil")
	}

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	return tx, nil
}

// RollbackTx is a wrapper for sql.Tx.Rollback in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (db *DB) RollbackTx(ctx context.Context, tx pgx.Tx, err error) error {
	const op errs.Op = "sqldb/DB.RollbackTx"

	if tx == nil {
		if err != nil {
			return errs.E(op, errs.Database, errs.Code("nil_tx"), fmt.Sprintf("RollbackTx() error = tx cannot be nil: Original error = %s", err.Error()))
		}
		return errs.E(op, errs.Database, errs.Code("nil_tx"), fmt.Sprintf("RollbackTx() error = tx cannot be nil: Original error is nil"))
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
			return errs.E(op, errs.Database, errs.Code("rollback_err"), fmt.Sprintf("PG Error Code: %s, PG Error Message: %s, RollbackTx() error = %v: Original error = %s", pgErr.Code, pgErr.Message, rollbackErr, err.Error()))
		}
		// in case it is somehow not a &pgconn.PgError type
		return errs.E(op, errs.Database, errs.Code("rollback_err"), fmt.Sprintf("RollbackTx() error = %v: Original error = %s", rollbackErr, err.Error()))
	}

	// If rollback was successful, send back original error
	return err
}

// CommitTx is a wrapper for sql.Tx.Commit in order to expose from
// the Datastore interface. Proper error handling is also considered.
func (db *DB) CommitTx(ctx context.Context, tx pgx.Tx) error {
	const op errs.Op = "sqldb/DB.CommitTx"

	if tx == nil {
		return errs.E(op, errs.Database, errs.Code("nil_tx"), "CommitTx() error = tx cannot be nil")
	}

	if err := tx.Commit(ctx); err != nil {
		return errs.E(op, errs.Database, err)
	}

	return nil
}
