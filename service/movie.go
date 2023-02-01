package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v4"

	"github.com/gilcrest/diygoapi"
	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/secure"
	"github.com/gilcrest/diygoapi/sqldb/datastore"
)

// movieAudit is the combination of a domain Movie and its audit data
type movieAudit struct {
	Movie       diygoapi.Movie
	SimpleAudit diygoapi.SimpleAudit
}

// newMovieResponse initializes MovieResponse
func newMovieResponse(ma movieAudit) *diygoapi.MovieResponse {
	return &diygoapi.MovieResponse{
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

// MovieService is a service for creating a Movie
type MovieService struct {
	Datastorer diygoapi.Datastorer
}

// Create is used to create a Movie
func (s *MovieService) Create(ctx context.Context, r *diygoapi.CreateMovieRequest, adt diygoapi.Audit) (mr *diygoapi.MovieResponse, err error) {
	const op errs.Op = "service/MovieService.Create"

	if r == nil {
		return nil, errs.E(op, errs.Validation, "CreateMovieRequest must have a value when creating a Movie")
	}

	var released time.Time
	released, err = time.Parse(time.RFC3339, r.Released)
	if err != nil {
		return nil, errs.E(op, errs.Validation,
			errs.Code("invalid_date_format"),
			errs.Parameter("release_date"),
			err)
	}

	// initialize Movie and inject dependent fields
	m := diygoapi.Movie{
		ID:         uuid.New(),
		ExternalID: secure.NewID(),
		Title:      r.Title,
		Rated:      r.Rated,
		Released:   released,
		RunTime:    r.RunTime,
		Director:   r.Director,
		Writer:     r.Writer,
	}

	sa := diygoapi.SimpleAudit{
		Create: adt,
		Update: adt,
	}

	err = m.IsValid()
	if err != nil {
		return nil, errs.E(op, err)
	}

	createMovieParams := datastore.CreateMovieParams{
		MovieID:         m.ID,
		ExtlID:          m.ExternalID.String(),
		Title:           m.Title,
		Rated:           diygoapi.NewNullString(m.Rated),
		Released:        diygoapi.NewNullTime(released),
		RunTime:         diygoapi.NewNullInt32(int32(m.RunTime)),
		Director:        diygoapi.NewNullString(m.Director),
		Writer:          diygoapi.NewNullString(m.Writer),
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
		return nil, errs.E(op, err)
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	_, err = datastore.New(tx).CreateMovie(ctx, createMovieParams)
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	mr = newMovieResponse(movieAudit{m, sa})

	return mr, nil
}

// Update is used to update a movie
func (s *MovieService) Update(ctx context.Context, r *diygoapi.UpdateMovieRequest, adt diygoapi.Audit) (mr *diygoapi.MovieResponse, err error) {
	const op errs.Op = "service/MovieService.Update"

	var released time.Time
	released, err = time.Parse(time.RFC3339, r.Released)
	if err != nil {
		return nil, errs.E(op, errs.Validation,
			errs.Code("invalid_date_format"),
			errs.Parameter("release_date"),
			err)
	}

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, errs.E(op, err)
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
			return nil, errs.E(op, errs.Validation, "No movie exists for the given external ID")
		}
		return nil, errs.E(op, errs.Database, err)
	}

	m := diygoapi.Movie{
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
		return nil, errs.E(op, err)
	}

	sa := diygoapi.SimpleAudit{
		Create: diygoapi.Audit{
			App: &diygoapi.App{
				ID:          row.CreateAppID,
				ExternalID:  secure.MustParseIdentifier(row.CreateAppExtlID),
				Org:         &diygoapi.Org{ID: row.CreateAppOrgID},
				Name:        row.CreateAppName,
				Description: row.CreateAppDescription,
				APIKeys:     nil,
			},
			User: &diygoapi.User{
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
		Rated:           diygoapi.NewNullString(m.Rated),
		Released:        diygoapi.NewNullTime(released),
		RunTime:         diygoapi.NewNullInt32(int32(m.RunTime)),
		Director:        diygoapi.NewNullString(m.Director),
		Writer:          diygoapi.NewNullString(m.Writer),
		UpdateAppID:     adt.App.ID,
		UpdateUserID:    adt.User.NullUUID(),
		UpdateTimestamp: adt.Moment,
		MovieID:         m.ID,
	}

	err = datastore.New(tx).UpdateMovie(ctx, updateMovieParams)
	if err != nil {
		return nil, errs.E(op, errs.Database, err)
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return nil, errs.E(op, err)
	}

	mr = newMovieResponse(movieAudit{m, sa})

	return mr, nil
}

// Delete is used to delete a movie
func (s *MovieService) Delete(ctx context.Context, extlID string) (dr diygoapi.DeleteResponse, err error) {
	const op errs.Op = "service/MovieService.Delete"

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return diygoapi.DeleteResponse{}, errs.E(op, err)
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
			return diygoapi.DeleteResponse{}, errs.E(op, errs.Validation, "No movie exists for the given external ID")
		}
		return diygoapi.DeleteResponse{}, errs.E(op, errs.Database, err)
	}

	var rowsAffected int64
	rowsAffected, err = datastore.New(tx).DeleteMovie(ctx, dbm.MovieID)
	if err != nil {
		return diygoapi.DeleteResponse{}, errs.E(op, errs.Database, err)
	}

	if rowsAffected != 1 {
		return diygoapi.DeleteResponse{}, errs.E(op, errs.Database, fmt.Sprintf("rows affected should be 1, actual: %d", rowsAffected))
	}

	// commit db txn using pgxpool
	err = s.Datastorer.CommitTx(ctx, tx)
	if err != nil {
		return diygoapi.DeleteResponse{}, errs.E(op, err)
	}

	response := diygoapi.DeleteResponse{
		ExternalID: dbm.ExtlID,
		Deleted:    true,
	}

	return response, nil
}

// FindMovieByExternalID is used to find an individual movie
func (s *MovieService) FindMovieByExternalID(ctx context.Context, extlID string) (mr *diygoapi.MovieResponse, err error) {
	const op errs.Op = "service/MovieService.FindMovieByExternalID"

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	var row datastore.FindMovieByExternalIDWithAuditRow
	row, err = datastore.New(tx).FindMovieByExternalIDWithAudit(ctx, extlID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.E(op, errs.Validation, "no movie exists for the given external ID")
		}
		return nil, errs.E(op, errs.Database, err)
	}

	m := diygoapi.Movie{
		ID:         row.MovieID,
		ExternalID: secure.MustParseIdentifier(row.ExtlID),
		Title:      row.Title,
		Rated:      row.Rated.String,
		Released:   row.Released.Time,
		RunTime:    int(row.RunTime.Int32),
		Director:   row.Director.String,
		Writer:     row.Writer.String,
	}

	sa := diygoapi.SimpleAudit{
		Create: diygoapi.Audit{
			App: &diygoapi.App{
				ID:          row.CreateAppID,
				ExternalID:  secure.MustParseIdentifier(row.CreateAppExtlID),
				Org:         &diygoapi.Org{ID: row.CreateAppOrgID},
				Name:        row.CreateAppName,
				Description: row.CreateAppDescription,
				APIKeys:     nil,
			},
			User: &diygoapi.User{
				ID:        row.CreateUserID.UUID,
				FirstName: row.CreateUserFirstName.String,
				LastName:  row.CreateUserLastName.String,
			},
			Moment: row.CreateTimestamp,
		},
		Update: diygoapi.Audit{
			App: &diygoapi.App{
				ID:          row.UpdateAppID,
				ExternalID:  secure.MustParseIdentifier(row.UpdateAppExtlID),
				Org:         &diygoapi.Org{ID: row.UpdateAppOrgID},
				Name:        row.UpdateAppName,
				Description: row.UpdateAppDescription,
				APIKeys:     nil,
			},
			User: &diygoapi.User{
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
func (s *MovieService) FindAllMovies(ctx context.Context) (smr []*diygoapi.MovieResponse, err error) {
	const op errs.Op = "service/MovieService.FindAllMovies"

	// start db txn using pgxpool
	var tx pgx.Tx
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, errs.E(op, err)
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	var rows []datastore.FindMoviesRow
	rows, err = datastore.New(tx).FindMovies(ctx)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errs.E(op, errs.Validation, "no movies exists")
		}
		return nil, errs.E(op, errs.Database, err)
	}

	for _, row := range rows {
		m := diygoapi.Movie{
			ID:         row.MovieID,
			ExternalID: secure.MustParseIdentifier(row.ExtlID),
			Title:      row.Title,
			Rated:      row.Rated.String,
			Released:   row.Released.Time,
			RunTime:    int(row.RunTime.Int32),
			Director:   row.Director.String,
			Writer:     row.Writer.String,
		}
		sa := diygoapi.SimpleAudit{
			Create: diygoapi.Audit{
				App: &diygoapi.App{
					ID:          row.CreateAppID,
					ExternalID:  secure.MustParseIdentifier(row.CreateAppExtlID),
					Org:         &diygoapi.Org{ID: row.CreateAppOrgID},
					Name:        row.CreateAppName,
					Description: row.CreateAppDescription,
					APIKeys:     nil,
				},
				User: &diygoapi.User{
					ID:        row.CreateUserID.UUID,
					FirstName: row.CreateUserFirstName.String,
					LastName:  row.CreateUserLastName.String,
				},
				Moment: row.CreateTimestamp,
			},
			Update: diygoapi.Audit{
				App: &diygoapi.App{
					ID:          row.UpdateAppID,
					ExternalID:  secure.MustParseIdentifier(row.UpdateAppExtlID),
					Org:         &diygoapi.Org{ID: row.UpdateAppOrgID},
					Name:        row.UpdateAppName,
					Description: row.UpdateAppDescription,
					APIKeys:     nil,
				},
				User: &diygoapi.User{
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
