package movieds

import (
	"context"

	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"
)

// MovieDS is the interface for the persistence layer for a movie
type MovieDS interface {
	Store(context.Context, *movie.Movie, *audit.Audit) error
}

// ProvideMovieDS sets up either a concrete MovieDB or a MockMovieDB
// depending on the underlying struct of the Datastore passed in
func ProvideMovieDS(app *app.Application) (MovieDS, error) {
	const op errs.Op = "movieds/ProvideMovieDS"

	// Use a type switch to determine if the app datastore is a Mock
	// Datastore, if so, then return MockMovieDB, otherwise use
	// composition to add the Datastore to the MovieDB struct
	switch ds := app.DS.(type) {
	case *datastore.MockDS:
		return &MockMovieDB{}, nil
	case *datastore.DS:
		return &MovieDB{DS: ds}, nil
	default:
		return nil, errs.E(op, "Unknown type for datastore.Datastore")
	}
}

// MovieDB is the database implementation for CRUD operations for a movie
type MovieDB struct {
	*datastore.DS
}

// Store creates a record in the user table using a stored function
func (mdb *MovieDB) Store(ctx context.Context, m *movie.Movie, a *audit.Audit) error {
	const op errs.Op = "movie/Movie.createDB"

	// Prepare the sql statement using bind variables
	stmt, err := mdb.Tx.PrepareContext(ctx, `
	select o_create_timestamp,
		   o_update_timestamp
	  from demo.create_movie (
		p_id => $1,
		p_title => $2,
		p_year => $3,
		p_rated => $4,
		p_released => $5,
		p_run_time => $6,
		p_director => $7,
		p_writer => $8,
		p_create_client_id => $9,
		p_create_user_id => $10)`)

	if err != nil {
		return errs.E(op, err)
	}
	defer stmt.Close()

	// Execute stored function that returns the create_date timestamp,
	// hence the use of QueryContext instead of Exec
	rows, err := stmt.QueryContext(ctx,
		m.ID,             //$1
		m.Title,          //$2
		m.Year,           //$3
		m.Rated,          //$4
		m.Released,       //$5
		m.RunTime,        //$6
		m.Director,       //$7
		m.Writer,         //$8
		a.CreateClientID, //$9
		a.CreatePersonID) //$10

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
