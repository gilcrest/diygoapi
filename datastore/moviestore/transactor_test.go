package moviestore

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/gilcrest/go-api-basic/datastore/datastoretest"
	"github.com/gilcrest/go-api-basic/domain/movie"
)

func TestNewTransactor(t *testing.T) {
	type args struct {
		ds Datastorer
	}

	datastore, cleanup := datastoretest.NewDatastore(t)
	t.Cleanup(cleanup)
	transactor := Transactor{datastore}

	tests := []struct {
		name string
		args args
		want Transactor
	}{
		{"typical", args{ds: datastore}, transactor},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTransactor(tt.args.ds); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTransactor() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTransactor_Create(t *testing.T) {
	type fields struct {
		datastorer Datastorer
	}
	type args struct {
		ctx context.Context
		m   *movie.Movie
	}

	datastore, cleanup := datastoretest.NewDatastore(t)
	transactor := NewTransactor(datastore)
	ctx := context.Background()
	m := newMovie(t)
	t.Cleanup(func() {
		err := transactor.Delete(ctx, m)
		if err != nil {
			t.Fatalf("transactor.Delete error = %v", err)
		}
		cleanup()
	})

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"typical", fields{datastorer: datastore}, args{ctx, m}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := Transactor{
				datastorer: tt.fields.datastorer,
			}
			if err := dt.Create(tt.args.ctx, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("transactor.Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransactor_Update(t *testing.T) {
	type fields struct {
		datastorer Datastorer
	}
	type args struct {
		ctx context.Context
		m   *movie.Movie
	}

	// I am intentionally not using the cleanup function that is
	// returned as I need the DB to stay open for the test
	// t.Cleanup function
	datastore, _ := datastoretest.NewDatastore(t)

	ctx := context.Background()
	// create a movie with the helper to ensure that at least one row
	// is returned
	m, mCleanup := NewMovieDBHelper(ctx, t, datastore)
	t.Cleanup(mCleanup)
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
		{"typical", fields{datastorer: datastore}, args{ctx, m}, false},
		{"no rows updated", fields{datastorer: datastore}, args{ctx, m2}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := Transactor{
				datastorer: tt.fields.datastorer,
			}
			if err := dt.Update(tt.args.ctx, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("transactor.Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTransactor_Delete(t *testing.T) {
	type fields struct {
		datastorer Datastorer
	}
	type args struct {
		ctx context.Context
		m   *movie.Movie
	}

	// I am intentionally not using the cleanup function that is
	// returned as I need the DB to stay open for the test
	// t.Cleanup function
	datastore, _ := datastoretest.NewDatastore(t)

	ctx := context.Background()
	// create a movie with the helper to ensure that at least one row
	// is returned
	m, _ := NewMovieDBHelper(ctx, t, datastore)

	m2 := &movie.Movie{}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"typical", fields{datastorer: datastore}, args{ctx, m}, false},
		{"no rows deleted", fields{datastorer: datastore}, args{ctx, m2}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dt := Transactor{
				datastorer: tt.fields.datastorer,
			}
			if err := dt.Delete(tt.args.ctx, tt.args.m); (err != nil) != tt.wantErr {
				t.Logf("%s yieled dt.Delete error = %v", tt.name, err)
				t.Errorf("transactor.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
