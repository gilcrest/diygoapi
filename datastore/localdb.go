package datastore

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/gilcrest/go-api-basic/domain/errs"
)

// NewLocalDB returns an open database handle of 0 or more underlying PostgreSQL connections
func NewLocalDB(n DSName) (*sql.DB, func(), error) {
	const op errs.Op = "datastore/newLocalDB"

	// Get Database connection credentials from environment variables
	dbNme := os.Getenv(dbEnv(n, "dbname"))
	dbUser := os.Getenv(dbEnv(n, "user"))
	dbPassword := os.Getenv(dbEnv(n, "password"))
	dbHost := os.Getenv(dbEnv(n, "host"))
	dbPort, err := strconv.Atoi(os.Getenv(dbEnv(n, "port")))
	if err != nil {
		return nil, nil, errs.E(op, err)
	}

	// Craft string for database connection
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbNme)

	// Open the postgres database using the postgres driver (pq)
	// func Open(driverName, dataSourceName string) (*DB, error)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		return nil, nil, errs.E(op, err)
	}

	return db, func() { db.Close() }, nil
}

func dbEnv(n DSName, attr string) string {
	var localDS = map[string]string{
		"dbname":   "PG_APP_DBNAME",
		"user":     "PG_APP_USERNAME",
		"password": "PG_APP_PASSWORD",
		"host":     "PG_APP_HOST",
		"port":     "PG_APP_PORT",
	}

	var gcpCPDS = map[string]string{
		"dbname":   "PG_GCP_CP_DBNAME",
		"user":     "PG_GCP_CP_USERNAME",
		"password": "PG_GCP_CP_PASSWORD",
		"host":     "PG_GCP_CP_HOST",
		"port":     "PG_GCP_CP_PORT",
	}

	var envName string

	// In the below switch, I am deliberately skipping the boolean
	// to check for existence in the map since this is such a
	// controlled implementation, it would make the code less
	// readable to have to include error handling
	switch n {
	case LocalDatastore:
		envName, _ = localDS[attr]
	case GCPCPDatastore:
		envName, _ = gcpCPDS[attr]
	default:
		envName = ""
	}

	return envName
}
