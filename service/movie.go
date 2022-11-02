package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"

	"github.com/gilcrest/diy-go-api"
	"github.com/gilcrest/diy-go-api/errs"
	"github.com/gilcrest/diy-go-api/movie"
	"github.com/gilcrest/diy-go-api/secure"
	"github.com/gilcrest/diy-go-api/sqldb/datastore"
)

// movieAudit is the combination of a domain Movie and its audit data
type movieAudit struct {
	Movie       movie.Movie
	SimpleAudit diy.SimpleAudit
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
	CreateUserFirstName string `json:"create_user_first_name"`
	CreateUserLastName  string `json:"create_user_last_name"`
	CreateDateTime      string `json:"create_date_time"`
	UpdateAppExtlID     string `json:"update_app_extl_id"`
	UpdateUserFirstName string `json:"update_user_first_name"`
	UpdateUserLastName  string `json:"update_user_last_name"`
	UpdateDateTime      string `json:"update_date_time"`
}

// newMovieResponse initializes MovieResponse
func newMovieResponse(ma movieAudit) *MovieResponse {
	return &MovieResponse{
		ExternalID:          ma.Movie.ExternalID.String(),
		Title:               ma.Movie.Title,
		Rated:               ma.Movie.Rated,
		Released:            ma.Movie.Released.Format(time.RFC3339),
		RunTime:             ma.Movie.RunTime,
		Director:            ma.Movie.Director,
		Writer:              ma.Movie.Writer,
		CreateAppExtlID:     ma.SimpleAudit.Create.App.ExternalID.String(),
		CreateUserFirstName: ma.SimpleAudit.Create.User.FirstName,
		CreateUserLastName:  ma.SimpleAudit.Create.User.LastName,
		CreateDateTime:      ma.SimpleAudit.Create.Moment.Format(time.RFC3339),
		UpdateAppExtlID:     ma.SimpleAudit.Update.App.ExternalID.String(),
		UpdateUserFirstName: ma.SimpleAudit.Update.User.FirstName,
		UpdateUserLastName:  ma.SimpleAudit.Update.User.LastName,
		UpdateDateTime:      ma.SimpleAudit.Update.Moment.Format(time.RFC3339),
	}
}

// CreateMovieService is a service for creating a Movie
type CreateMovieService struct {
	Datastorer diy.Datastorer
}

// Create is used to create a Movie
func (s CreateMovieService) Create(ctx context.Context, r *CreateMovieRequest, adt diy.Audit) (mr *MovieResponse, err error) {
	var released time.Time
	released, err = time.Parse(time.RFC3339, r.Released)
	if err != nil {
		return nil, errs.E(errs.Validation,
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

	sa := diy.SimpleAudit{
		Create: adt,
		Update: adt,
	}

	err = m.IsValid()
	if err != nil {
		return nil, err
	}

	createMovieParams := datastore.CreateMovieParams{
		MovieID:         m.ID,
		ExtlID:          m.ExternalID.String(),
		Title:           m.Title,
		Rated:           diy.NewNullString(m.Rated),
		Released:        diy.NewNullTime(released),
		RunTime:         diy.NewNullInt32(int32(m.RunTime)),
		Director:        diy.NewNullString(m.Director),
		Writer:          diy.NewNullString(m.Writer),
		CreateAppID:     sa.Create.App.ID,
		CreateUserID:    sa.Create.User.NullUUID(),
		CreateTimestamp: sa.Create.Moment,
		UpdateAppID:     sa.Update.App.ID,
		UpdateUserID:    sa.Update.User.NullUUID(),
		UpdateTimestamp: sa.Update.Moment,
	}

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	_, err = datastore.New(tx).CreateMovie(ctx, createMovieParams)
	if err != nil {
		return nil, errs.E(errs.Database, err)
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return nil, err
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
	Datastorer diy.Datastorer
}

// Update is used to update a movie
func (s UpdateMovieService) Update(ctx context.Context, r *UpdateMovieRequest, adt diy.Audit) (mr *MovieResponse, err error) {

	var released time.Time
	released, err = time.Parse(time.RFC3339, r.Released)
	if err != nil {
		return nil, errs.E(errs.Validation,
			errs.Code("invalid_date_format"),
			errs.Parameter("release_date"),
			err)
	}

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	// retrieve existing Movie
	var row datastore.FindMovieByExternalIDWithAuditRow
	row, err = datastore.New(tx).FindMovieByExternalIDWithAudit(ctx, r.ExternalID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.E(errs.Validation, "No movie exists for the given external ID")
		}
		return nil, errs.E(errs.Database, err)
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
		return nil, err
	}

	sa := diy.SimpleAudit{
		Create: diy.Audit{
			App: &diy.App{
				ID:          row.CreateAppID,
				ExternalID:  secure.MustParseIdentifier(row.CreateAppExtlID),
				Org:         &diy.Org{ID: row.CreateAppOrgID},
				Name:        row.CreateAppName,
				Description: row.CreateAppDescription,
				APIKeys:     nil,
			},
			User: &diy.User{
				ID:        row.CreateUserID.UUID,
				FirstName: row.CreateUserFirstName.String,
				LastName:  row.CreateUserLastName.String,
			},
			Moment: row.CreateTimestamp,
		},
	}
	// update audit with latest
	sa.Update = adt

	updateMovieParams := datastore.UpdateMovieParams{
		Title:           m.Title,
		Rated:           diy.NewNullString(m.Rated),
		Released:        diy.NewNullTime(released),
		RunTime:         diy.NewNullInt32(int32(m.RunTime)),
		Director:        diy.NewNullString(m.Director),
		Writer:          diy.NewNullString(m.Writer),
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    adt.User.NullUUID(),
		UpdateTimestamp: adt.Moment,
		MovieID:         m.ID,
	}

	err = datastore.New(tx).UpdateMovie(ctx, updateMovieParams)
	if err != nil {
		return nil, errs.E(errs.Database, err)
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return nil, err
	}

	mr = newMovieResponse(movieAudit{m, sa})

	return mr, nil
}

// DeleteMovieService is a service for deleting a Movie
type DeleteMovieService struct {
	Datastorer diy.Datastorer
}

// Delete is used to delete a movie
func (s DeleteMovieService) Delete(ctx context.Context, extlID string) (dr diy.DeleteResponse, err error) {

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return diy.DeleteResponse{}, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	// retrieve existing Movie
	var dbm datastore.Movie
	dbm, err = datastore.New(tx).FindMovieByExternalID(ctx, extlID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return diy.DeleteResponse{}, errs.E(errs.Validation, "No movie exists for the given external ID")
		}
		return diy.DeleteResponse{}, errs.E(errs.Database, err)
	}

	err = datastore.New(tx).DeleteMovie(ctx, dbm.MovieID)
	if err != nil {
		return diy.DeleteResponse{}, errs.E(errs.Database, err)
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return diy.DeleteResponse{}, err
	}

	response := diy.DeleteResponse{
		ExternalID: dbm.ExtlID,
		Deleted:    true,
	}

	return response, nil
}

// FindMovieService is a service for reading Movies from the DB
type FindMovieService struct {
	Datastorer diy.Datastorer
}

// FindMovieByID is used to find an individual movie
func (s FindMovieService) FindMovieByID(ctx context.Context, extlID string) (mr *MovieResponse, err error) {

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	var row datastore.FindMovieByExternalIDWithAuditRow
	row, err = datastore.New(tx).FindMovieByExternalIDWithAudit(ctx, extlID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.E(errs.Validation, "no movie exists for the given external ID")
		}
		return nil, errs.E(errs.Database, err)
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

	sa := diy.SimpleAudit{
		Create: diy.Audit{
			App: &diy.App{
				ID:          row.CreateAppID,
				ExternalID:  secure.MustParseIdentifier(row.CreateAppExtlID),
				Org:         &diy.Org{ID: row.CreateAppOrgID},
				Name:        row.CreateAppName,
				Description: row.CreateAppDescription,
				APIKeys:     nil,
			},
			User: &diy.User{
				ID:        row.CreateUserID.UUID,
				FirstName: row.CreateUserFirstName.String,
				LastName:  row.CreateUserLastName.String,
			},
			Moment: row.CreateTimestamp,
		},
		Update: diy.Audit{
			App: &diy.App{
				ID:          row.UpdateAppID,
				ExternalID:  secure.MustParseIdentifier(row.UpdateAppExtlID),
				Org:         &diy.Org{ID: row.UpdateAppOrgID},
				Name:        row.UpdateAppName,
				Description: row.UpdateAppDescription,
				APIKeys:     nil,
			},
			User: &diy.User{
				ID:        row.UpdateUserID.UUID,
				FirstName: row.UpdateUserFirstName.String,
				LastName:  row.UpdateUserLastName.String,
			},
			Moment: row.UpdateTimestamp,
		},
	}

	mr = newMovieResponse(movieAudit{m, sa})

	return mr, nil
}

// FindAllMovies is used to list all movies in the db
func (s FindMovieService) FindAllMovies(ctx context.Context) (smr []*MovieResponse, err error) {

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	var rows []datastore.FindMoviesRow
	rows, err = datastore.New(tx).FindMovies(ctx)
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
		sa := diy.SimpleAudit{
			Create: diy.Audit{
				App: &diy.App{
					ID:          row.CreateAppID,
					ExternalID:  secure.MustParseIdentifier(row.CreateAppExtlID),
					Org:         &diy.Org{ID: row.CreateAppOrgID},
					Name:        row.CreateAppName,
					Description: row.CreateAppDescription,
					APIKeys:     nil,
				},
				User: &diy.User{
					ID:        row.CreateUserID.UUID,
					FirstName: row.CreateUserFirstName.String,
					LastName:  row.CreateUserLastName.String,
				},
				Moment: row.CreateTimestamp,
			},
			Update: diy.Audit{
				App: &diy.App{
					ID:          row.UpdateAppID,
					ExternalID:  secure.MustParseIdentifier(row.UpdateAppExtlID),
					Org:         &diy.Org{ID: row.UpdateAppOrgID},
					Name:        row.UpdateAppName,
					Description: row.UpdateAppDescription,
					APIKeys:     nil,
				},
				User: &diy.User{
					ID:        row.UpdateUserID.UUID,
					FirstName: row.UpdateUserLastName.String,
					LastName:  row.UpdateUserLastName.String,
				},
				Moment: row.UpdateTimestamp,
			},
		}
		mr := newMovieResponse(movieAudit{m, sa})
		smr = append(smr, mr)
	}

	return smr, nil
}
