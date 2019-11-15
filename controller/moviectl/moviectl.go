package moviectl

import (
	"context"
	"time"

	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/datastore/movieds"
	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"
)

// MovieController is used as the base controller for the Movie logic
type MovieController struct {
	App *app.Application
}

// Add adds a movie to the catalog.
func (ctl *MovieController) Add(ctx context.Context, r *AddMovieRequest) (*AddMovieResponse, error) {
	const op errs.Op = "controller/moviectl/AddMovie"

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

	aud := new(audit.Audit)

	err = mds.Store(ctx, m, aud)
	if err != nil {
		return nil, errs.E(op, errs.Database, ctl.App.DS.RollbackTx(err))
	}

	resp := provideAddMovieResponse(m, aud)

	if err := ctl.App.DS.CommitTx(); err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

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

// provideAddMovieResponse is an initializer for AddMovieResponse
func provideAddMovieResponse(m *movie.Movie, a *audit.Audit) *AddMovieResponse {
	return &AddMovieResponse{
		ID:              m.ID.String(),
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

// Pull client information from Server token and set
// 	createClient, err := apiclient.ViaServerToken(ctx, tx)
// 	if err != nil {
// 		return errs.E(op, errs.Internal, err)
// 	}
// 	m.CreateClient.Number = createClient.Number
// 	m.UpdateClient.Number = createClient.Number
