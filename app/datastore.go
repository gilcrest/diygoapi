package main

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/gilcrest/go-api-basic/domain/errs"

	_ "github.com/lib/pq" // pq driver calls for blank identifier
)

// dbName defines database name
type dbName int

const (
	// AppDB represents main application database
	appDB dbName = iota

	// LogDB represents http logging database
	logDB
)

// OS Environment variables for the App DB PostgreSQL Database
const (
	envAppDBName     = "PG_APP_DBNAME"
	envAppDBUser     = "PG_APP_USERNAME"
	envAppDBPassword = "PG_APP_PASSWORD"
	envAppDBHost     = "PG_APP_HOST"
	envAppDBPort     = "PG_APP_PORT"
)

// OS Environment variables for the Log DB PostgreSQL Database
const (
	envLogDBName     = "PG_LOG_DBNAME"
	envLogDBUser     = "PG_LOG_USERNAME"
	envLogDBPassword = "PG_LOG_PASSWORD"
	envLogDBHost     = "PG_LOG_HOST"
	envLogDBPort     = "PG_LOG_PORT"
)

// ProvideDatastore initializes the datastore struct
func provideAppDB() (*sql.DB, error) {
	const op errs.Op = "main/provideAppDB"

	// Get an AppDB (PostgreSQL)
	adb, err := newDB(appDB)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return adb, nil
}

// newDB returns an open database handle of 0 or more underlying connections
func newDB(n dbName) (*sql.DB, error) {
	const op errs.Op = "main/newDB"

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

func dbEnvName(n dbName) string {
	switch n {
	case appDB:
		return envAppDBName
	case logDB:
		return envLogDBName
	default:
		return ""
	}
}

func dbEnvUser(n dbName) string {
	switch n {
	case appDB:
		return envAppDBUser
	case logDB:
		return envLogDBUser
	default:
		return ""
	}
}

func dbEnvPassword(n dbName) string {
	switch n {
	case appDB:
		return envAppDBPassword
	case logDB:
		return envLogDBPassword
	default:
		return ""
	}
}

func dbEnvHost(n dbName) string {
	switch n {
	case appDB:
		return envAppDBHost
	case logDB:
		return envLogDBHost
	default:
		return ""
	}
}

func dbEnvPort(n dbName) string {
	switch n {
	case appDB:
		return envAppDBPort
	case logDB:
		return envLogDBPort
	default:
		return ""
	}
}
