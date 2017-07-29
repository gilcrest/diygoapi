// General database configuration package that allows for creating
// a database pool, starting a db transaction and adding said db
// transaction to a context and pulling the db transaction from the
// context
package db

import (
	"context"
	"database/sql"

	"github.com/gilcrest/go-API-template/pkg/config/env"
	_ "github.com/lib/pq"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

// dbTxKey is the context key for the database txn.  Its value of zero is
// arbitrary.  If this package defined other context keys, they would have
// different integer values.
const dbTxKey key = 0

// returns Context carrying the database transaction (tx).
func AddDBTx2Context(ctx context.Context, env *env.Env, opts *sql.TxOptions) context.Context {
	tx, _ := DbTx(ctx, env, opts)
	return context.WithValue(ctx, dbTxKey, tx)
}

// extracts the database transaction from the context, if present.
func DBTxFromContext(ctx context.Context) (*sql.Tx, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	// the *sql.Tx type assertion returns ok=false for nil.
	dbTx, ok := ctx.Value(dbTxKey).(*sql.Tx)
	return dbTx, ok
}

// Opens a database connection and starts a database transaction using the
// BeginTx method which allows for rollback of the transaction if the context
// is cancelled
func DbTx(ctx context.Context, env *env.Env, opts *sql.TxOptions) (*sql.Tx, error) {

	// Calls the BeginTx method of the above opened database
	// func (db *DB) BeginTx(ctx context.Context, opts *TxOptions) (*Tx, error)
	tx, err := env.Db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return tx, nil

}
