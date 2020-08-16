package datastore

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/rs/zerolog"

	"github.com/gilcrest/errs"
)

// NewDB returns an open database handle of 0 or more underlying PostgreSQL connections
func NewDB(n Name, logger zerolog.Logger) (*sql.DB, func(), error) {
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
	var dbinfo string
	switch dbPassword {
	case "":
		dbinfo = fmt.Sprintf("host=%s port=%d dbname=%s user=%s sslmode=disable", dbHost, dbPort, dbNme, dbUser)
	default:
		dbinfo = fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable", dbHost, dbPort, dbNme, dbUser, dbPassword)
	}

	// Open the postgres database using the postgres driver (pq)
	// func Open(driverName, dataSourceName string) (*DB, error)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		return nil, nil, errs.E(op, err)
	}

	logger.Log().Msgf("sql database opened for %s on port %d", dbHost, dbPort)

	err = validateDB(db, logger)
	if err != nil {
		return nil, nil, errs.E(op, err)
	}

	return db, func() { db.Close() }, nil
}

// validateDB pings the database and logs the current user and database
func validateDB(db *sql.DB, log zerolog.Logger) error {
	const op errs.Op = "datastore/validateDB"

	err := db.Ping()
	if err != nil {
		return errs.E(op, err)
	}
	log.Log().Msg("sql database Ping returned successfully")

	var (
		currentDatabase string
		currentUser     string
		dbVersion       string
	)
	sqlStatement := `select current_database(), current_user, version();`
	row := db.QueryRow(sqlStatement)
	switch err := row.Scan(&currentDatabase, &currentUser, &dbVersion); err {
	case sql.ErrNoRows:
		return errs.E(op, "No rows were returned!")
	case nil:
		log.Log().Msgf("database version: %s", dbVersion)
		log.Log().Msgf("current database user: %s", currentUser)
		log.Log().Msgf("current database: %s", currentDatabase)
	default:
		return errs.E(op, err)
	}
	return nil
}
