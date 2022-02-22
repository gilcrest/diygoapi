package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/datastore/moviestore"
	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/gilcrest/go-api-basic/domain/secure"
)

// CreateMovieRequest is the request struct for Creating a Movie
type CreateMovieRequest struct {
	Title    string `json:"title"`
	Rated    string `json:"rated"`
	Released string `json:"release_date"`
	RunTime  int    `json:"run_time"`
	Director string `json:"director"`
	Writer   string `json:"writer"`
}

// MovieResponse is the response struct for a Movie
type MovieResponse struct {
	ExternalID  string        `json:"external_id"`
	Title       string        `json:"title"`
	Rated       string        `json:"rated"`
	Released    string        `json:"release_date"`
	RunTime     int           `json:"run_time"`
	Director    string        `json:"director"`
	Writer      string        `json:"writer"`
	CreateAudit auditResponse `json:"create_audit"`
	UpdateAudit auditResponse `json:"update_audit"`
}

// newMovieResponse initializes MovieResponse
func newMovieResponse(m movie.Movie, sa audit.SimpleAudit) MovieResponse {
	return MovieResponse{
		ExternalID:  m.ExternalID.String(),
		Title:       m.Title,
		Rated:       m.Rated,
		Released:    m.Released.Format(time.RFC3339),
		RunTime:     m.RunTime,
		Director:    m.Director,
		Writer:      m.Writer,
		CreateAudit: newAuditResponse(sa.First),
		UpdateAudit: newAuditResponse(sa.Last),
	}
}

// CreateMovieService is a service for creating a Movie
type CreateMovieService struct {
	Datastorer Datastorer
}

// Create is used to create an Movie
func (s CreateMovieService) Create(ctx context.Context, r *CreateMovieRequest, adt audit.Audit) (MovieResponse, error) {
	var err error

	released, err := time.Parse(time.RFC3339, r.Released)
	if err != nil {
		return MovieResponse{}, errs.E(errs.Validation,
			errs.Code("invalid_date_format"),
			errs.Parameter("release_date"),
			err)
	}

	// initialize Movie and inject dependent fields
	m := movie.Movie{
		ID:         uuid.New(),
		ExternalID: secure.NewID(),
		Title:      r.Title,
		Rated:      r.Rated,
		Released:   released,
		RunTime:    r.RunTime,
		Director:   r.Director,
		Writer:     r.Writer,
	}

	sa := audit.SimpleAudit{
		First: adt,
		Last:  adt,
	}

	err = m.IsValid()
	if err != nil {
		return MovieResponse{}, err
	}

	createMovieParams := moviestore.CreateMovieParams{
		MovieID:         m.ID,
		ExtlID:          m.ExternalID.String(),
		Title:           m.Title,
		Rated:           datastore.NewNullString(m.Rated),
		Released:        datastore.NewNullTime(released),
		RunTime:         datastore.NewNullInt32(int32(m.RunTime)),
		Director:        datastore.NewNullString(m.Director),
		Writer:          datastore.NewNullString(m.Writer),
		CreateAppID:     sa.First.App.ID,
		CreateUserID:    datastore.NewNullUUID(sa.First.User.ID),
		CreateTimestamp: sa.First.Moment,
		UpdateAppID:     sa.Last.App.ID,
		UpdateUserID:    datastore.NewNullUUID(sa.Last.User.ID),
		UpdateTimestamp: sa.Last.Moment,
	}

	// start db txn using pgxpool
	tx, err := s.Datastorer.BeginTx(ctx)
	if err != nil {
		return MovieResponse{}, err
	}

	_, err = moviestore.New(tx).CreateMovie(ctx, createMovieParams)
	if err != nil {
		return MovieResponse{}, errs.E(errs.Database, s.Datastorer.RollbackTx(ctx, tx, err))
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return MovieResponse{}, err
	}

	return newMovieResponse(m, sa), nil
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

// UpdateMovieService is a service for updating a Movie
type UpdateMovieService struct {
	Datastorer Datastorer
}

// Update is used to update a movie
func (s UpdateMovieService) Update(ctx context.Context, r *UpdateMovieRequest, adt audit.Audit) (MovieResponse, error) {

	// retrieve existing Movie
	var (
		m   movie.Movie
		dbm moviestore.Movie
		sa  audit.SimpleAudit
		err error
	)

	var released time.Time
	released, err = time.Parse(time.RFC3339, r.Released)
	if err != nil {
		return MovieResponse{}, errs.E(errs.Validation,
			errs.Code("invalid_date_format"),
			errs.Parameter("release_date"),
			err)
	}

	dbm, err = moviestore.New(s.Datastorer.Pool()).FindMovieByExternalID(ctx, r.ExternalID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return MovieResponse{}, errs.E(errs.Validation, "No movie exists for the given external ID")
		}
		return MovieResponse{}, errs.E(errs.Database, err)
	}

	m, sa, err = hydrateMovieFromDB(ctx, s.Datastorer.Pool(), dbm)
	if err != nil {
		return MovieResponse{}, err
	}

	// update fields from request
	m.Title = r.Title
	m.Rated = r.Rated
	m.Released = released
	m.RunTime = r.RunTime
	m.Director = r.Director
	m.Writer = r.Writer

	err = m.IsValid()
	if err != nil {
		return MovieResponse{}, err
	}

	// update audit with latest
	sa.Last = adt

	updateMovieParams := moviestore.UpdateMovieParams{
		Title:           m.Title,
		Rated:           datastore.NewNullString(m.Rated),
		Released:        datastore.NewNullTime(released),
		RunTime:         datastore.NewNullInt32(int32(m.RunTime)),
		Director:        datastore.NewNullString(m.Director),
		Writer:          datastore.NewNullString(m.Writer),
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    datastore.NewNullUUID(adt.User.ID),
		UpdateTimestamp: adt.Moment,
		MovieID:         m.ID,
	}

	// start db txn using pgxpool
	tx, err := s.Datastorer.BeginTx(ctx)
	if err != nil {
		return MovieResponse{}, err
	}

	err = moviestore.New(tx).UpdateMovie(ctx, updateMovieParams)
	if err != nil {
		return MovieResponse{}, errs.E(errs.Database, s.Datastorer.RollbackTx(ctx, tx, err))
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return MovieResponse{}, err
	}

	return newMovieResponse(m, sa), nil
}

// DeleteMovieResponse is the response struct for deleted Movies
type DeleteMovieResponse struct {
	ExternalID string `json:"extl_id"`
	Deleted    bool   `json:"deleted"`
}

// DeleteMovieService is a service for deleting a Movie
type DeleteMovieService struct {
	Datastorer Datastorer
}

// Delete is used to delete a movie
func (s DeleteMovieService) Delete(ctx context.Context, extlID string) (DeleteMovieResponse, error) {

	// retrieve existing Movie
	var (
		dbm moviestore.Movie
		err error
	)
	dbm, err = moviestore.New(s.Datastorer.Pool()).FindMovieByExternalID(ctx, extlID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return DeleteMovieResponse{}, errs.E(errs.Validation, "No movie exists for the given external ID")
		}
		return DeleteMovieResponse{}, errs.E(errs.Database, err)
	}

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return DeleteMovieResponse{}, err
	}

	err = moviestore.New(tx).DeleteMovie(ctx, dbm.MovieID)
	if err != nil {
		return DeleteMovieResponse{}, errs.E(errs.Database, s.Datastorer.RollbackTx(ctx, tx, err))
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return DeleteMovieResponse{}, err
	}

	response := DeleteMovieResponse{
		ExternalID: dbm.ExtlID,
		Deleted:    true,
	}

	return response, nil
}

// FindMovieService is a service for reading Movies from the DB
type FindMovieService struct {
	Datastorer Datastorer
}

// FindMovieByID is used to find an individual movie
func (s FindMovieService) FindMovieByID(ctx context.Context, extlID string) (MovieResponse, error) {

	var (
		m   movie.Movie
		dbm moviestore.Movie
		sa  audit.SimpleAudit
		err error
	)

	dbm, err = moviestore.New(s.Datastorer.Pool()).FindMovieByExternalID(ctx, extlID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return MovieResponse{}, errs.E(errs.Validation, "no movie exists for the given external ID")
		}
		return MovieResponse{}, errs.E(errs.Database, err)
	}

	m, sa, err = hydrateMovieFromDB(ctx, s.Datastorer.Pool(), dbm)
	if err != nil {
		return MovieResponse{}, err
	}

	return newMovieResponse(m, sa), nil
}

// FindAllMovies is used to list all movies in the db
func (s FindMovieService) FindAllMovies(ctx context.Context) ([]MovieResponse, error) {

	var response []MovieResponse

	movies, err := moviestore.New(s.Datastorer.Pool()).FindMovies(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.E(errs.Validation, "no movies exists")
		}
		return nil, errs.E(errs.Database, err)
	}

	for _, dbm := range movies {
		var (
			m  movie.Movie
			sa audit.SimpleAudit
		)
		m, sa, err = hydrateMovieFromDB(ctx, s.Datastorer.Pool(), dbm)
		if err != nil {
			return nil, err
		}

		mr := newMovieResponse(m, sa)
		response = append(response, mr)
	}

	return response, nil
}

// hydrateMovieFromDB populates a movie.Movie and an audit.SimpleAudit given a moviestore.Movie
func hydrateMovieFromDB(ctx context.Context, dbtx DBTX, dbm moviestore.Movie) (movie.Movie, audit.SimpleAudit, error) {
	var (
		err error
	)

	// Convert moviestore.Movie into a domain movie.Movie struct
	m := movie.Movie{
		ID:         dbm.MovieID,
		ExternalID: secure.MustParseIdentifier(dbm.ExtlID),
		Title:      dbm.Title,
		Rated:      dbm.Rated.String,
		Released:   dbm.Released.Time,
		RunTime:    int(dbm.RunTime.Int32),
		Director:   dbm.Director.String,
		Writer:     dbm.Writer.String,
	}

	var createAudit audit.Audit
	createAudit, err = newAudit(ctx, dbtx, dbm.CreateAppID, dbm.CreateUserID, dbm.CreateTimestamp)
	if err != nil {
		return movie.Movie{}, audit.SimpleAudit{}, err
	}

	var updateAudit audit.Audit
	updateAudit, err = newAudit(ctx, dbtx, dbm.UpdateAppID, dbm.UpdateUserID, dbm.UpdateTimestamp)
	if err != nil {
		return movie.Movie{}, audit.SimpleAudit{}, err
	}

	sa := audit.SimpleAudit{
		First: createAudit,
		Last:  updateAudit,
	}

	return m, sa, nil
}
