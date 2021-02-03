// Package movie contains the business or "domain" logic for creating
// a Movie for this demo
package movie

import (
	"context"
	"time"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/user"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Writer is used to create/update a Movie
type Writer interface {
	Add(ctx context.Context) error
	Update(ctx context.Context, id string) error
}

// NewMovie initializes a Movie struct for use in Movie creation
func NewMovie(id uuid.UUID, extlID string, u *user.User) (*Movie, error) {
	switch {
	case id == uuid.Nil:
		return nil, errs.E(errs.Validation, errs.Parameter("ID"), errors.New(errs.MissingField("ID").Error()))
	case extlID == "":
		return nil, errs.E(errs.Validation, errs.Parameter("ID"), errors.New(errs.MissingField("ID").Error()))
	case !u.IsValid():
		return nil, errs.E(errs.Validation, errs.Parameter("User"), errors.New("User is invalid"))
	}

	now := time.Now().UTC()

	return &Movie{
		ID:         id,
		ExternalID: extlID,
		CreateUser: *u,
		CreateTime: now,
		UpdateUser: *u,
		UpdateTime: now,
	}, nil
}

// Movie holds details of a movie
type Movie struct {
	ID         uuid.UUID
	ExternalID string
	Title      string
	Rated      string
	Released   time.Time
	RunTime    int
	Director   string
	Writer     string
	CreateUser user.User
	CreateTime time.Time
	UpdateUser user.User
	UpdateTime time.Time
}

func (m *Movie) SetExternalID(id string) *Movie {
	m.ExternalID = id
	return m
}

func (m *Movie) SetTitle(t string) *Movie {
	m.Title = t
	return m
}

func (m *Movie) SetRated(r string) *Movie {
	m.Rated = r
	return m
}

func (m *Movie) SetReleased(r string) (*Movie, error) {
	t, err := time.Parse(time.RFC3339, r)
	if err != nil {
		return nil, errs.E(errs.Validation,
			errs.Code("invalid_date_format"),
			errs.Parameter("release_date"),
			errors.WithStack(err))
	}
	m.Released = t
	return m, nil
}

func (m *Movie) SetRunTime(rt int) *Movie {
	m.RunTime = rt
	return m
}

func (m *Movie) SetDirector(d string) *Movie {
	m.Director = d
	return m
}

func (m *Movie) SetWriter(w string) *Movie {
	m.Writer = w
	return m
}

func (m *Movie) SetUpdateUser(u *user.User) *Movie {
	m.UpdateUser = *u
	return m
}

func (m *Movie) SetUpdateTime() *Movie {
	m.UpdateTime = time.Now().UTC()
	return m
}

// IsValid performs validation of the struct
func (m *Movie) IsValid() error {
	switch {
	case m.Title == "":
		return errs.E(errs.Validation, errs.Parameter("title"), errs.MissingField("title"))
	case m.Rated == "":
		return errs.E(errs.Validation, errs.Parameter("rated"), errs.MissingField("Rated"))
	case m.Released.IsZero():
		return errs.E(errs.Validation, errs.Parameter("release_date"), "Released must have a value")
	case m.RunTime <= 0:
		return errs.E(errs.Validation, errs.Parameter("run_time"), "Run time must be greater than zero")
	case m.Director == "":
		return errs.E(errs.Validation, errs.Parameter("director"), errs.MissingField("Director"))
	case m.Writer == "":
		return errs.E(errs.Validation, errs.Parameter("writer"), errs.MissingField("Writer"))
	}

	return nil
}
