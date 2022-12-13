package sqldbtest

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"

	"github.com/gilcrest/diygoapi/logger"
	"github.com/gilcrest/diygoapi/sqldb"
)

// newPGDatasourceName is a test helper to get a PGDatasourceName
// from environment variables
func newPostgreSQLDSN(t *testing.T) sqldb.PostgreSQLDSN {
	t.Helper()

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

	dbHost, ok = os.LookupEnv(sqldb.DBHostEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", sqldb.DBHostEnv)
	}

	var p string
	p, ok = os.LookupEnv(sqldb.DBPortEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", sqldb.DBPortEnv)
	}
	dbPort, err = strconv.Atoi(p)
	if err != nil {
		t.Fatalf("Unable to convert db port %s to int", p)
	}

	dbName, ok = os.LookupEnv(sqldb.DBNameEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", sqldb.DBNameEnv)
	}

	dbUser, ok = os.LookupEnv(sqldb.DBUserEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", sqldb.DBUserEnv)
	}

	dbPassword, ok = os.LookupEnv(sqldb.DBPasswordEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", sqldb.DBPasswordEnv)
	}

	dbSearchPath, ok = os.LookupEnv(sqldb.DBSearchPathEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", sqldb.DBSearchPathEnv)
	}

	return sqldb.PostgreSQLDSN{
		Host:       dbHost,
		Port:       dbPort,
		DBName:     dbName,
		SearchPath: dbSearchPath,
		User:       dbUser,
		Password:   dbPassword,
	}
}

// newPool provides a *pgxpool.Pool and cleanup function for testing.
// The following environment variables must be set to connect to the DB.
//
//	DB Host     = DB_HOST
//	Port        = DB_PORT
//	DB Name     = DB_NAME
//	DB User     = DB_USER
//	DB Password = DB_PASSWORD
func newPool(t *testing.T) (dbpool *pgxpool.Pool, cleanup func()) {
	t.Helper()

	ctx := context.Background()
	dsn := newPostgreSQLDSN(t)

	var err error
	// Open the postgres database using the postgres driver (pq)
	dbpool, err = pgxpool.Connect(ctx, dsn.KeywordValueConnectionString())
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}

	cleanup = func() { dbpool.Close() }

	return dbpool, cleanup
}

// NewDB provides a sqldb.DB struct
// initialized with a sql.DB and cleanup function for testing.
// The following environment variables must be set to connect to the DB.
//
//	DB Host        = DB_HOST
//	DB Port        = DB_PORT
//	DB Name        = DB_NAME
//	DB User        = DB_USER
//	DB Password    = DB_PASSWORD
//	DB Search Path = DB_SEARCH_PATH
func NewDB(t *testing.T) (*sqldb.DB, func()) {
	t.Helper()

	pool, cleanup := newPool(t)

	db := sqldb.NewDB(pool)

	lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)
	err := db.ValidatePool(context.Background(), lgr)
	if err != nil {
		t.Fatalf("db.ValidatePool() error = %v", err)
	}

	return db, cleanup
}
