package datastore

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/gilcrest/go-api-basic/domain/errs"

	_ "github.com/lib/pq" // pq driver calls for blank identifier
)

// ProvideDB returns an open database handle of 0 or more underlying connections
func provideDB(n DSName) (*sql.DB, error) {
	const op errs.Op = "datastore/provideDB"

	// If we are in "mock mode", we return a nil database
	if n == MockDatastore {
		return nil, nil
	}

	// Get Database connection credentials from environment variables
	dbNme := os.Getenv(dbEnvName(n))
	dbUser := os.Getenv(dbEnvUser(n))
	dbPassword := os.Getenv(dbEnvPassword(n))
	dbHost := os.Getenv(dbEnvHost(n))
	dbPort, err := strconv.Atoi(os.Getenv(dbEnvPort(n)))
	if err != nil {
		return nil, errs.E(op, err)
	}

	// Craft string for database connection
	dbinfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbNme)

	// Open the postgres database using the postgres driver (pq)
	// func Open(driverName, dataSourceName string) (*DB, error)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// Call Ping to validate the newly opened database is actually alive
	if err = db.Ping(); err != nil {
		return nil, errs.E(op, err)
	}
	return db, nil
}

func dbEnvName(n DSName) string {
	switch n {
	case AppDatastore:
		return envAppDBName
	case LogDatastore:
		return envLogDBName
	default:
		return ""
	}
}

func dbEnvUser(n DSName) string {
	switch n {
	case AppDatastore:
		return envAppDBUser
	case LogDatastore:
		return envLogDBUser
	default:
		return ""
	}
}

func dbEnvPassword(n DSName) string {
	switch n {
	case AppDatastore:
		return envAppDBPassword
	case LogDatastore:
		return envLogDBPassword
	default:
		return ""
	}
}

func dbEnvHost(n DSName) string {
	switch n {
	case AppDatastore:
		return envAppDBHost
	case LogDatastore:
		return envLogDBHost
	default:
		return ""
	}
}

func dbEnvPort(n DSName) string {
	switch n {
	case AppDatastore:
		return envAppDBPort
	case LogDatastore:
		return envLogDBPort
	default:
		return ""
	}
}
