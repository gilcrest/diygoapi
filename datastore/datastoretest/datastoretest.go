// Package datastoretest provides testing helper functions for the
// datastore package
package datastoretest

import (
	"database/sql"
	"os"
	"strconv"
	"testing"

	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/datastore"
)

// newPGDatasourceName is a test helper to get a PGDatasourceName
// from environment variables
func newPGDatasourceName(t *testing.T) datastore.PGDatasourceName {
	t.Helper()

	// Constants for the PostgreSQL Database connection
	const (
		pgDBHost     string = "PG_APP_HOST"
		pgDBPort     string = "PG_APP_PORT"
		pgDBName     string = "PG_APP_DBNAME"
		pgDBUser     string = "PG_APP_USERNAME"
		pgDBPassword string = "PG_APP_PASSWORD"
	)

	var (
		dbHost     string
		dbPort     int
		dbName     string
		dbUser     string
		dbPassword string
		ok         bool
		err        error
	)

	dbHost, ok = os.LookupEnv(pgDBHost)
	if !ok {
		t.Fatalf("No environment variable found for %s", pgDBHost)
	}

	p, ok := os.LookupEnv(pgDBPort)
	if !ok {
		t.Fatalf("No environment variable found for %s", pgDBPort)
	}
	dbPort, err = strconv.Atoi(p)
	if err != nil {
		t.Fatalf("Unable to convert db port %s to int", p)
	}

	dbName, ok = os.LookupEnv(pgDBName)
	if !ok {
		t.Fatalf("No environment variable found for %s", pgDBName)
	}

	dbUser, ok = os.LookupEnv(pgDBUser)
	if !ok {
		t.Fatalf("No environment variable found for %s", pgDBUser)
	}

	dbPassword, ok = os.LookupEnv(pgDBPassword)
	if !ok {
		t.Fatalf("No environment variable found for %s", pgDBPassword)
	}

	return datastore.NewPGDatasourceName(dbHost, dbName, dbUser, dbPassword, dbPort)
}

// NewDB provides a sql.DB and cleanup function for testing.
// The following environment variables must be set to connect to the DB.
//
// 		DB Host     = PG_APP_HOST
//		Port        = PG_APP_PORT
//		DB Name     = PG_APP_DBNAME
//		DB User     = PG_APP_USERNAME
//		DB Password = PG_APP_PASSWORD
func NewDB(t *testing.T, lgr zerolog.Logger) (*sql.DB, func()) {
	t.Helper()

	dsn := newPGDatasourceName(t)
	db, cleanup, err := datastore.NewDB(dsn, lgr)
	if err != nil {
		t.Fatalf("datastore.NewDB() error = %v", err)
	}
	return db, cleanup
}
