package service_test

import (
	"context"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jackc/pgx/v4"

	"github.com/gilcrest/saaswhip"
	"github.com/gilcrest/saaswhip/errs"
	"github.com/gilcrest/saaswhip/service"
	"github.com/gilcrest/saaswhip/sqldb/datastore"
	"github.com/gilcrest/saaswhip/sqldb/sqldbtest"
)

func TestMovieService(t *testing.T) {
	t.Run("create movie nil request error", func(t *testing.T) {
		c := qt.New(t)

		var err error
		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		// start db txn using pgxpool
		ctx := context.Background()
		var tx pgx.Tx
		tx, err = db.BeginTx(ctx)
		if err != nil {
			t.Fatalf("db.BeginTx error: %v", err)
		}
		// defer transaction rollback and handle error, if any
		defer func() {
			err = db.RollbackTx(ctx, tx, err)
		}()

		s := service.MovieService{Datastorer: db}
		adt := findTestAudit(ctx, c, tx)

		var got *saaswhip.MovieResponse
		got, err = s.Create(context.Background(), nil, adt)
		c.Assert(errs.KindIs(errs.Validation, err), qt.IsTrue)
		c.Assert(err.Error(), qt.Equals, "CreateMovieRequest must have a value when creating a Movie")
		c.Assert(got, qt.IsNil)
	})
	t.Run("create movie", func(t *testing.T) {
		c := qt.New(t)

		var err error

		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		// start db txn using pgxpool
		ctx := context.Background()
		var tx pgx.Tx
		tx, err = db.BeginTx(ctx)
		if err != nil {
			t.Fatalf("db.BeginTx error: %v", err)
		}
		// defer transaction rollback and handle error, if any
		defer func() {
			err = db.RollbackTx(ctx, tx, err)
		}()

		s := service.MovieService{Datastorer: db}

		rd, _ := time.Parse(time.RFC3339, "1985-08-16T00:00:00Z")
		r := saaswhip.CreateMovieRequest{
			Title:    "The Return of the Living Dead",
			Rated:    "R",
			Released: rd.Format(time.RFC3339),
			RunTime:  91,
			Director: "Dan O'Bannon",
			Writer:   "Russell Streiner",
		}

		adt := findPrincipalTestAudit(ctx, c, tx)

		var got *saaswhip.MovieResponse
		got, err = s.Create(context.Background(), &r, adt)
		c.Assert(err, qt.IsNil)

		want := &saaswhip.MovieResponse{
			ExternalID:          got.ExternalID,
			Title:               "The Return of the Living Dead",
			Rated:               "R",
			Released:            rd.Format(time.RFC3339),
			RunTime:             91,
			Director:            "Dan O'Bannon",
			Writer:              "Russell Streiner",
			CreateAppExtlID:     adt.App.ExternalID.String(),
			CreateUserFirstName: adt.User.FirstName,
			CreateUserLastName:  adt.User.LastName,
			UpdateAppExtlID:     adt.App.ExternalID.String(),
			UpdateUserFirstName: adt.User.FirstName,
			UpdateUserLastName:  adt.User.LastName,
		}
		ignoreFields := []string{"ExternalID", "CreateDateTime", "UpdateDateTime"}
		c.Assert(got, qt.CmpEquals(cmpopts.IgnoreFields(saaswhip.MovieResponse{}, ignoreFields...)), want)
	})
	t.Run("find Movie By External ID", func(t *testing.T) {
		c := qt.New(t)

		var err error

		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		// start db txn using pgxpool
		ctx := context.Background()
		var tx pgx.Tx
		tx, err = db.BeginTx(ctx)
		if err != nil {
			t.Fatalf("db.BeginTx error: %v", err)
		}
		// defer transaction rollback and handle error, if any
		defer func() {
			err = db.RollbackTx(ctx, tx, err)
		}()

		var movies []datastore.Movie
		movies, err = datastore.New(tx).FindMoviesByTitle(ctx, "The Return of the Living Dead")
		if err != nil {
			t.Fatalf("FindMoviesByTitle() error = %v", err)
		}

		// grab first movie that matches from list
		dbm := movies[0]

		s := service.MovieService{Datastorer: db}

		var got *saaswhip.MovieResponse
		got, err = s.FindMovieByExternalID(context.Background(), dbm.ExtlID)
		want := "The Return of the Living Dead"
		c.Assert(err, qt.IsNil)
		c.Assert(got.Title, qt.Equals, want)
	})
	t.Run("Find All Movies", func(t *testing.T) {
		c := qt.New(t)

		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		ctx := context.Background()

		s := service.MovieService{
			Datastorer: db,
		}

		var (
			got []*saaswhip.MovieResponse
			err error
		)
		got, err = s.FindAllMovies(ctx)
		c.Assert(err, qt.IsNil)
		c.Assert(len(got) >= 1, qt.IsTrue, qt.Commentf("movies found = %d", len(got)))
		c.Logf("movies found = %d", len(got))
	})
	t.Run("update movie", func(t *testing.T) {
		c := qt.New(t)

		var err error

		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		// start db txn using pgxpool
		ctx := context.Background()
		var tx pgx.Tx
		tx, err = db.BeginTx(ctx)
		if err != nil {
			t.Fatalf("db.BeginTx error: %v", err)
		}
		// defer transaction rollback and handle error, if any
		defer func() {
			err = db.RollbackTx(ctx, tx, err)
		}()

		var movies []datastore.Movie
		movies, err = datastore.New(tx).FindMoviesByTitle(ctx, "The Return of the Living Dead")
		if err != nil {
			t.Fatalf("FindMoviesByTitle() error = %v", err)
		}

		// grab first movie that matches from list
		dbm := movies[0]

		s := service.MovieService{Datastorer: db}

		rd, _ := time.Parse(time.RFC3339, "1985-08-16T00:00:00Z")
		r := saaswhip.UpdateMovieRequest{
			ExternalID: dbm.ExtlID,
			Title:      "The Return of the Living Thread",
			Rated:      "R",
			Released:   rd.Format(time.RFC3339),
			RunTime:    91,
			Director:   "Dan O'Bannon",
			Writer:     "Russell Streiner",
		}

		adt := findPrincipalTestAudit(ctx, c, tx)

		var got *saaswhip.MovieResponse
		got, err = s.Update(context.Background(), &r, adt)
		c.Assert(err, qt.IsNil)

		want := &saaswhip.MovieResponse{
			ExternalID:          got.ExternalID,
			Title:               "The Return of the Living Thread",
			Rated:               "R",
			Released:            rd.Format(time.RFC3339),
			RunTime:             91,
			Director:            "Dan O'Bannon",
			Writer:              "Russell Streiner",
			CreateAppExtlID:     adt.App.ExternalID.String(),
			CreateUserFirstName: adt.User.FirstName,
			CreateUserLastName:  adt.User.LastName,
			UpdateAppExtlID:     adt.App.ExternalID.String(),
			UpdateUserFirstName: adt.User.FirstName,
			UpdateUserLastName:  adt.User.LastName,
		}
		ignoreFields := []string{"ExternalID", "CreateDateTime", "UpdateDateTime"}
		c.Assert(got, qt.CmpEquals(cmpopts.IgnoreFields(saaswhip.MovieResponse{}, ignoreFields...)), want)
	})
	t.Run("delete movie", func(t *testing.T) {
		c := qt.New(t)

		var err error

		db, cleanup := sqldbtest.NewDB(t)
		c.Cleanup(cleanup)

		// start db txn using pgxpool
		ctx := context.Background()
		var tx pgx.Tx
		tx, err = db.BeginTx(ctx)
		if err != nil {
			t.Fatalf("db.BeginTx error: %v", err)
		}
		// defer transaction rollback and handle error, if any
		defer func() {
			err = db.RollbackTx(ctx, tx, err)
		}()

		var movies []datastore.Movie
		movies, err = datastore.New(tx).FindMoviesByTitle(ctx, "The Return of the Living Thread")
		if err != nil {
			t.Fatalf("FindMoviesByTitle() error = %v", err)
		}

		// grab first movie that matches from list
		dbm := movies[0]

		s := service.MovieService{
			Datastorer: db,
		}

		var got saaswhip.DeleteResponse
		got, err = s.Delete(context.Background(), dbm.ExtlID)
		want := saaswhip.DeleteResponse{
			ExternalID: dbm.ExtlID,
			Deleted:    true,
		}
		c.Assert(err, qt.IsNil)
		c.Assert(got, qt.CmpEquals(), want)
	})
}
