// Package datastore TODO
package datastore

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq" // pq driver calls for blank identifier
)

// Datastore struct stores common environment related items
type Datastore struct {
	MainDb *sql.DB
	LogDb  *sql.DB
}

// NewDatastore initializes the datastore struct
func NewDatastore() (*Datastore, error) {
	db, err := newMainDB()
	if err != nil {
		return nil, err
	}
	return &Datastore{MainDb: db, LogDb: nil}, nil
}

// NewMainDB returns an open database handle of 0 or more underlying connections
func newMainDB() (*sql.DB, error) {

	// Get Database connection credentials from environment variables
	dbName := os.Getenv("PG_DBNAME")
	dbUser := os.Getenv("PG_USERNAME")
	dbPassword := os.Getenv("PG_PASSWORD")

	// Craft string for database connection
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", dbUser, dbPassword, dbName)

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

// Tx opens a database connection and starts a database transaction using the
// BeginTx method which allows for rollback of the transaction if the context
// is cancelled
func (ds Datastore) Tx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {

	// Calls the BeginTx method of the MainDb opened database
	tx, err := ds.MainDb.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return tx, nil

}

// // The key type is unexported to prevent collisions with context keys defined in
// // other packages.
// type key int

// // dbTxKey is the context key for the database txn.  Its value of zero is
// // arbitrary.  If this package defined other context keys, they would have
// // different integer values.
// const dbTxKey key = 0

// // Tx2Context returns Context carrying the database transaction (tx).
// func Tx2Context(ctx context.Context, opts *sql.TxOptions) context.Context {
// 	tx, _ := Tx(ctx, db, opts)
// 	return context.WithValue(ctx, dbTxKey, tx)
// }

// // TxFromContext extracts the database transaction from the context, if present.
// func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
// 	// ctx.Value returns nil if ctx has no value for the key;
// 	// the *sql.Tx type assertion returns ok=false for nil.
// 	dbTx, ok := ctx.Value(dbTxKey).(*sql.Tx)
// 	return dbTx, ok
// }
