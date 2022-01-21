// Package datastoretest provides testing helper functions for the
// datastore package. I am intentionally repeating functions that
// are in datastore here as I want different versions as helpers
// with less logging
package datastoretest

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/domain/logger"
)

// newPGDatasourceName is a test helper to get a PGDatasourceName
// from environment variables
func newPostgreSQLDSN(t *testing.T) datastore.PostgreSQLDSN {
	t.Helper()

	// Constants for the PostgreSQL Database connection
	const (
		pgDBHost       string = "DB_HOST"
		pgDBPort       string = "DB_PORT"
		pgDBName       string = "DB_NAME"
		pgDBUser       string = "DB_USER"
		pgDBPassword   string = "DB_PASSWORD"
		pgDBSearchPath string = "DB_SEARCH_PATH"
	)

	var (
		dbHost       string
		dbPort       int
		dbName       string
		dbUser       string
		dbPassword   string
		dbSearchPath string
		ok           bool
		err          error
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

	dbSearchPath, ok = os.LookupEnv(pgDBSearchPath)
	if !ok {
		t.Fatalf("No environment variable found for %s", pgDBSearchPath)
	}

	return datastore.PostgreSQLDSN{
		Host:       dbHost,
		Port:       dbPort,
		DBName:     dbName,
		SearchPath: dbSearchPath,
		User:       dbUser,
		Password:   dbPassword,
	}
}

// newDB provides a *pgxpool.Pool and cleanup function for testing.
// The following environment variables must be set to connect to the DB.
//
// 		DB Host     = DB_HOST
//		Port        = DB_PORT
//		DB Name     = DB_NAME
//		DB User     = DB_USER
//		DB Password = DB_PASSWORD
func newDB(t *testing.T) (dbpool *pgxpool.Pool, cleanup func()) {
	t.Helper()

	ctx := context.Background()
	dsn := newPostgreSQLDSN(t)

	var err error
	// Open the postgres database using the postgres driver (pq)
	dbpool, err = pgxpool.Connect(ctx, dsn.KeywordValueConnectionString())
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}

	lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)

	err = datastore.ValidatePostgreSQLPool(ctx, dbpool, lgr)
	if err != nil {
		t.Fatalf("datastore.ValidatePostgreSQLPool() error = %v", err)
	}

	cleanup = func() { dbpool.Close() }

	return dbpool, cleanup
}

// NewDatastore provides a datastore.Datastore struct
// initialized with a sql.DB and cleanup function for testing.
// The following environment variables must be set to connect to the DB.
//
// 		DB Host        = DB_HOST
//		Port           = DB_PORT
//		DB Name        = DB_NAME
//		DB User        = DB_USER
//		DB Password    = DB_PASSWORD
//		DB Search Path = DB_SEARCH_PATH
func NewDatastore(t *testing.T) (datastore.Datastore, func()) {
	t.Helper()

	db, cleanup := newDB(t)

	return datastore.NewDatastore(db), cleanup
}
