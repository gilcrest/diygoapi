package moviestore

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/google/uuid"
)

// Transactor performs DML actions against the DB
type Transactor interface {
	Create(ctx context.Context, m *movie.Movie) error
	Update(ctx context.Context, m *movie.Movie) error
	Delete(ctx context.Context, m *movie.Movie) error
}

// Selector reads records from the db
type Selector interface {
	FindByID(context.Context, string) (*movie.Movie, error)
	FindAll(context.Context) ([]*movie.Movie, error)
}

// NewTx is an initializer for Tx
func NewTx(tx *sql.Tx) (*Tx, error) {
	if tx == nil {
		return nil, errs.E(errors.New(errs.MissingField("tx").Error()))
	}
	return &Tx{Tx: tx}, nil
}

// Tx is the database implementation for DML operations for a movie
type Tx struct {
	*sql.Tx
}

// Create inserts a record in the user table using a stored function
func (t *Tx) Create(ctx context.Context, m *movie.Movie) error {
	// Prepare the sql statement using bind variables
	stmt, err := t.Tx.PrepareContext(ctx, `
	select o_create_timestamp,
		   o_update_timestamp
	  from demo.create_movie (
		p_id => $1,
		p_extl_id => $2,
		p_title => $3,
		p_rated => $4,
		p_released => $5,
		p_run_time => $6,
		p_director => $7,
		p_writer => $8,
		p_create_client_id => $9,
		p_create_username => $10)`)

	if err != nil {
		return errs.E(errs.Database, err)
	}
	defer stmt.Close()

	// At some point, I will add a whole client flow, but for now
	// faking a client uuid....
	fakeClientID := uuid.New()

	// Execute stored function that returns the create_date timestamp,
	// hence the use of QueryContext instead of Exec
	rows, err := stmt.QueryContext(ctx,
		m.ID,               //$1
		m.ExternalID,       //$2
		m.Title,            //$3
		m.Rated,            //$4
		m.Released,         //$5
		m.RunTime,          //$6
		m.Director,         //$7
		m.Writer,           //$8
		fakeClientID,       //$9
		m.CreateUser.Email) //$10

	if err != nil {
		return errs.E(errs.Database, err)
	}
	defer rows.Close()

	// Iterate through the returned record(s)
	for rows.Next() {
		if err := rows.Scan(&m.CreateTime, &m.UpdateTime); err != nil {
			return errs.E(errs.Database, err)
		}
	}

	// If any error was encountered while iterating through rows.Next above
	// it will be returned here
	if err := rows.Err(); err != nil {
		return errs.E(errs.Database, err)
	}

	return nil
}

// Update updates a record in the database using the external ID of
// the Movie
func (t *Tx) Update(ctx context.Context, m *movie.Movie) error {
	// Prepare the sql statement using bind variables
	stmt, err := t.Tx.PrepareContext(ctx, `
	update demo.movie
	   set title = $1,
		   rated = $2,
		   released = $3,
		   run_time = $4,
		   director = $5,
		   writer = $6,
		   update_username = $7,
		   update_timestamp = $8
	 where extl_id = $9
 returning movie_id, create_username, create_timestamp`)

	if err != nil {
		return errs.E(errs.Database, err)
	}
	defer stmt.Close()

	// Execute stored function that returns the create_date timestamp,
	// hence the use of QueryContext instead of Exec
	rows, err := stmt.QueryContext(ctx,
		m.Title,            //$1
		m.Rated,            //$2
		m.Released,         //$3
		m.RunTime,          //$4
		m.Director,         //$5
		m.Writer,           //$6
		m.UpdateUser.Email, //$7
		m.UpdateTime,       //$8
		m.ExternalID)       //$9

	if err != nil {
		return errs.E(errs.Database, err)
	}
	defer rows.Close()

	// Iterate through the returned record(s)
	for rows.Next() {
		if err := rows.Scan(&m.ID, &m.CreateUser.Email, &m.CreateTime); err != nil {
			return errs.E(errs.Database, err)
		}
	}

	// If any error was encountered while iterating through rows.Next above
	// it will be returned here
	if err := rows.Err(); err != nil {
		return errs.E(errs.Database, err)
	}

	// If the table's primary key is not returned as part of the
	// RETURNING clause, this means the row was not actually updated.
	// The update request does not contain this key (I don't believe
	// in exposing primary keys), so this is a way of returning data
	// from an update statement and checking whether or not the
	// update was actually successful. Typically you would use
	// db.Exec and check RowsAffected (like I do in delete below),
	// but I wanted to show an alternative which can be useful here
	if m.ID == uuid.Nil {
		return errs.E(errs.Database, "Invalid ID - no records updated")
	}

	return nil
}

// Delete removes the Movie record from the table
func (t *Tx) Delete(ctx context.Context, m *movie.Movie) error {
	result, execErr := t.Tx.ExecContext(ctx,
		`DELETE from demo.movie
		        WHERE movie_id = $1`, m.ID)

	if execErr != nil {
		return errs.E(errs.Database, execErr)
	}

	// Only 1 row should be deleted, check the result count to
	// ensure this is correct
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errs.E(errs.Database, err)
	}
	if rowsAffected == 0 {
		return errs.E(errs.Database, "No Rows Deleted")
	} else if rowsAffected > 1 {
		return errs.E(errs.Database, "Too Many Rows Deleted")
	}

	return nil
}

// NewDB is an initializer for DB
func NewDB(db *sql.DB) (*DB, error) {
	if db == nil {
		return nil, errs.E(errs.Database, errors.New(errs.MissingField("db").Error()))
	}
	return &DB{DB: db}, nil
}

// DB is the database implementation for READ operations for a movie
type DB struct {
	*sql.DB
}

// FindByID returns a Movie struct to populate the response
func (d *DB) FindByID(ctx context.Context, extlID string) (*movie.Movie, error) {
	// Prepare the sql statement using bind variables
	row := d.DB.QueryRowContext(ctx,
		`select movie_id,
				extl_id,
				title,
				rated,
				released,
				run_time,
				director,
				writer,
				create_username,
				create_timestamp,
				update_username,
				update_timestamp
		   from demo.movie m
		  where extl_id = $1`, extlID)

	m := new(movie.Movie)
	err := row.Scan(
		&m.ID,
		&m.ExternalID,
		&m.Title,
		&m.Rated,
		&m.Released,
		&m.RunTime,
		&m.Director,
		&m.Writer,
		&m.CreateUser.Email,
		&m.CreateTime,
		&m.UpdateUser.Email,
		&m.UpdateTime)

	if err == sql.ErrNoRows {
		return nil, errs.E(errs.NotExist, "No record found for given ID")
	} else if err != nil {
		return nil, errs.E(errs.Database, err)
	}

	return m, nil
}

// FindAll returns a slice of Movie structs to populate the response
func (d *DB) FindAll(ctx context.Context) ([]*movie.Movie, error) {
	// use QueryContext to get back sql.Rows
	rows, err := d.DB.QueryContext(ctx,
		`select movie_id,
					  extl_id,
					  title,
					  rated,
					  released,
					  run_time,
					  director,
					  writer,
					  create_username,
					  create_timestamp,
					  update_username,
					  update_timestamp
				 from demo.movie m`)
	if err != nil {
		return nil, errs.E(errs.Database, err)
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
			&m.ExternalID,
			&m.Title,
			&m.Rated,
			&m.Released,
			&m.RunTime,
			&m.Director,
			&m.Writer,
			&m.CreateUser.Email,
			&m.CreateTime,
			&m.UpdateUser.Email,
			&m.UpdateTime)

		if err != nil {
			return nil, errs.E(errs.Database, err)
		}

		s = append(s, m)
	}

	// If the database is being written to ensure to check for Close
	// errors that may be returned from the driver. The query may
	// encounter an auto-commit error and be forced to rollback changes.
	rerr := rows.Close()
	if rerr != nil {
		return nil, errs.E(errs.Database, err)
	}

	// Rows.Err will report the last error encountered by Rows.Scan.
	err = rows.Err()
	if err != nil {
		return nil, errs.E(errs.Database, err)
	}

	// Determine if slice has not been populated. In this case, return
	// an error as we should receive rows
	if len(s) == 0 {
		return nil, errs.E(errs.Validation, errors.New("No rows returned"))
	}

	// return the slice
	return s, nil
}
