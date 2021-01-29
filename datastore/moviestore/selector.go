package moviestore

import (
	"context"
	"database/sql"

	"github.com/gilcrest/go-api-basic/datastore"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"

	"github.com/pkg/errors"
)

// Selector reads records from the db
type Selector interface {
	FindByID(context.Context, string) (*movie.Movie, error)
	FindAll(context.Context) ([]*movie.Movie, error)
}

// NewDefaultSelector is an initializer for DB
func NewDefaultSelector(ds datastore.Datastorer) (DefaultSelector, error) {
	if ds == nil {
		return DefaultSelector{}, errs.E(errs.Database, errors.New(errs.MissingField("ds").Error()))
	}
	return DefaultSelector{ds}, nil
}

// DefaultSelector is the database implementation for READ operations for a movie
type DefaultSelector struct {
	datastore.Datastorer
}

// FindByID returns a Movie struct to populate the response
func (d DefaultSelector) FindByID(ctx context.Context, extlID string) (*movie.Movie, error) {
	db := d.Datastorer.DB()

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
func (d DefaultSelector) FindAll(ctx context.Context) ([]*movie.Movie, error) {
	db := d.Datastorer.DB()

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
