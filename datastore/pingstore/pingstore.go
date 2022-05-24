// Package pingstore enables database health checks through the db
// Pool Ping method.
package pingstore

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/gilcrest/diy-go-api/domain/errs"
)

// PingDB pings the DB
func PingDB(ctx context.Context, pool *pgxpool.Pool) error {
	err := pool.Ping(ctx)
	if err != nil {
		return errs.E(errs.Database, err)
	}
	return err
}
