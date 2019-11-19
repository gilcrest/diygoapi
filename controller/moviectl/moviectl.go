package moviectl

import (
	"context"
	"time"

	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/datastore/movieds"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/rs/xid"
)

// MovieController is used as the base controller for the Movie logic
type MovieController struct {
	App *app.Application
}

// Add adds a movie to the catalog.
func (ctl *MovieController) Add(ctx context.Context, r *AddMovieRequest) (*MovieResponse, error) {
	const op errs.Op = "controller/moviectl/MovieController.Add"

	err := ctl.App.DS.BeginTx(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	mds, err := movieds.ProvideMovieDS(ctl.App)
	if err != nil {
		return nil, errs.E(op, err)
	}

	m, err := provideMovie(r)
	if err != nil {
		return nil, errs.E(op, err)
	}

	err = m.Add(ctx)
	if err != nil {
		return nil, errs.E(err)
	}

	err = mds.Store(ctx, m)
	if err != nil {
		return nil, errs.E(op, errs.Database, ctl.App.DS.RollbackTx(err))
	}

	resp := provideMovieResponse(m)

	if err := ctl.App.DS.CommitTx(); err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	return resp, nil
}

// FindByID finds a movie given its' unique ID
func (ctl *MovieController) FindByID(ctx context.Context, id string) (*MovieResponse, error) {
	const op errs.Op = "controller/moviectl/FindByID"

	mds, err := movieds.ProvideMovieDS(ctl.App)
	if err != nil {
		return nil, errs.E(op, err)
	}

	i, err := xid.FromString(id)
	if err != nil {
		return nil, errs.E(op, errs.Validation, "Invalid id in URL path")
	}

	m, err := mds.FindByID(ctx, i)
	if err != nil {
		return nil, errs.E(op, errs.Database, ctl.App.DS.RollbackTx(err))
	}

	resp := provideMovieResponse(m)

	return resp, nil
}

// ProvideMovieController initializes MovieController
func ProvideMovieController(app *app.Application) *MovieController {
	return &MovieController{App: app}
}

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

// MovieResponse is the response struct
type MovieResponse struct {
	ExtlID          string `json:"ExtlID"`
	Title           string `json:"Title"`
	Year            int    `json:"Year"`
	Rated           string `json:"Rated"`
	Released        string `json:"ReleaseDate"`
	RunTime         int    `json:"RunTime"`
	Director        string `json:"Director"`
	Writer          string `json:"Writer"`
	CreateTimestamp string `json:"CreateTimestamp"`
}

// provideMovieResponse is an initializer for AddMovieResponse
func provideMovieResponse(m *movie.Movie) *MovieResponse {
	return &MovieResponse{
		ExtlID:          m.ExtlID.String(),
		Title:           m.Title,
		Year:            m.Year,
		Rated:           m.Rated,
		Released:        m.Released.Format(time.RFC3339),
		RunTime:         m.RunTime,
		Director:        m.Director,
		Writer:          m.Writer,
		CreateTimestamp: m.CreateTimestamp.Format(time.RFC3339),
	}
}

// dateFormat is the expected date format for any date fields
// in the request
const dateFormat string = "Jan 02 2006"

// ProvideMovie is an initializer for the Movie struct
func provideMovie(am *AddMovieRequest) (*movie.Movie, error) {
	const op errs.Op = "controller/moviectl/ProvideMovie"

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
