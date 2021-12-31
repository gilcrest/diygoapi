// Package moviestore performs all DML and select operations for a movie
package moviestore

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"
)

// Datastorer is an interface for working with the Database
type Datastorer interface {
	// Pool returns *pgxpool.Pool
	Pool() *pgxpool.Pool
	// BeginTx starts a pgx.Tx using the input context
	BeginTx(ctx context.Context) (pgx.Tx, error)
	// RollbackTx rolls back the input pgx.Tx
	RollbackTx(ctx context.Context, tx pgx.Tx, err error) error
	// CommitTx commits the Tx
	CommitTx(ctx context.Context, tx pgx.Tx) error
}

// Selector is the database implementation for READ operations for a movie
type Selector struct {
	datastorer Datastorer
}

// NewSelector is an initializer for Selector
func NewSelector(ds Datastorer) Selector {
	return Selector{datastorer: ds}
}

// FindByID returns a Movie struct to populate the response
func (s Selector) FindByID(ctx context.Context, extlID string) (*movie.Movie, error) {
	dbpool := s.datastorer.Pool()

	// Prepare the sql statement using bind variables
	row := dbpool.QueryRow(ctx,
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

	if err == pgx.ErrNoRows {
		return nil, errs.E(errs.NotExist, "No record found for given ID")
	} else if err != nil {
		return nil, errs.E(errs.Database, err)
	}

	return m, nil
}

// FindAll returns a slice of Movie structs to populate the response
func (s Selector) FindAll(ctx context.Context) ([]*movie.Movie, error) {
	dbpool := s.datastorer.Pool()

	// use QueryContext to get back sql.Rows
	rows, err := dbpool.Query(ctx,
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

	// Rows.Err will report the last error encountered by Rows.Scan.
	if rows.Err() != nil {
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
