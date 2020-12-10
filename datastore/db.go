package datastore

import (
	"database/sql"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/domain/errs"
)

// NewDB returns an open database handle of 0 or more underlying PostgreSQL connections
func NewDB(pgds PGDatasourceName, logger zerolog.Logger) (*sql.DB, func(), error) {

	// Open the postgres database using the postgres driver (pq)
	// func Open(driverName, dataSourceName string) (*DB, error)
	db, err := sql.Open("postgres", pgds.String())
	if err != nil {
		return nil, nil, errs.E(err)
	}

	logger.Log().Msgf("sql database opened for %s on port %d", pgds.Host, pgds.Port)

	err = validateDB(db, logger)
	if err != nil {
		return nil, nil, err
	}

	return db, func() { db.Close() }, nil
}

// validateDB pings the database and logs the current user and database
func validateDB(db *sql.DB, log zerolog.Logger) error {
	err := db.Ping()
	if err != nil {
		return errs.E(err)
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
		return errs.E(errors.New("No rows were returned!"))
	case nil:
		log.Log().Msgf("database version: %s", dbVersion)
		log.Log().Msgf("current database user: %s", currentUser)
		log.Log().Msgf("current database: %s", currentDatabase)
	default:
		return errs.E(err)
	}
	return nil
}
