package movieController

import (
	"context"
	"fmt"
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

// RequestBody is the request struct
type RequestData struct {
	Title    string `json:"title"`
	Year     int    `json:"year"`
	Rated    string `json:"rated"`
	Released string `json:"release_date"`
	RunTime  int    `json:"run_time"`
	Director string `json:"director"`
	Writer   string `json:"writer"`
}

// ResponseData is the response struct for a single Movie
type ResponseData struct {
	ExternalID      string `json:"external_id"`
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
	Data []*ResponseData `json:"data"`
}

// SingleMovieResponse is the response struct for multiple Movies
type SingleMovieResponse struct {
	controller.StandardResponseFields
	Data *ResponseData `json:"data"`
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
			ExtlID:  m.ExternalID,
			Deleted: true,
		},
	}
}

// newMovieResponse is an initializer for MovieResponse
func newMovieResponse(i interface{}) (*ResponseData, error) {
	const op errs.Op = "controller/movieController/newMovieResponse"

	switch m := i.(type) {
	case *movie.Movie:
		return &ResponseData{
			ExternalID:      m.ExternalID,
			Title:           m.Title,
			Year:            m.Year,
			Rated:           m.Rated,
			Released:        m.Released.Format(time.RFC3339),
			RunTime:         m.RunTime,
			Director:        m.Director,
			Writer:          m.Writer,
			CreateTimestamp: m.CreateTimestamp.Format(time.RFC3339),
			UpdateTimestamp: m.UpdateTimestamp.Format(time.RFC3339),
		}, nil
	case *movie.MockMovie:
		return &ResponseData{
			ExternalID:      m.ExternalID,
			Title:           m.Title,
			Year:            m.Year,
			Rated:           m.Rated,
			Released:        m.Released.Format(time.RFC3339),
			RunTime:         m.RunTime,
			Director:        m.Director,
			Writer:          m.Writer,
			CreateTimestamp: m.CreateTimestamp.Format(time.RFC3339),
			UpdateTimestamp: m.UpdateTimestamp.Format(time.RFC3339),
		}, nil
	default:
		return nil, errs.E(op, fmt.Errorf("unknown type %T", m))
	}
}

// NewMovieController initializes MovieController
func NewMovieController(app *app.Application, srf controller.StandardResponseFields) *MovieController {
	return &MovieController{App: app, SRF: srf}
}

// Add adds a movie to the catalog.
func (ctl *MovieController) Add(ctx context.Context, r *RequestData) (*SingleMovieResponse, error) {
	const op errs.Op = "controller/movieController/MovieController.Add"

	// Convert request into a Movie struct
	m, err := ctl.newAdder(r)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// Call the Add method to perform domain business logic
	err = m.Add(ctx)
	if err != nil {
		return nil, errs.E(err)
	}

	// Begin a DB Tx, if the underlying struct is a MockDatastore then
	// the Tx will be nil
	tx, err := ctl.App.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// NewTransactor gives back either a MockTx or a concrete
	// Tx, depending upon whether the Tx from Datastorer is nil or not
	t, err := movieDatastore.NewTransactor(tx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// Call the Create method of the Transactor to insert data to
	// the database (unless mocked, of course). If an error occurs,
	// rollback the transaction
	err = t.Create(ctx, m)
	if err != nil {
		return nil, errs.E(op, ctl.App.Datastorer.RollbackTx(tx, err))
	}

	// Commit the Transaction
	if err := ctl.App.Datastorer.CommitTx(tx); err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	rd, err := newMovieResponse(m)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// Populate the response
	response := ctl.NewSingleMovieResponse(rd)

	return response, nil
}

// Update updates the movie given the id sent in
func (ctl *MovieController) Update(ctx context.Context, id string, r *RequestData) (*SingleMovieResponse, error) {
	const op errs.Op = "controller/movieController/MovieController.Update"

	// Convert request into a Movie struct
	m, err := ctl.newUpdater(r)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// Perform domain Update "business logic"
	err = m.Update(ctx, id)
	if err != nil {
		return nil, errs.E(err)
	}

	// Begin a DB Tx, if the underlying struct is a MockDatastore then
	// the Tx will be nil
	tx, err := ctl.App.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// NewTransactor gives back either a MockTx or a concrete
	// Tx, depending upon whether the Tx from Datastorer is nil or not
	t, err := movieDatastore.NewTransactor(tx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// Call the Update method of the Transactor to update data on
	// the database (unless mocked, of course). If an error occurs,
	// rollback the transaction
	err = t.Update(ctx, m)
	if err != nil {
		return nil, errs.E(op, ctl.App.Datastorer.RollbackTx(tx, err))
	}

	// Commit the Transaction
	if err := ctl.App.Datastorer.CommitTx(tx); err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	rd, err := newMovieResponse(m)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// Populate the response
	response := ctl.NewSingleMovieResponse(rd)

	return response, nil
}

// Delete removes the movie given the id sent in
func (ctl *MovieController) Delete(ctx context.Context, id string) (*DeleteMovieResponse, error) {
	const op errs.Op = "controller/movieController/MovieController.Update"

	// NewSelector returns either a concrete DB or a MockDB,
	// depending on whether the Datastorer sql.DB is nil or not
	s, err := movieDatastore.NewSelector(ctl.App.Datastorer.DB())
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	// Find the Movie by ID using the selector.FindByID method
	m, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	tx, err := ctl.App.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// NewTransactor gives back either a MockTx or a concrete
	// Tx, depending upon whether the Tx from Datastorer is nil or not
	t, err := movieDatastore.NewTransactor(tx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// Delete method of Transactor physically deletes the record
	// from the DB, unless mocked
	err = t.Delete(ctx, m)
	if err != nil {
		return nil, errs.E(op, ctl.App.Datastorer.RollbackTx(tx, err))
	}

	// Commit the Transaction
	if err := ctl.App.Datastorer.CommitTx(tx); err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	// Populate the response
	response := newDeleteMovieResponse(m, ctl.SRF)

	return response, nil
}

// FindByID finds a movie given its' unique ID
func (ctl *MovieController) FindByID(ctx context.Context, id string) (*SingleMovieResponse, error) {
	const op errs.Op = "controller/movieController/MovieController.FindByID"

	// NewSelector returns either a concrete DB or a MockDB,
	// depending on whether the Datastorer sql.DB is nil or not
	s, err := movieDatastore.NewSelector(ctl.App.Datastorer.DB())
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	// Find the Movie by ID using the selector.FindByID method
	m, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	rd, err := newMovieResponse(m)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// Populate the response
	response := ctl.NewSingleMovieResponse(rd)

	return response, nil
}

// FindAll finds the entire set of Movies
func (ctl *MovieController) FindAll(ctx context.Context) (*ListMovieResponse, error) {
	const op errs.Op = "controller/movieController/MovieController.FindAll"

	// NewSelector returns either a concrete DB or a MockDB,
	// depending on whether the Datastorer sql.DB is nil or not
	s, err := movieDatastore.NewSelector(ctl.App.Datastorer.DB())
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	// Find the list of all Movies using the selector.FindAll method
	movies, err := s.FindAll(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// Populate the response
	response, err := ctl.NewListMovieResponse(movies)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return response, nil
}

// NewListMovieResponse is an initializer for ListMovieResponse
func (ctl *MovieController) NewListMovieResponse(ms []*movie.Movie) (*ListMovieResponse, error) {
	const op errs.Op = "controller/movieController/MovieController.NewListMovieResponse"

	var s []*ResponseData

	for _, m := range ms {
		mr, err := newMovieResponse(m)
		if err != nil {
			return nil, errs.E(op, err)
		}
		s = append(s, mr)
	}

	return &ListMovieResponse{StandardResponseFields: ctl.SRF, Data: s}, nil
}

// NewSingleMovieResponse is an initializer for SingleMovieResponse
func (ctl *MovieController) NewSingleMovieResponse(mr *ResponseData) *SingleMovieResponse {
	return &SingleMovieResponse{StandardResponseFields: ctl.SRF, Data: mr}
}

// NewMovie is an initializer for the Movie struct
func (ctl *MovieController) newAdder(am *RequestData) (movie.Adder, error) {
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

	if ctl.App.Mock {
		return &movie.MockMovie{
			Title:    am.Title,
			Year:     am.Year,
			Rated:    am.Rated,
			Released: t,
			RunTime:  am.RunTime,
			Director: am.Director,
			Writer:   am.Writer,
		}, nil
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

// NewMovie is an initializer for the Movie struct
func (ctl *MovieController) newUpdater(am *RequestData) (movie.Updater, error) {
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

	if ctl.App.Mock {
		return &movie.MockMovie{
			Title:    am.Title,
			Year:     am.Year,
			Rated:    am.Rated,
			Released: t,
			RunTime:  am.RunTime,
			Director: am.Director,
			Writer:   am.Writer,
		}, nil
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
