package moviestore

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/gilcrest/go-api-basic/domain/user"
	"github.com/matryer/is"

	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/gilcrest/go-api-basic/domain/random"
	"github.com/google/uuid"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/datastore/datastoretest"
	"github.com/gilcrest/go-api-basic/domain/logger"
)

func TestNewDefaultSelector(t *testing.T) {
	type args struct {
		ds datastore.Datastorer
	}

	dsn := datastoretest.NewPGDatasourceName(t)
	lgr := logger.NewLogger(os.Stdout, true)

	db, cleanup, err := datastore.NewDB(dsn, lgr)
	defer cleanup()
	if err != nil {
		t.Errorf("datastore.NewDB error = %v", err)
	}
	defaultDatastore := datastore.NewDefaultDatastore(db)
	defaultSelector := DefaultSelector{defaultDatastore}

	tests := []struct {
		name string
		args args
		want DefaultSelector
	}{
		{"default datastore", args{ds: defaultDatastore}, defaultSelector},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDefaultSelector(tt.args.ds)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDefaultSelector() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultSelector_FindAll(t *testing.T) {
	type fields struct {
		Datastorer datastore.Datastorer
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
	ds := datastore.NewDefaultDatastore(db)
	ctx := context.Background()

	// create a movie with the helper to ensure that at least one row
	// is returned
	_ = newMovieDBHelper(ctx, t, ds, true)

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"standard test", fields{Datastorer: ds}, args{ctx: ctx}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DefaultSelector{
				Datastorer: tt.fields.Datastorer,
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

func TestDefaultSelector_FindByID(t *testing.T) {
	is := is.NewRelaxed(t)

	type fields struct {
		Datastorer datastore.Datastorer
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
	ds := datastore.NewDefaultDatastore(db)
	ctx := context.Background()

	m := newMovieDBHelper(ctx, t, ds, true)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *movie.Movie
		wantErr bool
	}{
		{"happy path", fields{Datastorer: ds}, args{ctx, m.ExternalID}, m, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DefaultSelector{
				Datastorer: tt.fields.Datastorer,
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

// newMovieDBHelper creates/inserts a new movie in the db and then
// registers a t.Cleanup function to delete it. The insert and
// delete are both in separate database transactions
func newMovieDBHelper(ctx context.Context, t *testing.T, ds datastore.Datastorer, cleanup bool) *movie.Movie {
	t.Helper()

	m := newMovie(t)

	defaultTransactor := NewDefaultTransactor(ds)

	err := defaultTransactor.Create(ctx, m)
	if err != nil {
		t.Fatalf("defaultTransactor.Create error = %v", err)
	}

	if cleanup == true {
		t.Cleanup(func() {
			err := defaultTransactor.Delete(ctx, m)
			if err != nil {
				t.Fatalf("t.Cleanup defaultTransactor.Delete error = %v", err)
			}
		})
	}

	return m
}

func newMovie(t *testing.T) *movie.Movie {
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
	return m
}
