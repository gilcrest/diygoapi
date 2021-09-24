// Package pingstore enables database health checks through the db
// Pool Ping method. The Pinger interface allows mocking the Ping
// when testing with no db connectivity.
package pingstore

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Datastorer is an interface for working with the Database
type Datastorer interface {
	// Pool returns *pgxpool.Pool
	Pool() *pgxpool.Pool
}

// Pinger is the default implementation for pinging the db
type Pinger struct {
	Datastorer
}

// NewPinger is an initializer for DefaultPinger
func NewPinger(ds Datastorer) Pinger {
	return Pinger{ds}
}

// PingDB pings the DB
func (d Pinger) PingDB(ctx context.Context) error {
	return d.Pool().Ping(ctx)
}
