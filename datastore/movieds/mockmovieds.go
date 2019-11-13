package movieds

import (
	"context"
	"time"

	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/rs/zerolog"
)

// MockMovieDB is the mock database implementation for CRUD operations for a movie
type MockMovieDB struct {
	Log zerolog.Logger
}

// Store creates a record in the user table using a stored function
func (mdb MockMovieDB) Store(ctx context.Context, m *movie.Movie, a *audit.Audit) error {
	const op errs.Op = "movie/Movie.createDB"

	a.CreateTimestamp = time.Now()
	a.UpdateTimestamp = time.Now()

	return nil
}
