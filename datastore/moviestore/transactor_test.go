package moviestore

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/datastore/datastoretest"
	"github.com/gilcrest/go-api-basic/domain/logger"
	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/google/uuid"
)

func TestNewDefaultTransactor(t *testing.T) {
	type args struct {
		ds datastore.Datastorer
	}
	lgr := logger.NewLogger(os.Stdout, true)

	db, cleanup := datastoretest.NewDB(t, lgr)
	defer cleanup()
	defaultDatastore := datastore.NewDefaultDatastore(db)
	defaultTransactor := DefaultTransactor{defaultDatastore}

	tests := []struct {
		name string
		args args
		want DefaultTransactor
	}{
		{"typical", args{ds: defaultDatastore}, defaultTransactor},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDefaultTransactor(tt.args.ds); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDefaultTransactor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultTransactor_Create(t *testing.T) {
	type fields struct {
		datastorer datastore.Datastorer
	}
	type args struct {
		ctx context.Context
		m   *movie.Movie
	}
	lgr := logger.NewLogger(os.Stdout, true)

	// I am intentionally not using the cleanup function that is
	// returned from NewDB as I need the DB to stay open for the test
	// t.Cleanup function
	db, _ := datastoretest.NewDB(t, lgr)
	defaultDatastore := datastore.NewDefaultDatastore(db)
	defaultTransactor := NewDefaultTransactor(defaultDatastore)
	ctx := context.Background()
	m := newMovie(t)
	t.Cleanup(func() {
		err := defaultTransactor.Delete(ctx, m)
		if err != nil {
			t.Fatalf("defaultTransactor.Delete error = %v", err)
		}
	})

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"typical", fields{datastorer: defaultDatastore}, args{ctx, m}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := DefaultTransactor{
				datastorer: tt.fields.datastorer,
			}
			if err := dt.Create(tt.args.ctx, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("DefaultTransactor.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultTransactor_Update(t *testing.T) {
	type fields struct {
		datastorer datastore.Datastorer
	}
	type args struct {
		ctx context.Context
		m   *movie.Movie
	}
	lgr := logger.NewLogger(os.Stdout, true)

	// I am intentionally not using the cleanup function that is
	// returned from NewDB as I need the DB to stay open for the test
	// t.Cleanup function
	db, _ := datastoretest.NewDB(t, lgr)
	defaultDatastore := datastore.NewDefaultDatastore(db)
	// defaultTransactor := NewDefaultTransactor(defaultDatastore)
	ctx := context.Background()
	// create a movie with the helper to ensure that at least one row
	// is returned
	m := newMovieDBHelper(ctx, t, defaultDatastore, true)
	// The ID would not be set on an update, as only the external ID
	// is known to the client
	m.ID = uuid.Nil
	m.SetDirector("Alex Cox")

	m2 := &movie.Movie{}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"typical", fields{datastorer: defaultDatastore}, args{ctx, m}, false},
		{"no rows updated", fields{datastorer: defaultDatastore}, args{ctx, m2}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := DefaultTransactor{
				datastorer: tt.fields.datastorer,
			}
			if err := dt.Update(tt.args.ctx, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("DefaultTransactor.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDefaultTransactor_Delete(t *testing.T) {
	type fields struct {
		datastorer datastore.Datastorer
	}
	type args struct {
		ctx context.Context
		m   *movie.Movie
	}
	lgr := logger.NewLogger(os.Stdout, true)

	// I am intentionally not using the cleanup function that is
	// returned from NewDB as I need the DB to stay open for the test
	// t.Cleanup function
	db, _ := datastoretest.NewDB(t, lgr)
	defaultDatastore := datastore.NewDefaultDatastore(db)
	// defaultTransactor := NewDefaultTransactor(defaultDatastore)
	ctx := context.Background()
	// create a movie with the helper to ensure that at least one row
	// is returned
	m := newMovieDBHelper(ctx, t, defaultDatastore, false)

	m2 := &movie.Movie{}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"typical", fields{datastorer: defaultDatastore}, args{ctx, m}, false},
		{"no rows deleted", fields{datastorer: defaultDatastore}, args{ctx, m2}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := DefaultTransactor{
				datastorer: tt.fields.datastorer,
			}
			if err := dt.Delete(tt.args.ctx, tt.args.m); (err != nil) != tt.wantErr {
				t.Logf("%s yieled dt.Delete error = %v", tt.name, err)
				t.Errorf("DefaultTransactor.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
