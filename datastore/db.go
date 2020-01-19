package datastore

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/gilcrest/errs"
)

// NewDB returns an open database handle of 0 or more underlying PostgreSQL connections
func NewDB(n DSName) (*sql.DB, func(), error) {
	const op errs.Op = "datastore/NewDB"

	dbEnvMap, err := dbEnv(n)
	if err != nil {
		return nil, nil, errs.E(op, err)
	}

	// Get Database connection credentials from environment variables
	dbNme := dbEnvMap["dbname"]
	dbUser := dbEnvMap["user"]
	dbPassword := dbEnvMap["password"]
	dbHost := dbEnvMap["host"]
	dbPort, err := strconv.Atoi(dbEnvMap["port"])
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
