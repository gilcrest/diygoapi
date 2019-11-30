package movieds

import (
	"context"
	"database/sql"

	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/google/uuid"
	"github.com/rs/xid"
)

// MovieDS is the interface for the persistence layer for a movie
type MovieDS interface {
	Store(context.Context, *movie.Movie) error
	FindByID(context.Context, xid.ID) (*movie.Movie, error)
	FindAll(context.Context) ([]*movie.Movie, error)
	Update(context.Context, xid.ID, *movie.Movie) error
}

// NewMovieDS sets up either a concrete MovieDB or a MockMovieDB
// depending on the underlying struct of the Datastore passed in
func NewMovieDS(app *app.Application) (MovieDS, error) {
	const op errs.Op = "movieds/NewMovieDS"

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
func (mdb *MovieDB) Store(ctx context.Context, m *movie.Movie) error {
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
		m.ID,              //$1
		m.ExtlID.String(), //$2
		m.Title,           //$3
		m.Year,            //$4
		m.Rated,           //$5
		m.Released,        //$6
		m.RunTime,         //$7
		m.Director,        //$8
		m.Writer,          //$9
		fakeUserID,        //$10
		fakeUserID)        //$11

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
func (mdb *MovieDB) Update(context.Context, xid.ID, *movie.Movie) error {
	const op errs.Op = "movieds/MockMovieDB.Update"

	return nil
}

// FindByID returns a Movie struct to populate the response
func (mdb *MovieDB) FindByID(ctx context.Context, extlID xid.ID) (*movie.Movie, error) {
	const op errs.Op = "movieds/MovieDB.FindByID"

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
		  where extl_id = $1;`, extlID)

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
		return nil, errs.E(op, errs.NotExist, err)
	} else if err != nil {
		return nil, errs.E(op, err)
	}

	return m, nil
}

// FindAll returns a slice of Movie structs to populate the response
func (mdb *MovieDB) FindAll(ctx context.Context) ([]*movie.Movie, error) {
	const op errs.Op = "movieds/MockMovieDB.FindAll"

	// declare a slice of pointers to movie.Movie
	var s []*movie.Movie

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
		   from demo.movie m;`)
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}
	defer rows.Close()

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

		if err == sql.ErrNoRows {
			return nil, errs.E(op, errs.NotExist, err)
		} else if err != nil {
			return nil, errs.E(op, err)
		}

		s = append(s, m)
	}
	// Rows.Err will report the last error encountered by Rows.Scan.
	err = rows.Err()
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	// return the slice
	return s, nil
}
