package datastore

import (
	"database/sql"

	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/domain/errs"
)

// NewPostgreSQLDB returns an open database handle of 0 or more underlying PostgreSQL connections
func NewPostgreSQLDB(dsn PostgreSQLDSN, logger zerolog.Logger) (*sql.DB, func(), error) {

	f := func() {}

	// Open the postgres database using the postgres driver (pq)
	// func Open(driverName, dataSourceName string) (*DB, error)
	db, err := sql.Open("postgres", dsn.String())
	if err != nil {
		return nil, f, errs.E(errs.Database, err)
	}

	logger.Info().Msgf("sql database opened for %s on port %d", dsn.Host, dsn.Port)

	err = validatePostgreSQLDB(db, logger)
	if err != nil {
		return nil, f, err
	}

	return db, func() { db.Close() }, nil
}

// validatePostgreSQLDB pings the database and logs the current user and database
func validatePostgreSQLDB(db *sql.DB, log zerolog.Logger) error {
	err := db.Ping()
	if err != nil {
		return errs.E(errs.Database, err)
	}
	log.Info().Msg("sql database Ping returned successfully")

	var (
		currentDatabase string
		currentUser     string
		dbVersion       string
	)
	sqlStatement := `select current_database(), current_user, version();`
	row := db.QueryRow(sqlStatement)
	err = row.Scan(&currentDatabase, &currentUser, &dbVersion)
	switch {
	case err == sql.ErrNoRows:
		return errs.E(errs.Database, "no rows were returned")
	case err != nil:
		return errs.E(errs.Database, err)
	default:
		log.Info().Msgf("database version: %s", dbVersion)
		log.Info().Msgf("current database user: %s", currentUser)
		log.Info().Msgf("current database: %s", currentDatabase)
	}

	return nil
}
