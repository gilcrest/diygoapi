package movieController

import (
	"context"
	"time"

	"github.com/gilcrest/errs"
	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/controller"
	"github.com/gilcrest/go-api-basic/controller/authcontroller"
	"github.com/gilcrest/go-api-basic/datastore/moviestore"
	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/gilcrest/go-api-basic/domain/user"
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
	CreateUsername  string `json:"create_username"`
	CreateTimestamp string `json:"create_timestamp"`
	UpdateUsername  string `json:"update_username"`
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
func newMovieResponse(m *movie.Movie) (*ResponseData, error) {
	const op errs.Op = "controller/movieController/newMovieResponse"

	return &ResponseData{
		ExternalID:      m.ExternalID,
		Title:           m.Title,
		Year:            m.Year,
		Rated:           m.Rated,
		Released:        m.Released.Format(time.RFC3339),
		RunTime:         m.RunTime,
		Director:        m.Director,
		Writer:          m.Writer,
		CreateUsername:  m.CreateUsername,
		CreateTimestamp: m.CreateTimestamp.Format(time.RFC3339),
		UpdateUsername:  m.UpdateUsername,
		UpdateTimestamp: m.UpdateTimestamp.Format(time.RFC3339),
	}, nil
}

// NewMovieController initializes MovieController
func NewMovieController(app *app.Application, srf controller.StandardResponseFields) *MovieController {
	return &MovieController{App: app, SRF: srf}
}

// AddMovie adds a movie to the catalog.
func (ctl *MovieController) AddMovie(ctx context.Context, r *RequestData, token string) (*SingleMovieResponse, error) {
	const op errs.Op = "controller/movieController/MovieController.AddMovie"

	// authorize and get user from token
	u, err := authcontroller.AuthorizeAccessToken(ctx, token)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// Construct request/user into a movie.Movie struct
	m, err := ctl.newMovie(r, u)
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

	// declare variable as the Transactor interface
	var movieTransactor moviestore.Transactor

	// If app is in Mock mode, use MockTx to satisfy the interface,
	// otherwise use a true sql.Tx
	if ctl.App.Mock {
		movieTransactor = moviestore.NewMockTx()
	} else {
		movieTransactor, err = moviestore.NewTx(tx)
		if err != nil {
			return nil, errs.E(op, err)
		}
	}

	// Call the Create method of the Transactor to insert data to
	// the database (unless mocked, of course). If an error occurs,
	// rollback the transaction
	err = movieTransactor.Create(ctx, m)
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

// Update updates the movie given the external id sent in
func (ctl *MovieController) Update(ctx context.Context, externalID string, r *RequestData, token string) (*SingleMovieResponse, error) {
	const op errs.Op = "controller/movieController/MovieController.Update"

	// authorize and get user from token
	u, err := authcontroller.AuthorizeAccessToken(ctx, token)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// Convert request into a Movie struct
	m, err := ctl.newMovie4Update(r, externalID, u)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// Perform domain Update "business logic"
	err = m.Update(ctx, externalID)
	if err != nil {
		return nil, errs.E(err)
	}

	// Begin a DB Tx, if the underlying struct is a MockDatastore then
	// the Tx will be nil
	tx, err := ctl.App.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// declare variable as the Transactor interface
	var movieTransactor moviestore.Transactor

	// If app is in Mock mode, use MockTx to satisfy the interface,
	// otherwise use a true sql.Tx for moviestore.Tx
	if ctl.App.Mock {
		movieTransactor = moviestore.NewMockTx()
	} else {
		movieTransactor, err = moviestore.NewTx(tx)
		if err != nil {
			return nil, errs.E(op, err)
		}
	}

	// Call the Update method of the Transactor to update data on
	// the database (unless mocked, of course). If an error occurs,
	// rollback the transaction
	err = movieTransactor.Update(ctx, m)
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
func (ctl *MovieController) Delete(ctx context.Context, id string, token string) (*DeleteMovieResponse, error) {
	const op errs.Op = "controller/movieController/MovieController.Delete"

	// authorize and get user from token
	u, err := authcontroller.AuthorizeAccessToken(ctx, token)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// TODO something to properly authorize Delete
	ctl.App.Logger.Info().
		Str("email", u.Email).
		Str("first name", u.FirstName).
		Str("last name", u.LastName).
		Str("full name", u.FullName).
		Msgf("Delete authorized for %s", u.Email)

	// declare variable as the Transactor interface
	var movieSelector moviestore.Selector

	// If app is in Mock mode, use MockDB to satisfy the interface,
	// otherwise use a true sql.DB for moviestore.DB
	if ctl.App.Mock {
		movieSelector = moviestore.NewMockDB()
	} else {
		movieSelector, err = moviestore.NewDB(ctl.App.Datastorer.DB())
		if err != nil {
			return nil, errs.E(op, err)
		}
	}

	// Find the Movie by ID using the selector.FindByID method
	m, err := movieSelector.FindByID(ctx, id)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// start a new database transaction
	tx, err := ctl.App.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// declare variable as the Transactor interface
	var movieTransactor moviestore.Transactor

	// If app is in Mock mode, use MockTx to satisfy the interface,
	// otherwise use a true sql.Tx for moviestore.Tx
	if ctl.App.Mock {
		movieTransactor = moviestore.NewMockTx()
	} else {
		movieTransactor, err = moviestore.NewTx(tx)
		if err != nil {
			return nil, errs.E(op, err)
		}
	}

	// Delete method of Transactor physically deletes the record
	// from the DB, unless mocked
	err = movieTransactor.Delete(ctx, m)
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
func (ctl *MovieController) FindByID(ctx context.Context, id string, token string) (*SingleMovieResponse, error) {
	const op errs.Op = "controller/movieController/MovieController.FindByID"

	// authorize and get user from token
	u, err := authcontroller.AuthorizeAccessToken(ctx, token)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// TODO something to properly authorize FindByID
	ctl.App.Logger.Info().
		Str("email", u.Email).
		Str("first name", u.FirstName).
		Str("last name", u.LastName).
		Str("full name", u.FullName).
		Msgf("Delete authorized for %s", u.Email)

	// declare variable as the Transactor interface
	var movieSelector moviestore.Selector

	// If app is in Mock mode, use MockDB to satisfy the interface,
	// otherwise use a true sql.DB for moviestore.DB
	if ctl.App.Mock {
		movieSelector = moviestore.NewMockDB()
	} else {
		movieSelector, err = moviestore.NewDB(ctl.App.Datastorer.DB())
		if err != nil {
			return nil, errs.E(op, err)
		}
	}

	// Find the Movie by ID using the selector.FindByID method
	m, err := movieSelector.FindByID(ctx, id)
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
func (ctl *MovieController) FindAll(ctx context.Context, token string) (*ListMovieResponse, error) {
	const op errs.Op = "controller/movieController/MovieController.FindAll"

	// authorize and get user from token
	u, err := authcontroller.AuthorizeAccessToken(ctx, token)
	if err != nil {
		return nil, errs.E(op, err)
	}

	// TODO something to properly authorize FindByID
	ctl.App.Logger.Info().
		Str("email", u.Email).
		Str("first name", u.FirstName).
		Str("last name", u.LastName).
		Str("full name", u.FullName).
		Msgf("Delete authorized for %s", u.Email)

	// declare variable as the Transactor interface
	var movieSelector moviestore.Selector

	// If app is in Mock mode, use MockDB to satisfy the interface,
	// otherwise use a true sql.DB for moviestore.DB
	if ctl.App.Mock {
		movieSelector = moviestore.NewMockDB()
	} else {
		movieSelector, err = moviestore.NewDB(ctl.App.Datastorer.DB())
		if err != nil {
			return nil, errs.E(op, err)
		}
	}

	// Find the list of all Movies using the selector.FindAll method
	movies, err := movieSelector.FindAll(ctx)
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

// newMovie is an initializer for the Movie struct
func (ctl *MovieController) newMovie(rd *RequestData, u *user.User) (*movie.Movie, error) {
	const op errs.Op = "controller/movieController/newMovie"

	// Parse Release Date according to RFC3339
	t, err := time.Parse(time.RFC3339, rd.Released)
	if err != nil {
		return nil, errs.E(op,
			errs.Validation,
			errs.Code("invalid_date_format"),
			errs.Parameter("ReleaseDate"),
			err)
	}

	return &movie.Movie{
		Title:          rd.Title,
		Year:           rd.Year,
		Rated:          rd.Rated,
		Released:       t,
		RunTime:        rd.RunTime,
		Director:       rd.Director,
		Writer:         rd.Writer,
		CreateUsername: u.Email,
		UpdateUsername: u.Email,
	}, nil
}

// newMovie4Update is an initializer for the Movie struct for the
// update operation
func (ctl *MovieController) newMovie4Update(rd *RequestData, externalID string, u *user.User) (*movie.Movie, error) {
	const op errs.Op = "controller/movieController/newMovie4Update"

	// Parse Release Date according to RFC3339
	t, err := time.Parse(time.RFC3339, rd.Released)
	if err != nil {
		return nil, errs.E(op,
			errs.Validation,
			errs.Code("invalid_date_format"),
			errs.Parameter("ReleaseDate"),
			err)
	}

	return &movie.Movie{
		ExternalID:     externalID,
		Title:          rd.Title,
		Year:           rd.Year,
		Rated:          rd.Rated,
		Released:       t,
		RunTime:        rd.RunTime,
		Director:       rd.Director,
		Writer:         rd.Writer,
		UpdateUsername: u.Email,
	}, nil
}
