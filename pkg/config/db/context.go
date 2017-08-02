package db

import (
	"context"
	"database/sql"

	"github.com/gilcrest/go-API-template/pkg/config/env"
)

// The key type is unexported to prevent collisions with context keys defined in
// other packages.
type key int

// dbTxKey is the context key for the database txn.  Its value of zero is
// arbitrary.  If this package defined other context keys, they would have
// different integer values.
const dbTxKey key = 0

// Tx2Context returns Context carrying the database transaction (tx).
func Tx2Context(ctx context.Context, env *env.Env, opts *sql.TxOptions) context.Context {
	tx, _ := Tx(ctx, env, opts)
	return context.WithValue(ctx, dbTxKey, tx)
}

// TxFromContext extracts the database transaction from the context, if present.
func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
	// ctx.Value returns nil if ctx has no value for the key;
	// the *sql.Tx type assertion returns ok=false for nil.
	dbTx, ok := ctx.Value(dbTxKey).(*sql.Tx)
	return dbTx, ok
}
