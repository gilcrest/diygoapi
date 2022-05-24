package datastore

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/gilcrest/diy-go-api/domain/logger"

	qt "github.com/frankban/quicktest"
	"github.com/rs/zerolog"
)

func TestNewDatastore(t *testing.T) {
	t.Run("typical", func(t *testing.T) {
		c := qt.New(t)

		ctx := context.Background()
		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)

		dsn := newPostgreSQLDSN(t)

		dbpool, cleanup, err := NewPostgreSQLPool(ctx, dsn, lgr)
		c.Assert(err, qt.IsNil)
		t.Cleanup(cleanup)

		got := NewDatastore(dbpool)
		want := Datastore{dbpool: dbpool}

		c.Assert(got, qt.Equals, want)
	})
}

// newPGDatasourceName is a test helper to get a PGDatasourceName
// from environment variables
func newPostgreSQLDSN(t *testing.T) PostgreSQLDSN {
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

	dbHost, ok = os.LookupEnv(DBHostEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", DBHostEnv)
	}

	var p string
	p, ok = os.LookupEnv(DBPortEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", DBPortEnv)
	}
	dbPort, err = strconv.Atoi(p)
	if err != nil {
		t.Fatalf("Unable to convert db port %s to int", p)
	}

	dbName, ok = os.LookupEnv(DBNameEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", DBNameEnv)
	}

	dbUser, ok = os.LookupEnv(DBUserEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", DBUserEnv)
	}

	dbPassword, ok = os.LookupEnv(DBPasswordEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", DBPasswordEnv)
	}

	dbSearchPath, ok = os.LookupEnv(DBSearchPathEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", DBSearchPathEnv)
	}

	return PostgreSQLDSN{
		Host:       dbHost,
		Port:       dbPort,
		DBName:     dbName,
		SearchPath: dbSearchPath,
		User:       dbUser,
		Password:   dbPassword,
	}
}
