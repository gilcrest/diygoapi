package movieDatastore

import (
	"context"
	"database/sql"

	"github.com/gilcrest/errs"
	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/google/uuid"
)

// MovieDS is the interface for the persistence layer for a movie
type MovieDatastorer interface {
	Create(context.Context, *movie.Movie) error
	FindByID(context.Context, string) (*movie.Movie, error)
	FindAll(context.Context) ([]*movie.Movie, error)
	Update(context.Context, *movie.Movie) error
	Delete(context.Context, *movie.Movie) error
}

// NewMovieDS sets up either a concrete MovieDB or a MockMovieDB
// depending on the underlying struct of the Datastore passed in
func NewMovieDatastorer(app *app.Application) (MovieDatastorer, error) {
	const op errs.Op = "movieDatastore/NewMovieDatastorer"

	// Use a type switch to determine if the app datastore is a Mock
	// Datastore, if so, then return MockMovieDB, otherwise use
	// composition to add the Datastore to the MovieDB struct
	switch ds := app.Datastorer.(type) {
	case *datastore.MockDatastore:
		return &MockMovieDB{}, nil
	case *datastore.Datastore:
		return &MovieDatastore{Datastore: ds}, nil
	default:
		return nil, errs.E(op, "Unknown type for datastore.Datastorer")
	}
}

// MovieDB is the database implementation for CRUD operations for a movie
type MovieDatastore struct {
	*datastore.Datastore
}

// Create inserts a record in the user table using a stored function
func (mdb *MovieDatastore) Create(ctx context.Context, m *movie.Movie) error {
	const op errs.Op = "movie/Movie.createDB"

	// Prepare the sql statement using bind variables
	stmt, err := mdb.Tx.PrepareContext(ctx, `
	select o_create_timestamp,
		   o_update_timestamp
	  from demo.create_movie (
		p_id => $1,
		p_extl_id => $2,
		p_title => $3,
		p_year => $4,
		p_rated => $5,
		p_released => $6,
		p_run_time => $7,
		p_director => $8,
		p_writer => $9,
		p_create_client_id => $10,
		p_create_user_id => $11)`)

	if err != nil {
		return errs.E(op, err)
	}
	defer stmt.Close()

	// At some point, I will add a whole user flow, but for now
	// faking a user uuid....
	fakeUserID := uuid.New()

	// Execute stored function that returns the create_date timestamp,
	// hence the use of QueryContext instead of Exec
	rows, err := stmt.QueryContext(ctx,
		m.ID,       //$1
		m.ExtlID,   //$2
		m.Title,    //$3
		m.Year,     //$4
		m.Rated,    //$5
		m.Released, //$6
		m.RunTime,  //$7
		m.Director, //$8
		m.Writer,   //$9
		fakeUserID, //$10
		fakeUserID) //$11

	if err != nil {
		return errs.E(op, err)
	}
	defer rows.Close()

	// Iterate through the returned record(s)
	for rows.Next() {
		if err := rows.Scan(&m.CreateTimestamp, &m.UpdateTimestamp); err != nil {
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

// Update updates a record in the database using the external ID of
// the Movie
func (mdb *MovieDatastore) Update(ctx context.Context, m *movie.Movie) error {
	const op errs.Op = "movieDatastore/MockMovieDB.Update"

	// Prepare the sql statement using bind variables
	stmt, err := mdb.Tx.PrepareContext(ctx, `
	update demo.movie
	   set title = $1,
		   year = $2,
		   rated = $3,
		   released = $4,
		   run_time = $5,
		   director = $6,
		   writer = $7,
		   update_user_id = $8,
		   update_timestamp = $9
	 where extl_id = $10
 returning movie_id, create_timestamp`)

	if err != nil {
		return errs.E(op, err)
	}
	defer stmt.Close()

	// At some point, I will add a whole user flow, but for now
	// faking a user uuid....
	fakeUserID := uuid.New()

	// Execute stored function that returns the create_date timestamp,
	// hence the use of QueryContext instead of Exec
	rows, err := stmt.QueryContext(ctx,
		m.Title,           //$1
		m.Year,            //$2
		m.Rated,           //$3
		m.Released,        //$4
		m.RunTime,         //$5
		m.Director,        //$6
		m.Writer,          //$7
		fakeUserID,        //$8
		m.UpdateTimestamp, //$9
		m.ExtlID)          //$10

	if err != nil {
		return errs.E(op, err)
	}
	defer rows.Close()

	// Iterate through the returned record(s)
	for rows.Next() {
		if err := rows.Scan(&m.ID, &m.CreateTimestamp); err != nil {
			return errs.E(op, err)
		}
	}

	// If any error was encountered while iterating through rows.Next above
	// it will be returned here
	if err := rows.Err(); err != nil {
		return errs.E(op, err)
	}

	// If the table's primary key is not returned as part of the
	// RETURNING clause, this means the row was not actually updated.
	// The update request does not contain this key (I don't believe
	// in exposing primary keys), so this is a way of returning data
	// from an update statement and checking whether or not the
	// update was actually successful. Typically you would use
	// db.Exec and check RowsAffected (like I do in delete), but I
	// wanted to show an alternative which can be useful here
	if m.ID == uuid.Nil {
		return errs.E(op, errs.Database, "Invalid ID - no records updated")
	}

	return nil
}

// FindByID returns a Movie struct to populate the response
func (mdb *MovieDatastore) FindByID(ctx context.Context, extlID string) (*movie.Movie, error) {
	const op errs.Op = "movieDatastore/MovieDB.FindByID"

	// Prepare the sql statement using bind variables
	row := mdb.DB.QueryRowContext(ctx,
		`select movie_id,
				extl_id,
				title,
				year,
				rated,
				released,
				run_time,
				director,
				writer,
				create_timestamp,
				update_timestamp
		   from demo.movie m
		  where extl_id = $1`, extlID)

	m := new(movie.Movie)
	err := row.Scan(
		&m.ID,
		&m.ExtlID,
		&m.Title,
		&m.Year,
		&m.Rated,
		&m.Released,
		&m.RunTime,
		&m.Director,
		&m.Writer,
		&m.CreateTimestamp,
		&m.UpdateTimestamp)

	if err == sql.ErrNoRows {
		return nil, errs.E(op, errs.NotExist, "No record found for given ID")
	} else if err != nil {
		return nil, errs.E(op, err)
	}

	return m, nil
}

// FindAll returns a slice of Movie structs to populate the response
func (mdb *MovieDatastore) FindAll(ctx context.Context) ([]*movie.Movie, error) {
	const op errs.Op = "movieDatastore/MovieDB.FindAll"

	// use QueryContext to get back sql.Rows
	rows, err := mdb.DB.QueryContext(ctx,
		`select movie_id,
				extl_id,
				title,
				year,
				rated,
				released,
				run_time,
				director,
				writer,
				create_timestamp,
				update_timestamp
		   from demo.movie m`)
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}
	defer rows.Close()
	// declare a slice of pointers to movie.Movie
	// var s []*movie.Movie
	s := make([]*movie.Movie, 0)

	// iterate through each row and scan the results into
	// a movie.Movie. Append movie.Movie to the slice
	// defined above
	for rows.Next() {
		m := new(movie.Movie)
		err = rows.Scan(
			&m.ID,
			&m.ExtlID,
			&m.Title,
			&m.Year,
			&m.Rated,
			&m.Released,
			&m.RunTime,
			&m.Director,
			&m.Writer,
			&m.CreateTimestamp,
			&m.UpdateTimestamp)

		if err != nil {
			return nil, errs.E(op, errs.Database, err)
		}

		s = append(s, m)
	}

	// If the database is being written to ensure to check for Close
	// errors that may be returned from the driver. The query may
	// encounter an auto-commit error and be forced to rollback changes.
	rerr := rows.Close()
	if rerr != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	// Rows.Err will report the last error encountered by Rows.Scan.
	err = rows.Err()
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	// Determine if slice has not been populated. In this case, return
	// an error as we should receive rows
	if len(s) == 0 {
		return nil, errs.E(op, errs.Validation, "No rows returned")
	}

	// return the slice
	return s, nil
}

// Delete removes the Movie record from the table
func (mdb *MovieDatastore) Delete(ctx context.Context, m *movie.Movie) error {
	const op errs.Op = "movie/MovieDB.Delete"

	result, execErr := mdb.Tx.ExecContext(ctx,
		`DELETE from demo.movie
		  WHERE movie_id = $1`, m.ID)

	if execErr != nil {
		return errs.E(op, errs.Database, execErr)
	}

	// Only 1 row should be deleted, check the result count to
	// ensure this is correct
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errs.E(op, errs.Database, err)
	}
	if rowsAffected == 0 {
		return errs.E(op, errs.Database, "No Rows Deleted")
	} else if rowsAffected > 1 {
		return errs.E(op, errs.Database, "Too Many Rows Deleted")
	}

	return nil
}
