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

	m1 := new(movie.Movie)
	m1.ExtlID = xid.New()
	m1.Title = "The Thing"
	m1.Year = 1982
	m1.Rated = "R"
	m1.Released = time.Date(1982, time.June, 25, 0, 0, 0, 0, time.UTC)
	m1.RunTime = 109
	m1.Director = "John Carpenter"
	m1.Writer = "Bill Lancaster"
	m1.CreateTimestamp = time.Now()

	return m1, nil
}

// FindAll returns a slice of Movie structs to populate the response
func (mdb MockMovieDB) FindAll(ctx context.Context) ([]*movie.Movie, error) {
	const op errs.Op = "movieds/MockMovieDB.FindAll"

	m1 := new(movie.Movie)
	m1.ExtlID = xid.New()
	m1.Title = "The Thing"
	m1.Year = 1982
	m1.Rated = "R"
	m1.Released = time.Date(1982, time.June, 25, 0, 0, 0, 0, time.UTC)
	m1.RunTime = 109
	m1.Director = "John Carpenter"
	m1.Writer = "Bill Lancaster"
	m1.CreateTimestamp = time.Now()

	m2 := new(movie.Movie)
	m2.ExtlID = xid.New()
	m2.Title = "Repo Man"
	m2.Year = 1984
	m2.Rated = "R"
	m2.Released = time.Date(1984, time.March, 2, 0, 0, 0, 0, time.UTC)
	m2.RunTime = 109
	m2.Director = "Alex Cox"
	m2.Writer = "Alex Cox"
	m2.CreateTimestamp = time.Now()

	s := []*movie.Movie{m1, m2}

	return s, nil
}
