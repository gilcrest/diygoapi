package movie

import (
	"context"
	"time"

	"github.com/gilcrest/errs"
	"github.com/gilcrest/go-api-basic/domain/random"
	"github.com/google/uuid"
)

// Adder interface is for actions that Add a new movie to the
type Adder interface {
	Add(ctx context.Context) error
}

type Updater interface {
	Update(ctx context.Context, id string) error
}

// Movie holds details of a movie
type Movie struct {
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

// Validate does basic input validation and ensures the struct is
// properly constructed
func (m *Movie) validate() error {
	const op errs.Op = "domain/Movie.validate"

	switch {
	case m.Title == "":
		return errs.E(op, errs.Validation, errs.Parameter("Title"), errs.MissingField("Title"))
	case m.Year < 1878:
		return errs.E(op, errs.Validation, errs.Parameter("Year"), "The first film was in 1878, Year must be >= 1878")
	case m.Rated == "":
		return errs.E(op, errs.Validation, errs.Parameter("Rated"), errs.MissingField("Rated"))
	case m.Released.IsZero() == true:
		return errs.E(op, errs.Validation, errs.Parameter("ReleaseDate"), "Released must have a value")
	case m.RunTime <= 0:
		return errs.E(op, errs.Validation, errs.Parameter("RunTime"), "Run time must be greater than zero")
	case m.Director == "":
		return errs.E(op, errs.Validation, errs.Parameter("Director"), errs.MissingField("Director"))
	case m.Writer == "":
		return errs.E(op, errs.Validation, errs.Parameter("Writer"), errs.MissingField("Writer"))
	}

	return nil
}

// Add performs business validations prior to writing to the db
func (m *Movie) Add(ctx context.Context) error {
	const op errs.Op = "movie/Movie.Add"

	err := m.validate()
	if err != nil {
		return errs.E(op, err)
	}

	m.ID = uuid.New()
	id, err := random.CryptoString(15)
	if err != nil {
		return errs.E(op, errs.Validation, errs.Internal, err)
	}
	m.ExternalID = id

	return nil
}

// Update performs business validations prior to writing to the db
func (m *Movie) Update(ctx context.Context, id string) error {
	const op errs.Op = "movie/Movie.Update"

	err := m.validate()
	if err != nil {
		return errs.E(op, err)
	}

	m.UpdateTimestamp = time.Now().UTC()

	return nil
}
