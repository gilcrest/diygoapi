package moviestore

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"

	qt "github.com/frankban/quicktest"

	"github.com/gilcrest/go-api-basic/domain/movie"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/datastore/datastoretest"
	"github.com/gilcrest/go-api-basic/domain/logger"
)

func TestNewDefaultSelector(t *testing.T) {
	type args struct {
		ds datastore.Datastorer
	}

	lgr := logger.NewLogger(os.Stdout, true)

	db, cleanup := datastoretest.NewDB(t, lgr)
	defer cleanup()

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

	lgr := logger.NewLogger(os.Stdout, true)

	// I am intentionally not using the cleanup function that is
	// returned from NewDB as I need the DB to stay open for the test
	// t.Cleanup function
	db, _ := datastoretest.NewDB(t, lgr)
	ds := datastore.NewDefaultDatastore(db)
	ctx := context.Background()

	// create a movie with the helper to ensure that at least one row
	// is returned
	_, movieCleanup := NewMovieDBHelper(t, ctx, ds)
	t.Cleanup(movieCleanup)

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
	c := qt.New(t)

	type fields struct {
		Datastorer datastore.Datastorer
	}
	type args struct {
		ctx    context.Context
		extlID string
	}

	lgr := logger.NewLogger(os.Stdout, true)

	// I am intentionally not using the cleanup function that is
	// returned from NewDB as I need the DB to stay open for the test
	// t.Cleanup function
	db, _ := datastoretest.NewDB(t, lgr)
	ds := datastore.NewDefaultDatastore(db)
	ctx := context.Background()

	m, _ := NewMovieDBHelper(t, ctx, ds)

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
			ignoreFields := cmpopts.IgnoreFields(movie.Movie{},
				"ExternalID", "CreateUser", "UpdateUser", "CreateTime", "UpdateTime")

			c.Assert(got, qt.CmpEquals(ignoreFields), tt.want)
			if got.CreateTime.IsZero() == true {
				t.Error("CreateTime is zero, it should have a value")
			}
			if got.UpdateTime.IsZero() == true {
				t.Error("UpdateTime is zero, it should have a value")
			}
		})
	}
}
