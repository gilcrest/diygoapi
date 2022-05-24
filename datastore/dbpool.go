package datastore

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"

	"github.com/gilcrest/diy-go-api/domain/errs"
)

// NewPostgreSQLPool returns an open database handle of 0 or more underlying PostgreSQL connections
func NewPostgreSQLPool(ctx context.Context, dsn PostgreSQLDSN, logger zerolog.Logger) (*pgxpool.Pool, func(), error) {

	f := func() {}

	// Open the postgres database using the pgxpool driver (pq)
	// func Open(driverName, dataSourceName string) (*DB, error)
	pool, err := pgxpool.Connect(ctx, dsn.KeywordValueConnectionString())
	if err != nil {
		return nil, f, errs.E(errs.Database, err)
	}

	logger.Info().Msgf("sql database opened for %s on port %d", dsn.Host, dsn.Port)

	err = ValidatePostgreSQLPool(ctx, pool, logger)
	if err != nil {
		return nil, f, err
	}

	return pool, func() { pool.Close() }, nil
}

// ValidatePostgreSQLPool pings the database and logs the current user and database
func ValidatePostgreSQLPool(ctx context.Context, pool *pgxpool.Pool, log zerolog.Logger) error {
	err := pool.Ping(ctx)
	if err != nil {
		return errs.E(errs.Database, err)
	}
	log.Info().Msg("sql database Ping returned successfully")

	var (
		currentDatabase string
		currentUser     string
		dbVersion       string
		searchPath      string
	)
	sqlStatement := `select current_database(), current_user, version();`
	row := pool.QueryRow(ctx, sqlStatement)
	err = row.Scan(&currentDatabase, &currentUser, &dbVersion)
	switch {
	case err == sql.ErrNoRows:
		return errs.E(errs.Database, "no rows returned")
	case err != nil:
		return errs.E(errs.Database, err)
	default:
		log.Info().Msgf("database version: %s", dbVersion)
		log.Info().Msgf("current database user: %s", currentUser)
		log.Info().Msgf("current database: %s", currentDatabase)
	}

	searchPathSQL := `SHOW search_path;`
	searchPathRow := pool.QueryRow(ctx, searchPathSQL)
	err = searchPathRow.Scan(&searchPath)
	switch {
	case err == sql.ErrNoRows:
		return errs.E(errs.Database, "no rows returned for search_path")
	case err != nil:
		return errs.E(errs.Database, err)
	default:
		log.Info().Msgf("current search_path: %s", searchPath)
	}

	return nil
}
