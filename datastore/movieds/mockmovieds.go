package movieds

import (
	"context"
	"time"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
)

// MockMovieDB is the mock database implementation for CRUD operations for a movie
type MockMovieDB struct {
	Log zerolog.Logger
}

// Store creates a record in the user table using a stored function
func (mdb MockMovieDB) Store(ctx context.Context, m *movie.Movie) error {
	const op errs.Op = "movieds/MockMovieDB.Store"

	return nil
}

// FindByID returns a Movie struct to populate the response
func (mdb MockMovieDB) FindByID(ctx context.Context, extlID xid.ID) (*movie.Movie, error) {
	const op errs.Op = "movieds/MockMovieDB.FindByID"

	m := new(movie.Movie)
	m.ExtlID = extlID
	m.Title = "Clockwork Orange"
	m.CreateTimestamp = time.Now()

	return m, nil
}
