package movieController

import (
	"context"
	"net/http"
	"time"

	"github.com/gilcrest/errs"
	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/controller"
	"github.com/gilcrest/go-api-basic/datastore/movieDatastore"
	"github.com/gilcrest/go-api-basic/domain/movie"
)

// MovieController is used as the base controller for the Movie logic
type MovieController struct {
	App *app.Application
	SRF controller.StandardResponseFields
}

// MovieRequest is the request struct
type MovieRequest struct {
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Rated    string `json:"rated"`
	Released string `json:"release_date"`
	RunTime  int    `json:"run_time"`
	Director string `json:"director"`
	Writer   string `json:"writer"`
}

// MovieResponse is the response struct for a single Movie
type MovieResponse struct {
	ExtlID          string `json:"extl_id"`
	Title           string `json:"title"`
	Year            int    `json:"year"`
	Rated           string `json:"rated"`
	Released        string `json:"release_date"`
	RunTime         int    `json:"run_time"`
	Director        string `json:"director"`
	Writer          string `json:"writer"`
	CreateTimestamp string `json:"create_timestamp"`
	UpdateTimestamp string `json:"update_timestamp"`
}

// ListMovieResponse is the response struct for multiple Movies
type ListMovieResponse struct {
	controller.StandardResponseFields
	Data []*MovieResponse `json:"data"`
}

// SingleMovieResponse is the response struct for multiple Movies
type SingleMovieResponse struct {
	controller.StandardResponseFields
	Data *MovieResponse `json:"data"`
}

// DeleteMovieResponse is the response struct for deleted Movies
type DeleteMovieResponse struct {
	controller.StandardResponseFields
	Data struct {
		ExtlID  string `json:"extl_id"`
		Deleted bool   `json:"deleted"`
	} `json:"data"`
}

func newDeleteMovieResponse(m *movie.Movie, srf controller.StandardResponseFields) *DeleteMovieResponse {
	return &DeleteMovieResponse{
		StandardResponseFields: srf,
		Data: struct {
			ExtlID  string "json:\"extl_id\""
			Deleted bool   "json:\"deleted\""
		}{
			ExtlID:  m.ExtlID,
			Deleted: true,
		},
	}
}

// newMovieResponse is an initializer for MovieResponse
func (ctl *MovieController) newMovieResponse(m *movie.Movie) *MovieResponse {
	return &MovieResponse{
		ExtlID:          m.ExtlID,
		Title:           m.Title,
		Year:            m.Year,
		Rated:           m.Rated,
		Released:        m.Released.Format(time.RFC3339),
		RunTime:         m.RunTime,
		Director:        m.Director,
		Writer:          m.Writer,
		CreateTimestamp: m.CreateTimestamp.Format(time.RFC3339),
		UpdateTimestamp: m.UpdateTimestamp.Format(time.RFC3339),
	}
}

// NewMovieController initializes MovieController
func NewMovieController(app *app.Application, srf controller.StandardResponseFields) *MovieController {
	return &MovieController{App: app, SRF: srf}
}

// Add adds a movie to the catalog.
func (ctl *MovieController) Add(ctx context.Context, r *MovieRequest) (*MovieResponse, error) {
	const op errs.Op = "controller/movieController/MovieController.Add"

	err := ctl.App.DS.BeginTx(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	mds, err := movieDatastore.NewMovieDS(ctl.App)
	if err != nil {
		return nil, errs.E(op, err)
	}

	m, err := newMovie(r)
	if err != nil {
		return nil, errs.E(op, err)
	}

	err = m.Add(ctx)
	if err != nil {
		return nil, errs.E(err)
	}

	err = mds.Create(ctx, m)
	if err != nil {
		return nil, errs.E(op, errs.Database, ctl.App.DS.RollbackTx(err))
	}

	resp := ctl.newMovieResponse(m)

	if err := ctl.App.DS.CommitTx(); err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	return resp, nil
}

// Update updates the movie given the id sent in
func (ctl *MovieController) Update(ctx context.Context, id string, r *MovieRequest) (*MovieResponse, error) {
	const op errs.Op = "controller/movieController/MovieController.Update"

	err := ctl.App.DS.BeginTx(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	mds, err := movieDatastore.NewMovieDS(ctl.App)
	if err != nil {
		return nil, errs.E(op, err)
	}

	m, err := newMovie(r)
	if err != nil {
		return nil, errs.E(op, err)
	}

	err = m.Update(ctx, id)
	if err != nil {
		return nil, errs.E(err)
	}

	err = mds.Update(ctx, m)
	if err != nil {
		return nil, errs.E(op, errs.Database, ctl.App.DS.RollbackTx(err))
	}

	resp := ctl.newMovieResponse(m)

	if err := ctl.App.DS.CommitTx(); err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	return resp, nil
}

// FindByID finds a movie given its' unique ID
func (ctl *MovieController) FindByID(ctx context.Context, id string) (*SingleMovieResponse, error) {
	const op errs.Op = "controller/movieController/MovieController.FindByID"

	mds, err := movieDatastore.NewMovieDS(ctl.App)
	if err != nil {
		return nil, errs.E(op, err)
	}

	m, err := mds.FindByID(ctx, id)
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	mr := ctl.newMovieResponse(m)

	response := ctl.NewSingleMovieResponse(mr)

	return response, nil
}

// FindAll finds the entire set of Movies
func (ctl *MovieController) FindAll(ctx context.Context, r *http.Request) (*ListMovieResponse, error) {
	const op errs.Op = "controller/movieController/MovieController.FindAll"

	mds, err := movieDatastore.NewMovieDS(ctl.App)
	if err != nil {
		return nil, errs.E(op, err)
	}

	ms, err := mds.FindAll(ctx)
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	response := ctl.NewListMovieResponse(ms, r)

	return response, nil
}

// NewListMovieResponse is an initializer for ListMovieResponse
func (ctl *MovieController) NewListMovieResponse(ms []*movie.Movie, r *http.Request) *ListMovieResponse {
	const op errs.Op = "controller/movieController/MovieController.NewListMovieResponse"

	var s []*MovieResponse

	for _, m := range ms {
		mr := ctl.newMovieResponse(m)
		s = append(s, mr)
	}

	return &ListMovieResponse{StandardResponseFields: ctl.SRF, Data: s}
}

// NewSingleMovieResponse is an initializer for SingleMovieResponse
func (ctl *MovieController) NewSingleMovieResponse(mr *MovieResponse) *SingleMovieResponse {
	const op errs.Op = "controller/movieController/MovieController.NewSingleMovieResponse"

	return &SingleMovieResponse{StandardResponseFields: ctl.SRF, Data: mr}
}

// NewMovie is an initializer for the Movie struct
func newMovie(am *MovieRequest) (*movie.Movie, error) {
	const op errs.Op = "controller/movieController/newMovie"

	// Parse Release Date according to RFC3339
	t, err := time.Parse(time.RFC3339, am.Released)
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

// Delete removes the movie given the id sent in
func (ctl *MovieController) Delete(ctx context.Context, id string) (*DeleteMovieResponse, error) {
	const op errs.Op = "controller/movieController/MovieController.Update"

	err := ctl.App.DS.BeginTx(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	mds, err := movieDatastore.NewMovieDS(ctl.App)
	if err != nil {
		return nil, errs.E(op, err)
	}

	m, err := mds.FindByID(ctx, id)
	if err != nil {
		return nil, errs.E(op, errs.Database, ctl.App.DS.RollbackTx(err))
	}

	err = mds.Delete(ctx, m)
	if err != nil {
		return nil, errs.E(op, errs.Database, ctl.App.DS.RollbackTx(err))
	}

	resp := newDeleteMovieResponse(m, ctl.SRF)

	if err := ctl.App.DS.CommitTx(); err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	return resp, nil
}
