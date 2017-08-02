package db

import (
	"context"
	"database/sql"

	"github.com/gilcrest/go-API-template/pkg/config/env"
)

// Tx opens a database connection and starts a database transaction using the
// BeginTx method which allows for rollback of the transaction if the context
// is cancelled
func Tx(ctx context.Context, env *env.Env, opts *sql.TxOptions) (*sql.Tx, error) {

	// Calls the BeginTx method of the above opened database
	// func (db *DB) BeginTx(ctx context.Context, opts *TxOptions) (*Tx, error)
	tx, err := env.Db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return tx, nil

}
