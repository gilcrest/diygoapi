package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var (
	// DBCon is the connection handle for the database
	DBCon *sql.DB
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

// dbTxKey is the context key for the database txn.  Its value of zero is
// arbitrary.  If this package defined other context keys, they would have
// different integer values.
const dbTxKey key = 0

// returns Context carrying the database transaction (tx).
func AddDBTx2Context(ctx context.Context, opts *sql.TxOptions) context.Context {
	tx, _ := DbTx(ctx, opts)
	return context.WithValue(ctx, dbTxKey, tx)
}

// extracts the database transaction from the context, if present.
func DBTxFromContext(ctx context.Context) (*sql.Tx, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	// the *sql.Tx type assertion returns ok=false for nil.
	dbTx, ok := ctx.Value(dbTxKey).(*sql.Tx)
	return dbTx, ok
}

// returns an open database handle of 0 or more underlying connections
func NewDB() (*sql.DB, error) {

	// Get Database connection credentials from environment variables
	DB_NAME := os.Getenv("PG_DBNAME")
	DB_USER := os.Getenv("PG_USERNAME")
	DB_PASSWORD := os.Getenv("PG_PASSWORD")

	// Craft string for database connection
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", DB_USER, DB_PASSWORD, DB_NAME)

	// Open the postgres database using the postgres driver (pq)
	// func Open(driverName, dataSourceName string) (*DB, error)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

// Opens a database connection and starts a database transaction using the
// BeginTx method which allows for rollback of the transaction if the context
// is cancelled
func DbTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {

	// Calls the BeginTx method of the above opened database
	// func (db *DB) BeginTx(ctx context.Context, opts *TxOptions) (*Tx, error)
	tx, err := DBCon.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return tx, nil

}
