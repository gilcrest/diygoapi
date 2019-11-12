package datastore

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"github.com/gilcrest/go-api-basic/domain/errs"

	_ "github.com/lib/pq" // pq driver calls for blank identifier
)

// DBName defines database name
type DBName int

const (
	// AppDB represents main application database
	AppDB DBName = iota

	// LogDB represents http logging database
	LogDB
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

// ProvideDB returns an open database handle of 0 or more underlying connections
func ProvideDB(n DBName) (*sql.DB, error) {
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

func dbEnvName(n DBName) string {
	switch n {
	case AppDB:
		return envAppDBName
	case LogDB:
		return envLogDBName
	default:
		return ""
	}
}

func dbEnvUser(n DBName) string {
	switch n {
	case AppDB:
		return envAppDBUser
	case LogDB:
		return envLogDBUser
	default:
		return ""
	}
}

func dbEnvPassword(n DBName) string {
	switch n {
	case AppDB:
		return envAppDBPassword
	case LogDB:
		return envLogDBPassword
	default:
		return ""
	}
}

func dbEnvHost(n DBName) string {
	switch n {
	case AppDB:
		return envAppDBHost
	case LogDB:
		return envLogDBHost
	default:
		return ""
	}
}

func dbEnvPort(n DBName) string {
	switch n {
	case AppDB:
		return envAppDBPort
	case LogDB:
		return envLogDBPort
	default:
		return ""
	}
}
