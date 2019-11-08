package movie

import (
	"context"
	"time"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/rs/xid"
)

// Movie holds details of a movie
type Movie struct {
	ID       string
	Title    string
	Year     int
	Rated    string
	Released time.Time
	RunTime  int
	Director string
	Writer   string
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
	const op errs.Op = "movie/Movie.Create"

	m.ID = xid.New().String()

	// Validate input data
	err := m.validate()
	if err != nil {
		if e, ok := err.(*errs.Error); ok {
			return errs.E(errs.Validation, e.Param, err)
		}
		// should not get here, but just in case
		return errs.E(errs.Validation, err)
	}

	return nil
}
