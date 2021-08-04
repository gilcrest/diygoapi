package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/gilcrest/go-api-basic/domain/user"
)

// RandomStringGenerator is the interface that generates random strings
type RandomStringGenerator interface {
	CryptoString(n int) (string, error)
}

// MovieTransactor is the interface that wraps the DML actions for
// a Movie in the DB
type MovieTransactor interface {
	Create(ctx context.Context, m *movie.Movie) error
	Update(ctx context.Context, m *movie.Movie) error
	Delete(ctx context.Context, m *movie.Movie) error
}

// MovieSelector reads Movie records from the db
type MovieSelector interface {
	FindByID(context.Context, string) (*movie.Movie, error)
	FindAll(context.Context) ([]*movie.Movie, error)
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
	ExternalID      string `json:"external_id"`
	Title           string `json:"title"`
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

// DeleteMovieResponse is the response struct for deleted Movies
type DeleteMovieResponse struct {
	ExternalID string `json:"extl_id"`
	Deleted    bool   `json:"deleted"`
}

func newMovieResponse(m *movie.Movie) MovieResponse {
	return MovieResponse{
		ExternalID:      m.ExternalID,
		Title:           m.Title,
		Rated:           m.Rated,
		Released:        m.Released.Format(time.RFC3339),
		RunTime:         m.RunTime,
		Director:        m.Director,
		Writer:          m.Writer,
		CreateUsername:  m.CreateUser.Email,
		CreateTimestamp: m.CreateTime.Format(time.RFC3339),
		UpdateUsername:  m.UpdateUser.Email,
		UpdateTimestamp: m.UpdateTime.Format(time.RFC3339),
	}
}

// CreateMovieService is a service for creating a Movie
type CreateMovieService struct {
	RandomStringGenerator RandomStringGenerator
	MovieTransactor       MovieTransactor
}

// NewCreateMovieService is an initializer for CreateMovieService
func NewCreateMovieService(rsg RandomStringGenerator, mt MovieTransactor) *CreateMovieService {
	return &CreateMovieService{RandomStringGenerator: rsg, MovieTransactor: mt}
}

// Create is used to create a Movie
func (cms CreateMovieService) Create(ctx context.Context, r *CreateMovieRequest, u user.User) (MovieResponse, error) {

	mr := MovieResponse{}

	extlID, err := cms.RandomStringGenerator.CryptoString(15)
	if err != nil {
		return mr, err
	}

	// Call the NewMovie method for struct initialization
	m, err := movie.NewMovie(uuid.New(), extlID, u)
	if err != nil {
		return mr, err
	}

	m, err = m.SetReleased(r.Released)
	if err != nil {
		return mr, err
	}
	m.SetTitle(r.Title).
		SetRated(r.Rated).
		SetRunTime(r.RunTime).
		SetDirector(r.Director).
		SetWriter(r.Writer)

	err = m.IsValid()
	if err != nil {
		return mr, err
	}

	// Call the Create method of the Transactor to insert data to
	// the database. If an error occurs, rollback the transaction
	err = cms.MovieTransactor.Create(ctx, m)
	if err != nil {
		return mr, err
	}

	return newMovieResponse(m), nil
}

// UpdateMovieService is a service for updating a Movie
type UpdateMovieService struct {
	MovieTransactor MovieTransactor
}

// NewUpdateMovieService is an initializer for UpdateMovieService
func NewUpdateMovieService(mt MovieTransactor) *UpdateMovieService {
	return &UpdateMovieService{MovieTransactor: mt}
}

// Update is used to update a movie
func (ums UpdateMovieService) Update(ctx context.Context, r *UpdateMovieRequest, u user.User) (MovieResponse, error) {

	mr := MovieResponse{}

	// Convert request into a Movie struct
	m := new(movie.Movie)
	m.SetExternalID(r.ExternalID)
	m.SetTitle(r.Title)
	m.SetRated(r.Rated)
	m, err := m.SetReleased(r.Released)
	if err != nil {
		return mr, err
	}
	m.SetRunTime(r.RunTime)
	m.SetDirector(r.Director)
	m.SetWriter(r.Writer)
	m.SetUpdateUser(u)
	m.SetUpdateTime()

	err = m.IsValid()
	if err != nil {
		return mr, err
	}

	// Call the Update method of the Transactor to update the record
	// in the database.
	err = ums.MovieTransactor.Update(ctx, m)
	if err != nil {
		return mr, err
	}

	return newMovieResponse(m), nil
}

// DeleteMovieService is a service for deleting a Movie
type DeleteMovieService struct {
	MovieSelector   MovieSelector
	MovieTransactor MovieTransactor
}

// NewDeleteMovieService is an initializer for DeleteMovieService
func NewDeleteMovieService(ms MovieSelector, mt MovieTransactor) *DeleteMovieService {
	return &DeleteMovieService{MovieSelector: ms, MovieTransactor: mt}
}

// Delete is used to delete a movie
func (dms DeleteMovieService) Delete(ctx context.Context, extlID string) (DeleteMovieResponse, error) {

	dmr := DeleteMovieResponse{}

	// Find the Movie by ID using the selector.FindMovieByID method
	// It's arguable I don't need to do this and can just send
	// the external ID to the database Transactor directly instead,
	// (I'd have to rework it slightly) but this way works as an
	// example
	m, err := dms.MovieSelector.FindByID(ctx, extlID)
	if err != nil {
		return dmr, err
	}

	// Delete method of Transactor physically deletes the record
	// from the DB, unless mocked
	err = dms.MovieTransactor.Delete(ctx, m)
	if err != nil {
		return dmr, err
	}

	response := DeleteMovieResponse{
		ExternalID: m.ExternalID,
		Deleted:    true,
	}

	return response, nil
}

// FindMovieService is a service for reading a Movie from the DB
type FindMovieService struct {
	MovieSelector MovieSelector
}

// NewFindMovieService is an initializer for FindMovieService
func NewFindMovieService(ms MovieSelector) *FindMovieService {
	return &FindMovieService{MovieSelector: ms}
}

// FindMovieByID is used to find an individual movie
func (fms FindMovieService) FindMovieByID(ctx context.Context, extlID string) (MovieResponse, error) {

	mr := MovieResponse{}

	m, err := fms.MovieSelector.FindByID(ctx, extlID)
	if err != nil {
		return mr, err
	}

	return newMovieResponse(m), nil
}

// FindAllMovies is used to list all movies in the db
func (fms FindMovieService) FindAllMovies(ctx context.Context) ([]MovieResponse, error) {

	var response []MovieResponse

	// Find the list of all Movies using the selector.FindAll method
	movies, err := fms.MovieSelector.FindAll(ctx)
	if err != nil {
		return response, err
	}

	for _, m := range movies {
		mr := newMovieResponse(m)
		response = append(response, mr)
	}

	return response, nil
}
