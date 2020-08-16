package moviestore

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/gilcrest/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/rs/zerolog"
)

func NewMockTx() *MockTx {
	return &MockTx{}
}

// MockMovieDB is the mock database implementation for CRUD operations for a movie
type MockTx struct {
	Log zerolog.Logger
}

// Create is a mock for creating a record
func (t MockTx) Create(ctx context.Context, m *movie.Movie) error {
	const op errs.Op = "movieDatastore/MockTx.Create"

	// I would not recommend actually getting timestamps from the
	// database on create, but I put in an example of doing it anyway
	// Because of that, I have to set the timestamps in this mock
	// as if they were being set by the DB procedure that is being
	// called
	m.CreateTimestamp = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)
	m.UpdateTimestamp = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)

	return nil
}

// Update is a mock for updating a record
func (t MockTx) Update(ctx context.Context, m *movie.Movie) error {
	const op errs.Op = "movieDatastore/MockTx.Update"

	// Updates are a little different - on the non-mock, I am
	// actually getting back data as part of the update of the
	// original record that is helpful, so I am recreating that
	m.ID = uuid.MustParse("b7f34380-386d-4142-b9a0-3834d6e2288e")
	m.CreateTimestamp = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)

	return nil
}

// Delete mocks removing the Movie record from the table
func (t MockTx) Delete(ctx context.Context, m *movie.Movie) error {
	return nil
}

func NewMockDB() *MockDB {
	return &MockDB{}
}

type MockDB struct {
}

// FindByID returns a Movie struct to populate the response
func (d MockDB) FindByID(ctx context.Context, extlID string) (*movie.Movie, error) {
	m1 := new(movie.Movie)
	m1.ExternalID = extlID
	m1.Title = "The Thing"
	m1.Year = 1982
	m1.Rated = "R"
	m1.Released = time.Date(1982, time.June, 25, 0, 0, 0, 0, time.UTC)
	m1.RunTime = 109
	m1.Director = "John Carpenter"
	m1.Writer = "Bill Lancaster"
	m1.CreateTimestamp = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)
	m1.UpdateTimestamp = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)

	return m1, nil
}

// FindAll returns a slice of Movie structs to populate the response
func (d MockDB) FindAll(ctx context.Context) ([]*movie.Movie, error) {
	const op errs.Op = "movieDatastore/MockMovieDB.FindAll"

	m1 := new(movie.Movie)
	m1.ID = uuid.MustParse("4e58fa6a-5c4e-4e39-b6df-341087d1074b")
	m1.ExternalID = "Z8MnDR5iw70Z-Q9OIUgH"
	m1.Title = "The Thing"
	m1.Year = 1982
	m1.Rated = "R"
	m1.Released = time.Date(1982, time.June, 25, 0, 0, 0, 0, time.UTC)
	m1.RunTime = 109
	m1.Director = "John Carpenter"
	m1.Writer = "Bill Lancaster"
	m1.CreateTimestamp = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)
	m1.UpdateTimestamp = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)

	m2 := new(movie.Movie)
	m2.ID = uuid.MustParse("b7f34380-386d-4142-b9a0-3834d6e2288e")
	m2.ExternalID = "QxKhsURZ08sBP68MufYu"
	m2.Title = "Repo Man"
	m2.Year = 1984
	m2.Rated = "R"
	m2.Released = time.Date(1984, time.March, 2, 0, 0, 0, 0, time.UTC)
	m2.RunTime = 109
	m2.Director = "Alex Cox"
	m2.Writer = "Alex Cox"
	m2.CreateTimestamp = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)
	m2.UpdateTimestamp = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)

	s := []*movie.Movie{m1, m2}

	return s, nil
}
