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

	m.ExternalID = "8675309"

	return nil
}

// Add performs business validations prior to writing to the db
func (m *MockMovie) Update(ctx context.Context, id string) error {
	const op errs.Op = "movie/MockMovie.Update"

	//m.ID = "1234567"
	m.ExternalID = "8675309"

	return nil
}
