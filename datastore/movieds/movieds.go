package movieds

import (
	"context"
	"database/sql"

	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/rs/zerolog"
)

// MovieDS is the interface for the persistence layer for a movie
type MovieDS interface {
	Store(context.Context, *movie.Movie, *audit.Audit) error
}

// MovieDB is the database implementation for CRUD operations for a movie
type MovieDB struct {
	Tx  *sql.Tx
	Log zerolog.Logger
}

// Store creates a record in the user table using a stored function
func (mdb *MovieDB) Store(ctx context.Context, m *movie.Movie, a *audit.Audit) error {
	const op errs.Op = "movie/Movie.createDB"

	// Prepare the sql statement using bind variables
	stmt, err := mdb.Tx.PrepareContext(ctx, `
	select o_create_timestamp,
		   o_update_timestamp
	  from demo.create_movie (
		p_title => $1,
		p_year => $2,
		p_rated => $3,
		p_released => $4,
		p_run_time => $5,
		p_director => $6,
		p_writer => $7,
		p_create_client_id => $8,
		p_create_user_id => $9)`)

	if err != nil {
		return errs.E(op, err)
	}
	defer stmt.Close()

	// Execute stored function that returns the create_date timestamp,
	// hence the use of QueryContext instead of Exec
	rows, err := stmt.QueryContext(ctx,
		m.Title,          //$1
		m.Year,           //$2
		m.Rated,          //$3
		m.Released,       //$4
		m.RunTime,        //$5
		m.Director,       //$6
		m.Writer,         //$7
		a.CreateClientID, //$8
		a.CreatePersonID) //$9

	if err != nil {
		return errs.E(op, err)
	}
	defer rows.Close()

	// Iterate through the returned record(s)
	for rows.Next() {
		if err := rows.Scan(&a.CreateTimestamp, &a.UpdateTimestamp); err != nil {
			return errs.E(op, err)
		}
	}

	// If any error was encountered while iterating through rows.Next above
	// it will be returned here
	if err := rows.Err(); err != nil {
		return errs.E(op, err)
	}

	return nil
}

// ProvideMovieDS takes in a Datastore, pulls a txn from it
// and returns
func ProvideMovieDS(tx *sql.Tx, log zerolog.Logger) MovieDS {
	return &MovieDB{Tx: tx, Log: log}
}
