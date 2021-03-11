package moviestore

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"
)

// Transactor performs DML actions against the DB
type Transactor interface {
	Create(ctx context.Context, m *movie.Movie) error
	Update(ctx context.Context, m *movie.Movie) error
	Delete(ctx context.Context, m *movie.Movie) error
}

// NewDefaultTransactor is an initializer for DefaultMovieStore
func NewDefaultTransactor(ds datastore.Datastorer) DefaultTransactor {
	return DefaultTransactor{ds}
}

// DefaultTransactor is the default database implementation
// for DML operations for a movie
type DefaultTransactor struct {
	datastorer datastore.Datastorer
}

// Create inserts a record in the user table using a stored function
func (dt DefaultTransactor) Create(ctx context.Context, m *movie.Movie) error {
	tx, err := dt.datastorer.BeginTx(ctx)
	if err != nil {
		return err
	}

	// Prepare the sql statement using bind variables
	stmt, err := tx.PrepareContext(ctx, `
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
		return errs.E(errs.Database, dt.datastorer.RollbackTx(tx, err))
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
		return errs.E(errs.Database, dt.datastorer.RollbackTx(tx, err))
	}
	defer rows.Close()

	// Iterate through the returned record(s)
	for rows.Next() {
		if err := rows.Scan(&m.CreateTime, &m.UpdateTime); err != nil {
			return errs.E(errs.Database, dt.datastorer.RollbackTx(tx, err))
		}
	}

	// If any error was encountered while iterating through rows.Next above
	// it will be returned here
	if err := rows.Err(); err != nil {
		return errs.E(errs.Database, dt.datastorer.RollbackTx(tx, err))
	}

	// Commit the Transaction
	if err := dt.datastorer.CommitTx(tx); err != nil {
		return errs.E(errs.Database, dt.datastorer.RollbackTx(tx, err))
	}

	return nil
}

// Update updates a record in the database using the external ID of
// the Movie
func (dt DefaultTransactor) Update(ctx context.Context, m *movie.Movie) error {
	tx, err := dt.datastorer.BeginTx(ctx)
	if err != nil {
		return err
	}

	// Prepare the sql statement using bind variables
	stmt, err := tx.PrepareContext(ctx, `
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
		return errs.E(errs.Database, dt.datastorer.RollbackTx(tx, err))
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
		return errs.E(errs.Database, dt.datastorer.RollbackTx(tx, err))
	}
	defer rows.Close()

	// Iterate through the returned record(s)
	for rows.Next() {
		if err := rows.Scan(&m.ID, &m.CreateUser.Email, &m.CreateTime); err != nil {
			return errs.E(errs.Database, dt.datastorer.RollbackTx(tx, err))
		}
	}

	// If any error was encountered while iterating through rows.Next above
	// it will be returned here
	if err := rows.Err(); err != nil {
		return errs.E(errs.Database, dt.datastorer.RollbackTx(tx, err))
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
		return errs.E(errs.Database, dt.datastorer.RollbackTx(tx, errors.New("Invalid ID - no records updated")))
	}

	// Commit the Transaction
	if err := dt.datastorer.CommitTx(tx); err != nil {
		return errs.E(errs.Database, dt.datastorer.RollbackTx(tx, err))
	}

	return nil
}

// Delete removes the Movie record from the table
func (dt DefaultTransactor) Delete(ctx context.Context, m *movie.Movie) error {
	tx, err := dt.datastorer.BeginTx(ctx)
	if err != nil {
		return err
	}

	result, execErr := tx.ExecContext(ctx,
		`DELETE from demo.movie
		        WHERE movie_id = $1`, m.ID)

	if execErr != nil {
		return errs.E(errs.Database, dt.datastorer.RollbackTx(tx, execErr))
	}

	// Only 1 row should be deleted, check the result count to
	// ensure this is correct
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errs.E(errs.Database, dt.datastorer.RollbackTx(tx, err))
	}
	if rowsAffected == 0 {
		return errs.E(errs.Database, dt.datastorer.RollbackTx(tx, errors.New("No Rows Deleted")))
	} else if rowsAffected > 1 {
		return errs.E(errs.Database, dt.datastorer.RollbackTx(tx, errors.New("Too Many Rows Deleted")))
	}

	// Commit the Transaction
	if err := dt.datastorer.CommitTx(tx); err != nil {
		return errs.E(errs.Database, dt.datastorer.RollbackTx(tx, err))
	}

	return nil
}
