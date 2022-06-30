package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"

	"github.com/gilcrest/diy-go-api/datastore"
	"github.com/gilcrest/diy-go-api/datastore/moviestore"
	"github.com/gilcrest/diy-go-api/domain/app"
	"github.com/gilcrest/diy-go-api/domain/audit"
	"github.com/gilcrest/diy-go-api/domain/errs"
	"github.com/gilcrest/diy-go-api/domain/movie"
	"github.com/gilcrest/diy-go-api/domain/org"
	"github.com/gilcrest/diy-go-api/domain/person"
	"github.com/gilcrest/diy-go-api/domain/secure"
	"github.com/gilcrest/diy-go-api/domain/user"
)

// movieAudit is the combination of a domain Movie and its audit data
type movieAudit struct {
	Movie       movie.Movie
	SimpleAudit audit.SimpleAudit
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

// MovieResponse is the response struct for a Movie
type MovieResponse struct {
	ExternalID          string `json:"external_id"`
	Title               string `json:"title"`
	Rated               string `json:"rated"`
	Released            string `json:"release_date"`
	RunTime             int    `json:"run_time"`
	Director            string `json:"director"`
	Writer              string `json:"writer"`
	CreateAppExtlID     string `json:"create_app_extl_id"`
	CreateUsername      string `json:"create_username"`
	CreateUserFirstName string `json:"create_user_first_name"`
	CreateUserLastName  string `json:"create_user_last_name"`
	CreateDateTime      string `json:"create_date_time"`
	UpdateAppExtlID     string `json:"update_app_extl_id"`
	UpdateUsername      string `json:"update_username"`
	UpdateUserFirstName string `json:"update_user_first_name"`
	UpdateUserLastName  string `json:"update_user_last_name"`
	UpdateDateTime      string `json:"update_date_time"`
}

// newMovieResponse initializes MovieResponse
func newMovieResponse(ma movieAudit) MovieResponse {
	return MovieResponse{
		ExternalID:          ma.Movie.ExternalID.String(),
		Title:               ma.Movie.Title,
		Rated:               ma.Movie.Rated,
		Released:            ma.Movie.Released.Format(time.RFC3339),
		RunTime:             ma.Movie.RunTime,
		Director:            ma.Movie.Director,
		Writer:              ma.Movie.Writer,
		CreateAppExtlID:     ma.SimpleAudit.First.App.ExternalID.String(),
		CreateUsername:      ma.SimpleAudit.First.User.Username,
		CreateUserFirstName: ma.SimpleAudit.First.User.Profile.FirstName,
		CreateUserLastName:  ma.SimpleAudit.First.User.Profile.LastName,
		CreateDateTime:      ma.SimpleAudit.First.Moment.Format(time.RFC3339),
		UpdateAppExtlID:     ma.SimpleAudit.Last.App.ExternalID.String(),
		UpdateUsername:      ma.SimpleAudit.Last.User.Username,
		UpdateUserFirstName: ma.SimpleAudit.Last.User.Profile.FirstName,
		UpdateUserLastName:  ma.SimpleAudit.Last.User.Profile.LastName,
		UpdateDateTime:      ma.SimpleAudit.Last.Moment.Format(time.RFC3339),
	}
}

// CreateMovieService is a service for creating a Movie
type CreateMovieService struct {
	Datastorer Datastorer
}

// Create is used to create a Movie
func (s CreateMovieService) Create(ctx context.Context, r *CreateMovieRequest, adt audit.Audit) (mr MovieResponse, err error) {
	var released time.Time
	released, err = time.Parse(time.RFC3339, r.Released)
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
		CreateUserID:    sa.First.User.NullUUID(),
		CreateTimestamp: sa.First.Moment,
		UpdateAppID:     sa.Last.App.ID,
		UpdateUserID:    sa.Last.User.NullUUID(),
		UpdateTimestamp: sa.Last.Moment,
	}

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return MovieResponse{}, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	_, err = moviestore.New(tx).CreateMovie(ctx, createMovieParams)
	if err != nil {
		return MovieResponse{}, errs.E(errs.Database, err)
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return MovieResponse{}, err
	}

	mr = newMovieResponse(movieAudit{m, sa})

	return mr, nil
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
func (s UpdateMovieService) Update(ctx context.Context, r *UpdateMovieRequest, adt audit.Audit) (mr MovieResponse, err error) {

	var released time.Time
	released, err = time.Parse(time.RFC3339, r.Released)
	if err != nil {
		return MovieResponse{}, errs.E(errs.Validation,
			errs.Code("invalid_date_format"),
			errs.Parameter("release_date"),
			err)
	}

	// retrieve existing Movie
	var row moviestore.FindMovieByExternalIDWithAuditRow
	row, err = moviestore.New(s.Datastorer.Pool()).FindMovieByExternalIDWithAudit(ctx, r.ExternalID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return MovieResponse{}, errs.E(errs.Validation, "No movie exists for the given external ID")
		}
		return MovieResponse{}, errs.E(errs.Database, err)
	}

	m := movie.Movie{
		ID:         row.MovieID,
		ExternalID: secure.MustParseIdentifier(row.ExtlID),
		Title:      row.Title,
		Rated:      row.Rated.String,
		Released:   row.Released.Time,
		RunTime:    int(row.RunTime.Int32),
		Director:   row.Director.String,
		Writer:     row.Writer.String,
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

	sa := audit.SimpleAudit{
		First: audit.Audit{
			App: app.App{
				ID:          row.CreateAppID,
				ExternalID:  secure.MustParseIdentifier(row.CreateAppExtlID),
				Org:         org.Org{ID: row.CreateAppOrgID},
				Name:        row.CreateAppName,
				Description: row.CreateAppDescription,
				APIKeys:     nil,
			},
			User: user.User{
				ID:       row.CreateUserID.UUID,
				Username: row.CreateUsername,
				Org:      org.Org{ID: row.CreateUserOrgID},
				Profile: person.Profile{
					FirstName: row.CreateUserFirstName,
					LastName:  row.CreateUserLastName,
				},
			},
			Moment: row.CreateTimestamp,
		},
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
		UpdateUserID:    adt.User.NullUUID(),
		UpdateTimestamp: adt.Moment,
		MovieID:         m.ID,
	}

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return MovieResponse{}, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	err = moviestore.New(tx).UpdateMovie(ctx, updateMovieParams)
	if err != nil {
		return MovieResponse{}, errs.E(errs.Database, err)
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return MovieResponse{}, err
	}

	mr = newMovieResponse(movieAudit{m, sa})

	return mr, nil
}

// DeleteMovieService is a service for deleting a Movie
type DeleteMovieService struct {
	Datastorer Datastorer
}

// Delete is used to delete a movie
func (s DeleteMovieService) Delete(ctx context.Context, extlID string) (dr DeleteResponse, err error) {

	// retrieve existing Movie
	var dbm moviestore.Movie
	dbm, err = moviestore.New(s.Datastorer.Pool()).FindMovieByExternalID(ctx, extlID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return DeleteResponse{}, errs.E(errs.Validation, "No movie exists for the given external ID")
		}
		return DeleteResponse{}, errs.E(errs.Database, err)
	}

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return DeleteResponse{}, err
	}

	err = moviestore.New(tx).DeleteMovie(ctx, dbm.MovieID)
	if err != nil {
		return DeleteResponse{}, errs.E(errs.Database, err)
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return DeleteResponse{}, err
	}

	response := DeleteResponse{
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
func (s FindMovieService) FindMovieByID(ctx context.Context, extlID string) (mr MovieResponse, err error) {

	var row moviestore.FindMovieByExternalIDWithAuditRow
	row, err = moviestore.New(s.Datastorer.Pool()).FindMovieByExternalIDWithAudit(ctx, extlID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return MovieResponse{}, errs.E(errs.Validation, "no movie exists for the given external ID")
		}
		return MovieResponse{}, errs.E(errs.Database, err)
	}

	m := movie.Movie{
		ID:         row.MovieID,
		ExternalID: secure.MustParseIdentifier(row.ExtlID),
		Title:      row.Title,
		Rated:      row.Rated.String,
		Released:   row.Released.Time,
		RunTime:    int(row.RunTime.Int32),
		Director:   row.Director.String,
		Writer:     row.Writer.String,
	}

	sa := audit.SimpleAudit{
		First: audit.Audit{
			App: app.App{
				ID:          row.CreateAppID,
				ExternalID:  secure.MustParseIdentifier(row.CreateAppExtlID),
				Org:         org.Org{ID: row.CreateAppOrgID},
				Name:        row.CreateAppName,
				Description: row.CreateAppDescription,
				APIKeys:     nil,
			},
			User: user.User{
				ID:       row.CreateUserID.UUID,
				Username: row.CreateUsername,
				Org:      org.Org{ID: row.CreateUserOrgID},
				Profile: person.Profile{
					FirstName: row.CreateUserFirstName,
					LastName:  row.CreateUserLastName,
				},
			},
			Moment: row.CreateTimestamp,
		},
		Last: audit.Audit{
			App: app.App{
				ID:          row.UpdateAppID,
				ExternalID:  secure.MustParseIdentifier(row.UpdateAppExtlID),
				Org:         org.Org{ID: row.UpdateAppOrgID},
				Name:        row.UpdateAppName,
				Description: row.UpdateAppDescription,
				APIKeys:     nil,
			},
			User: user.User{
				ID:       row.UpdateUserID.UUID,
				Username: row.UpdateUsername,
				Org:      org.Org{ID: row.UpdateUserOrgID},
				Profile: person.Profile{
					FirstName: row.UpdateUserFirstName,
					LastName:  row.UpdateUserLastName,
				},
			},
			Moment: row.UpdateTimestamp,
		},
	}

	mr = newMovieResponse(movieAudit{m, sa})

	return mr, nil
}

// FindAllMovies is used to list all movies in the db
func (s FindMovieService) FindAllMovies(ctx context.Context) (smr []MovieResponse, err error) {

	var rows []moviestore.FindMoviesRow
	rows, err = moviestore.New(s.Datastorer.Pool()).FindMovies(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.E(errs.Validation, "no movies exists")
		}
		return nil, errs.E(errs.Database, err)
	}

	for _, row := range rows {
		m := movie.Movie{
			ID:         row.MovieID,
			ExternalID: secure.MustParseIdentifier(row.ExtlID),
			Title:      row.Title,
			Rated:      row.Rated.String,
			Released:   row.Released.Time,
			RunTime:    int(row.RunTime.Int32),
			Director:   row.Director.String,
			Writer:     row.Writer.String,
		}
		sa := audit.SimpleAudit{
			First: audit.Audit{
				App: app.App{
					ID:          row.CreateAppID,
					ExternalID:  secure.MustParseIdentifier(row.CreateAppExtlID),
					Org:         org.Org{ID: row.CreateAppOrgID},
					Name:        row.CreateAppName,
					Description: row.CreateAppDescription,
					APIKeys:     nil,
				},
				User: user.User{
					ID:       row.CreateUserID.UUID,
					Username: row.CreateUsername,
					Org:      org.Org{ID: row.CreateUserOrgID},
					Profile: person.Profile{
						FirstName: row.CreateUserFirstName,
						LastName:  row.CreateUserLastName,
					},
				},
				Moment: row.CreateTimestamp,
			},
			Last: audit.Audit{
				App: app.App{
					ID:          row.UpdateAppID,
					ExternalID:  secure.MustParseIdentifier(row.UpdateAppExtlID),
					Org:         org.Org{ID: row.UpdateAppOrgID},
					Name:        row.UpdateAppName,
					Description: row.UpdateAppDescription,
					APIKeys:     nil,
				},
				User: user.User{
					ID:       row.UpdateUserID.UUID,
					Username: row.UpdateUsername,
					Org:      org.Org{ID: row.UpdateUserOrgID},
					Profile: person.Profile{
						FirstName: row.UpdateUserFirstName,
						LastName:  row.UpdateUserLastName,
					},
				},
				Moment: row.UpdateTimestamp,
			},
		}
		mr := newMovieResponse(movieAudit{m, sa})
		smr = append(smr, mr)
	}

	return smr, nil
}
