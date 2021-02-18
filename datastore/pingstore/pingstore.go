// Package pingstore enables database health checks through the db
// PingContext method. The Pinger interface allows mocking the Ping
// when testing with no db connectivity.
package pingstore

import (
	"context"

	"github.com/gilcrest/go-api-basic/datastore"
)

// Pinger pings the database
type Pinger interface {
	PingDB(context.Context) error
}

// NewDefaultPinger is an initializer for DefaultPinger
func NewDefaultPinger(ds datastore.Datastorer) DefaultPinger {
	return DefaultPinger{ds}
}

// DefaultPinger is the default implementation for pinging the db
type DefaultPinger struct {
	datastore.Datastorer
}

// PingDB pings the DB
func (d DefaultPinger) PingDB(ctx context.Context) error {
	return d.DB().PingContext(ctx)
}
