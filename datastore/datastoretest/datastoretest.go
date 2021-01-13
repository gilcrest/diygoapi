// Package datastoretest is for datastore helper functions
package datastoretest

import (
	"os"
	"strconv"
	"testing"

	"github.com/gilcrest/go-api-basic/datastore"
)

func NewPGDatasourceName(t *testing.T) datastore.PGDatasourceName {
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
		t.Errorf("No environment variable found for %s", pgDBHost)
	}

	p, ok := os.LookupEnv(pgDBPort)
	if !ok {
		t.Errorf("No environment variable found for %s", pgDBPort)
	}
	dbPort, err = strconv.Atoi(p)
	if err != nil {
		t.Errorf("Unable to convert db port %s to int", p)
	}

	dbName, ok = os.LookupEnv(pgDBName)
	if !ok {
		t.Errorf("No environment variable found for %s", pgDBName)
	}

	dbUser, ok = os.LookupEnv(pgDBUser)
	if !ok {
		t.Errorf("No environment variable found for %s", pgDBUser)
	}

	dbPassword, ok = os.LookupEnv(pgDBPassword)
	if !ok {
		t.Errorf("No environment variable found for %s", pgDBPassword)
	}

	return datastore.NewPGDatasourceName(dbHost, dbName, dbUser, dbPassword, dbPort)
}
