// Package pingstore enables database health checks through the db
// PingContext method. The Pinger interface allows mocking the Ping
// when testing with no db connectivity.
package pingstore

import (
	"context"
	"database/sql"
)

// Datastorer is an interface for working with the Database
type Datastorer interface {
	// DB returns a sql.DB
	DB() *sql.DB
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
	return d.DB().PingContext(ctx)
}
