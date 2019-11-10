package moviectl

import (
	"context"
	"database/sql"
	"time"

	"github.com/gilcrest/errors"
	"github.com/gilcrest/go-api-basic/datastore/movieds"
	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/rs/zerolog"
)

// AddMovieRequest is the request struct
type AddMovieRequest struct {
	Title    string `json:"Title"`
	Year     int    `json:"Year"`
	Rated    string `json:"Rated"`
	Released string `json:"ReleaseDate"`
	RunTime  int    `json:"RunTime"`
	Director string `json:"Director"`
	Writer   string `json:"Writer"`
}

// AddMovieResponse is the response struct
type AddMovieResponse struct {
	ID              string `json:"ID"`
	Title           string `json:"Title"`
	Year            int    `json:"Year"`
	Rated           string `json:"Rated"`
	Released        string `json:"ReleaseDate"`
	RunTime         int    `json:"RunTime"`
	Director        string `json:"Director"`
	Writer          string `json:"Writer"`
	CreateTimestamp string `json:"CreateTimestamp"`
}

type addMovieController struct {
	Request *AddMovieRequest
	MovieDS movieds.MovieDS
}

func (amc addMovieController) Add(ctx context.Context) (*AddMovieResponse, error) {
	const op errs.Op = "domain/movie/AddMovie"

	m, err := provideMovie(amc.Request)
	if err != nil {
		return nil, errs.E(op, err)
	}

	err = m.Add(ctx)
	if err != nil {
		return nil, errs.E(err)
	}

	aud := new(audit.Audit)

	err = amc.MovieDS.Store(ctx, m, aud)
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	return provideAddMovieResponse(m, aud), nil
}

// provideAddMovieResponse is an initializer for AddMovieResponse
func provideAddMovieResponse(m *movie.Movie, a *audit.Audit) *AddMovieResponse {
	return &AddMovieResponse{
		ID:              m.ID,
		Title:           m.Title,
		Year:            m.Year,
		Rated:           m.Rated,
		Released:        m.Released.Format(time.RFC3339),
		RunTime:         m.RunTime,
		Director:        m.Director,
		Writer:          m.Writer,
		CreateTimestamp: a.CreateTimestamp.Format(time.RFC3339),
	}
}

// dateFormat is the expected date format for any date fields
// in the request
const dateFormat string = "Jan 02 2006"

// ProvideMovie is an initializer for the Movie struct
func provideMovie(am *AddMovieRequest) (*movie.Movie, error) {
	const op errs.Op = "domain/movie/ProvideMovie"

	t, err := time.Parse(dateFormat, am.Released)
	if err != nil {
		return nil, errs.E(op,
			errs.Validation,
			errs.Code("invalid_date_format"),
			errs.Parameter("ReleaseDate"),
			err)
	}

	return &movie.Movie{
		Title:    am.Title,
		Year:     am.Year,
		Rated:    am.Rated,
		Released: t,
		RunTime:  am.RunTime,
		Director: am.Director,
		Writer:   am.Writer,
	}, nil
}

// AddMovie adds a movie to the catalog
func AddMovie(ctx context.Context, db *sql.DB, log zerolog.Logger, r *AddMovieRequest) (*AddMovieResponse, error) {
	const op errs.Op = "controller/moviectl/AddMovie"

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	mdb := movieds.ProvideMovieDS(tx, log)

	amc := addMovieController{Request: r, MovieDS: mdb}
	resp, err := amc.Add(ctx)
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return nil, errs.E(op, errors.Database, err)
		}
		// Kind could be Database or Exist from db, so
		// use type assertion and send both up
		if e, ok := err.(*errors.Error); ok {
			return nil, errs.E(e.Kind, e.Code, e.Param, err)
		}
		// Should not actually fall to here, but including as
		// good practice
		return nil, errs.E(op, errors.Database, err)
	}

	if err := tx.Commit(); err != nil {
		return nil, errs.E(op, errors.Database, err)
	}

	return resp, nil
}

// 	// Pull client information from Server token and set
// 	createClient, err := apiclient.ViaServerToken(ctx, tx)
// 	if err != nil {
// 		return errors.E(op, errors.Internal, err)
// 	}
// 	m.CreateClient.Number = createClient.Number
// 	m.UpdateClient.Number = createClient.Number
