package moviestore

import (
	"context"
	"database/sql"
	"os"
	"reflect"
	"testing"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/datastore/datastoretest"
	"github.com/gilcrest/go-api-basic/domain/logger"
	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/gilcrest/go-api-basic/domain/random"
	"github.com/gilcrest/go-api-basic/domain/user"
	"github.com/google/uuid"
	"github.com/matryer/is"
)

func TestNewDB(t *testing.T) {
	type args struct {
		db *sql.DB
	}

	dsn := datastoretest.NewPGDatasourceName(t)
	lgr := logger.NewLogger(os.Stdout, true)

	db, cleanup, err := datastore.NewDB(dsn, lgr)
	defer cleanup()
	if err != nil {
		t.Errorf("datastore.NewDB error = %v", err)
	}
	moviestoreDB := &DB{db}

	tests := []struct {
		name    string
		args    args
		want    *DB
		wantErr bool
	}{
		{"postgresql db", args{db: db}, moviestoreDB, false},
		{"nil db", args{db: nil}, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDB(tt.args.db)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDB() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDB_FindAll(t *testing.T) {
	type fields struct {
		DB *sql.DB
	}
	type args struct {
		ctx context.Context
	}

	dsn := datastoretest.NewPGDatasourceName(t)
	lgr := logger.NewLogger(os.Stdout, true)

	// I am intentionally not using the cleanup function that is
	// returned from NewDB as I need the DB to stay open for the test
	// t.Cleanup function
	db, _, err := datastore.NewDB(dsn, lgr)
	if err != nil {
		t.Errorf("datastore.NewDB error = %v", err)
	}
	ctx := context.Background()

	// create a movie with the helper to ensure that at least one row
	// is returned
	_ = newMovieHelper(t, ctx, db)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"standard test", fields{DB: db}, args{ctx: ctx}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DB{
				DB: tt.fields.DB,
			}
			got, err := d.FindAll(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("FindAll() returned %d records", len(got))
		})
	}
}

func TestDB_FindByID(t *testing.T) {
	is := is.NewRelaxed(t)

	type fields struct {
		DB *sql.DB
	}
	type args struct {
		ctx    context.Context
		extlID string
	}

	dsn := datastoretest.NewPGDatasourceName(t)
	lgr := logger.NewLogger(os.Stdout, true)

	// I am intentionally not using the cleanup function that is
	// returned from NewDB as I need the DB to stay open for the test
	// t.Cleanup function
	db, _, err := datastore.NewDB(dsn, lgr)
	if err != nil {
		t.Errorf("datastore.NewDB error = %v", err)
	}
	ctx := context.Background()

	m := newMovieHelper(t, ctx, db)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *movie.Movie
		wantErr bool
	}{
		{"happy path", fields{db}, args{ctx, m.ExternalID}, m, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DB{
				DB: tt.fields.DB,
			}
			got, err := d.FindByID(tt.args.ctx, tt.args.extlID)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			is.Equal(got.ID, tt.want.ID)
			if got.CreateTime.IsZero() == true {
				t.Error("CreateTime is zero, it should have a value")
			}
			if got.UpdateTime.IsZero() == true {
				t.Error("UpdateTime is zero, it should have a value")
			}
		})
	}
}

// newMovieHelper creates/inserts a new movie in the db and then
// registers a t.Cleanup function to delete it. The insert and
// delete are both in separate database transactions
func newMovieHelper(t *testing.T, ctx context.Context, db *sql.DB) *movie.Movie {
	t.Helper()

	id := uuid.New()
	extlID, err := random.CryptoString(15)
	if err != nil {
		t.Fatalf("random.CryptoString() error = %v", err)
	}
	u := &user.User{
		Email:     "gilcrest@gmail.com",
		FirstName: "Dan",
		LastName:  "Gillis",
	}

	m, err := movie.NewMovie(id, extlID, u)
	if err != nil {
		t.Fatalf("movie.NewMovie() error = %v", err)
	}

	sqltx1, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("db.BeginTx error = %v", err)
	}

	tx := Tx{sqltx1}
	err = tx.Create(ctx, m)
	if err != nil {
		t.Fatalf("tx.Create() error = %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("tx.Commit() error = %v", err)
	}

	t.Cleanup(func() {
		sqltx2, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("db.BeginTx error = %v", err)
		}

		tx2 := Tx{sqltx2}
		if err := tx2.Delete(ctx, m); err != nil {
			t.Fatalf("tx.Delete error = %v", err)
		}
		if err := tx2.Commit(); err != nil {
			t.Fatalf("tx.Commit() error = %v", err)
		}

	})

	return m
}
