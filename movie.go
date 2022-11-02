package diy

import (
	"context"
)

// MovieServicer is used to create, read, update and delete movies.
type MovieServicer interface {
	Create(ctx context.Context, r *CreateMovieRequest, adt Audit) (*MovieResponse, error)
	Update(ctx context.Context, r *UpdateMovieRequest, adt Audit) (*MovieResponse, error)
	Delete(ctx context.Context, extlID string) (DeleteResponse, error)
	FindMovieByID(ctx context.Context, extlID string) (*MovieResponse, error)
	FindAllMovies(ctx context.Context) ([]*MovieResponse, error)
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
