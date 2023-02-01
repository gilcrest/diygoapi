package diygoapi

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/secure"
)

// MovieServicer is used to create, read, update and delete movies.
type MovieServicer interface {
	Create(ctx context.Context, r *CreateMovieRequest, adt Audit) (*MovieResponse, error)
	Update(ctx context.Context, r *UpdateMovieRequest, adt Audit) (*MovieResponse, error)
	Delete(ctx context.Context, extlID string) (DeleteResponse, error)
	FindMovieByExternalID(ctx context.Context, extlID string) (*MovieResponse, error)
	FindAllMovies(ctx context.Context) ([]*MovieResponse, error)
}

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
	const op errs.Op = "diygoapi/Movie.IsValid"

	switch {
	case m.ExternalID.String() == "":
		return errs.E(op, errs.Validation, errs.Parameter("extlID"), errs.MissingField("extlID"))
	case m.Title == "":
		return errs.E(op, errs.Validation, errs.Parameter("title"), errs.MissingField("title"))
	case m.Rated == "":
		return errs.E(op, errs.Validation, errs.Parameter("rated"), errs.MissingField("rated"))
	case m.Released.IsZero():
		return errs.E(op, errs.Validation, errs.Parameter("release_date"), "release_date must have a value")
	case m.RunTime <= 0:
		return errs.E(op, errs.Validation, errs.Parameter("run_time"), "run_time must be greater than zero")
	case m.Director == "":
		return errs.E(op, errs.Validation, errs.Parameter("director"), errs.MissingField("director"))
	case m.Writer == "":
		return errs.E(op, errs.Validation, errs.Parameter("writer"), errs.MissingField("writer"))
	}

	return nil
}

// CreateMovieRequest is the request struct for Creating a Movie
type CreateMovieRequest struct {
	Title    string `json:"title"`
	Rated    string `json:"rated"`
	Released string `json:"release_date"`
	RunTime  int    `json:"run_time"`
	Director string `json:"director"`
	Writer   string `json:"writer"`
}

// UpdateMovieRequest is the request struct for updating a Movie
type UpdateMovieRequest struct {
	ExternalID string
	Title      string `json:"title"`
	Rated      string `json:"rated"`
	Released   string `json:"release_date"`
	RunTime    int    `json:"run_time"`
	Director   string `json:"director"`
	Writer     string `json:"writer"`
}

// MovieResponse is the response struct for a Movie
type MovieResponse struct {
	ExternalID          string `json:"external_id"`
	Title               string `json:"title"`
	Rated               string `json:"rated"`
	Released            string `json:"release_date"`
	RunTime             int    `json:"run_time"`
	Director            string `json:"director"`
	Writer              string `json:"writer"`
	CreateAppExtlID     string `json:"create_app_extl_id"`
	CreateUserFirstName string `json:"create_user_first_name"`
	CreateUserLastName  string `json:"create_user_last_name"`
	CreateDateTime      string `json:"create_date_time"`
	UpdateAppExtlID     string `json:"update_app_extl_id"`
	UpdateUserFirstName string `json:"update_user_first_name"`
	UpdateUserLastName  string `json:"update_user_last_name"`
	UpdateDateTime      string `json:"update_date_time"`
}
