package moviectl

import (
	"context"
	"time"

	"github.com/gilcrest/errors"
	"github.com/gilcrest/go-api-basic/datastore"
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

func (amc addMovieController) add(ctx context.Context) (*AddMovieResponse, error) {
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

func provideAddMovieController(r *AddMovieRequest, ds movieds.MovieDS) *addMovieController {
	return &addMovieController{Request: r, MovieDS: ds}
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

// AddMovie adds a movie to the catalog.
func AddMovie(ctx context.Context, ds datastore.Datastore, log zerolog.Logger, r *AddMovieRequest) (*AddMovieResponse, error) {
	const op errs.Op = "controller/moviectl/AddMovie"

	err := ds.BeginTx(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	mds, err := movieds.ProvideMovieDS(ds, log)
	if err != nil {
		return nil, errs.E(op, err)
	}

	amc := provideAddMovieController(r, mds)
	resp, err := amc.add(ctx)
	if err != nil {
		return nil, ds.RollbackTx(err)
	}

	if err := ds.CommitTx(); err != nil {
		return nil, errs.E(op, errors.Database, err)
	}

	return resp, nil
}

// Pull client information from Server token and set
// 	createClient, err := apiclient.ViaServerToken(ctx, tx)
// 	if err != nil {
// 		return errors.E(op, errors.Internal, err)
// 	}
// 	m.CreateClient.Number = createClient.Number
// 	m.UpdateClient.Number = createClient.Number
