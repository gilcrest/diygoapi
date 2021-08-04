// Package pingstore enables database health checks through the db
// PingContext method. The Pinger interface allows mocking the Ping
// when testing with no db connectivity.
package pingstore

import (
	"context"

	"github.com/gilcrest/go-api-basic/datastore"
)

// NewPinger is an initializer for DefaultPinger
func NewPinger(ds datastore.Datastorer) Pinger {
	return Pinger{ds}
}

// Pinger is the default implementation for pinging the db
type Pinger struct {
	datastore.Datastorer
}

// PingDB pings the DB
func (d Pinger) PingDB(ctx context.Context) error {
	return d.DB().PingContext(ctx)
}
