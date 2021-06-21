package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gilcrest/go-api-basic/domain/user"

	"github.com/gilcrest/go-api-basic/datastore/moviestore"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/gilcrest/go-api-basic/domain/random"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/hlog"
)

// NewCreateMovieHandler is a provider for CreateMovieHandler
func NewCreateMovieHandler(h DefaultMovieHandlers) CreateMovieHandler {
	return http.HandlerFunc(h.CreateMovie)
}

// NewUpdateMovieHandler is a provider for UpdateMovieHandler
func NewUpdateMovieHandler(h DefaultMovieHandlers) UpdateMovieHandler {
	return http.HandlerFunc(h.UpdateMovie)
}

// NewDeleteMovieHandler is a provider for DeleteMovieHandler
func NewDeleteMovieHandler(h DefaultMovieHandlers) DeleteMovieHandler {
	return http.HandlerFunc(h.DeleteMovie)
}

// NewFindMovieByIDHandler is a provider for the
// FindMovieByIDHandler for wire
func NewFindMovieByIDHandler(h DefaultMovieHandlers) FindMovieByIDHandler {
	return http.HandlerFunc(h.FindByID)
}

// NewFindAllMoviesHandler is a provider for FindAllMoviesHandler
func NewFindAllMoviesHandler(h DefaultMovieHandlers) FindAllMoviesHandler {
	return http.HandlerFunc(h.FindAllMovies)
}

// DefaultMovieHandlers are the default handlers for CRUD operations
// for a Movie. Each method on the struct is a separate handler.
type DefaultMovieHandlers struct {
	RandomStringGenerator random.StringGenerator
	Transactor            moviestore.Transactor
	Selector              moviestore.Selector
}

// CreateMovie is a HandlerFunc used to create a Movie
func (h DefaultMovieHandlers) CreateMovie(w http.ResponseWriter, r *http.Request) {
	// createMovieRequestBody is the request struct for Create
	type createMovieRequestBody struct {
		Title    string `json:"title"`
		Rated    string `json:"rated"`
		Released string `json:"release_date"`
		RunTime  int    `json:"run_time"`
		Director string `json:"director"`
		Writer   string `json:"writer"`
	}

	// CreateMovieResponse is the response struct for a Movie
	type createMovieResponse struct {
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

	logger := *hlog.FromRequest(r)

	u, err := user.FromRequest(r)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// Declare request body (rb) as an instance of createMovieRequestBody
	rb := new(createMovieRequestBody)

	// Decode JSON HTTP request body into a Decoder type
	// and unmarshal that into the MovieRequest struct in the
	// AddMovieHandler
	err = json.NewDecoder(r.Body).Decode(&rb)
	defer r.Body.Close()
	// Call DecoderErr to determine if body is nil, json is malformed
	// or any other error
	err = DecoderErr(err)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	extlID, err := h.RandomStringGenerator.CryptoString(15)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// Call the NewMovie method for struct initialization
	m, err := movie.NewMovie(uuid.New(), extlID, u)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	m, err = m.SetReleased(rb.Released)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}
	m.SetTitle(rb.Title).
		SetRated(rb.Rated).
		SetRunTime(rb.RunTime).
		SetDirector(rb.Director).
		SetWriter(rb.Writer)

	err = m.IsValid()
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// Call the Create method of the Transactor to insert data to
	// the database (unless mocked, of course). If an error occurs,
	// rollback the transaction
	err = h.Transactor.Create(r.Context(), m)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	response := createMovieResponse{
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

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
		return
	}
}

// UpdateMovie handles PUT requests for the /movies/{id} endpoint
// and updates the given movie
func (h DefaultMovieHandlers) UpdateMovie(w http.ResponseWriter, r *http.Request) {
	// updateMovieRequestBody is the request struct for Update
	type updateMovieRequestBody struct {
		Title    string `json:"title"`
		Rated    string `json:"rated"`
		Released string `json:"release_date"`
		RunTime  int    `json:"run_time"`
		Director string `json:"director"`
		Writer   string `json:"writer"`
	}

	// updateMovieResponse is the response struct for a Movie
	type updateMovieResponse struct {
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

	logger := *hlog.FromRequest(r)

	u, err := user.FromRequest(r)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// gorilla mux Vars function returns the route variables for the
	// current request, if any. id is the external id given for the
	// movie
	vars := mux.Vars(r)
	extlid := vars["extlID"]

	// Declare rb as an instance of updateMovieRequestBody
	rb := new(updateMovieRequestBody)

	// Decode JSON HTTP request body into a Decoder type
	// and unmarshal that into requestData
	err = json.NewDecoder(r.Body).Decode(&rb)
	defer r.Body.Close()
	// Call DecoderErr to determine if body is nil, json is malformed
	// or any other error
	err = DecoderErr(err)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// Convert request into a Movie struct
	m := new(movie.Movie)
	m.SetExternalID(extlid)
	m.SetTitle(rb.Title)
	m.SetRated(rb.Rated)
	m, err = m.SetReleased(rb.Released)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}
	m.SetRunTime(rb.RunTime)
	m.SetDirector(rb.Director)
	m.SetWriter(rb.Writer)
	m.SetUpdateUser(u)
	m.SetUpdateTime()

	err = m.IsValid()
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// Call the Update method of the Transactor to update the record
	// in the database.
	err = h.Transactor.Update(r.Context(), m)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	response := updateMovieResponse{
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

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
		return
	}
}

// DeleteMovie handles DELETE requests for the /movies/{id} endpoint
// and updates the given movie
func (h DefaultMovieHandlers) DeleteMovie(w http.ResponseWriter, r *http.Request) {
	// deleteMovieResponse is the response struct for deleted Movies
	type deleteMovieResponse struct {
		ExternalID string `json:"extl_id"`
		Deleted    bool   `json:"deleted"`
	}

	logger := *hlog.FromRequest(r)

	// gorilla mux Vars function returns the route variables for the
	// current request, if any. id is the external id given for the
	// movie
	vars := mux.Vars(r)
	extlid := vars["extlID"]

	// Find the Movie by ID using the selector.FindByID method
	// It's arguable I don't need to do this and can just send
	// the external ID to the database Transactor directly instead,
	// (I'd have to rework it slightly) but this way works as an
	// example
	m, err := h.Selector.FindByID(r.Context(), extlid)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// Delete method of Transactor physically deletes the record
	// from the DB, unless mocked
	err = h.Transactor.Delete(r.Context(), m)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	response := deleteMovieResponse{
		ExternalID: m.ExternalID,
		Deleted:    true,
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
		return
	}
}

// FindByID handles GET requests for the /movies/{id} endpoint
// and finds a movie by it's ID
func (h DefaultMovieHandlers) FindByID(w http.ResponseWriter, r *http.Request) {
	// movieResponse is the response struct for a Movie
	type movieResponse struct {
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

	logger := *hlog.FromRequest(r)

	// gorilla mux Vars function returns the route variables for the
	// current request, if any. id is the external id given for the
	// movie
	vars := mux.Vars(r)
	extlid := vars["extlID"]

	// Find the Movie by ID using the selector.FindByID method
	m, err := h.Selector.FindByID(r.Context(), extlid)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	response := movieResponse{
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

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
		return
	}
}

// FindAllMovies handles GET requests for the /movies endpoint and finds
// all movies
func (h DefaultMovieHandlers) FindAllMovies(w http.ResponseWriter, r *http.Request) {
	// movieResponse is the response struct for a Movie
	type movieResponse struct {
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

	logger := *hlog.FromRequest(r)

	// Find the list of all Movies using the selector.FindAll method
	movies, err := h.Selector.FindAll(r.Context())
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	var response []movieResponse
	for _, m := range movies {
		mr := movieResponse{
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
		response = append(response, mr)
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
		return
	}
}
