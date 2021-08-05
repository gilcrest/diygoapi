// Package moviestore performs all DML and select operations for a movie
package moviestore

import (
	"context"
	"database/sql"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"
)

// Datastorer is an interface for working with the Database
type Datastorer interface {
	// DB returns a sql.DB
	DB() *sql.DB
	// BeginTx starts a sql.Tx using the input context
	BeginTx(context.Context) (*sql.Tx, error)
	// RollbackTx rolls back the input sql.Tx
	RollbackTx(*sql.Tx, error) error
	// CommitTx commits the Tx
	CommitTx(*sql.Tx) error
}

// Selector is the database implementation for READ operations for a movie
type Selector struct {
	Datastorer
}

// NewSelector is an initializer for Selector
func NewSelector(ds Datastorer) Selector {
	return Selector{ds}
}

// FindByID returns a Movie struct to populate the response
func (s Selector) FindByID(ctx context.Context, extlID string) (*movie.Movie, error) {
	db := s.Datastorer.DB()

	// Prepare the sql statement using bind variables
	row := db.QueryRowContext(ctx,
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
func (s Selector) FindAll(ctx context.Context) ([]*movie.Movie, error) {
	db := s.Datastorer.DB()

	// use QueryContext to get back sql.Rows
	rows, err := db.QueryContext(ctx,
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
	movies := make([]*movie.Movie, 0)

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

		movies = append(movies, m)
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
	if len(movies) == 0 {
		return nil, errs.E(errs.Validation, "No rows returned")
	}

	// return the slice
	return movies, nil
}
