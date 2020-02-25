package movie

import (
	"context"
	"time"

	"github.com/gilcrest/errs"
	"github.com/google/uuid"
)

type MockMovie struct {
	ID              uuid.UUID
	ExternalID      string
	Title           string
	Year            int
	Rated           string
	Released        time.Time
	RunTime         int
	Director        string
	Writer          string
	CreateTimestamp time.Time
	UpdateTimestamp time.Time
}

// Add performs business validations prior to writing to the db
func (m *MockMovie) Add(ctx context.Context) error {
	const op errs.Op = "movie/MockMovie.Add"

	m.ID = uuid.MustParse("b7f34380-386d-4142-b9a0-3834d6e2288e")
	m.ExternalID = "mlPb1YimScrEsmJJa3Xd"

	return nil
}

// Add performs business validations prior to writing to the db
func (m *MockMovie) Update(ctx context.Context, id string) error {
	const op errs.Op = "movie/MockMovie.Update"

	m.UpdateTimestamp = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)

	return nil
}
