// Package movie contains the business or "domain" logic for creating
// a Movie for this demo
package movie

import (
	"time"

	"github.com/google/uuid"

	"github.com/gilcrest/diy-go-api/domain/errs"
	"github.com/gilcrest/diy-go-api/domain/secure"
)

// Movie holds details of a movie
type Movie struct {
	ID         uuid.UUID
	ExternalID secure.Identifier
	Title      string
	Rated      string
	Released   time.Time
	RunTime    int
	Director   string
	Writer     string
}

// IsValid performs validation of the struct
func (m *Movie) IsValid() error {
	switch {
	case m.ExternalID.String() == "":
		return errs.E(errs.Validation, errs.Parameter("extlID"), errs.MissingField("extlID"))
	case m.Title == "":
		return errs.E(errs.Validation, errs.Parameter("title"), errs.MissingField("title"))
	case m.Rated == "":
		return errs.E(errs.Validation, errs.Parameter("rated"), errs.MissingField("rated"))
	case m.Released.IsZero():
		return errs.E(errs.Validation, errs.Parameter("release_date"), "release_date must have a value")
	case m.RunTime <= 0:
		return errs.E(errs.Validation, errs.Parameter("run_time"), "run_time must be greater than zero")
	case m.Director == "":
		return errs.E(errs.Validation, errs.Parameter("director"), errs.MissingField("director"))
	case m.Writer == "":
		return errs.E(errs.Validation, errs.Parameter("writer"), errs.MissingField("writer"))
	}

	return nil
}
