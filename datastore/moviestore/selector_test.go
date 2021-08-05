package moviestore

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp/cmpopts"

	qt "github.com/frankban/quicktest"

	"github.com/gilcrest/go-api-basic/domain/movie"

	"github.com/gilcrest/go-api-basic/datastore/datastoretest"
)

func TestNewSelector(t *testing.T) {
	type args struct {
		ds Datastorer
	}

	datastore, cleanup := datastoretest.NewDatastore(t)
	t.Cleanup(cleanup)

	selector := Selector{datastore}

	tests := []struct {
		name string
		args args
		want Selector
	}{
		{"typical", args{ds: datastore}, selector},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewSelector(tt.args.ds)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSelector() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSelector_FindAll(t *testing.T) {
	type fields struct {
		Datastorer Datastorer
	}
	type args struct {
		ctx context.Context
	}

	// I am intentionally not using the cleanup function that is
	// returned as I need the DB to stay open for the test
	// t.Cleanup function
	ds, _ := datastoretest.NewDatastore(t)
	ctx := context.Background()

	// create a movie with the helper to ensure that at least one row
	// is returned
	_, movieCleanup := NewMovieDBHelper(ctx, t, ds)
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
			d := &Selector{
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

func TestSelector_FindByID(t *testing.T) {
	c := qt.New(t)

	type fields struct {
		Datastorer Datastorer
	}
	type args struct {
		ctx    context.Context
		extlID string
	}

	// I am intentionally not using the cleanup function that is
	// returned as I need the DB to stay open for the test
	// t.Cleanup function
	ds, _ := datastoretest.NewDatastore(t)
	ctx := context.Background()

	m, _ := NewMovieDBHelper(ctx, t, ds)

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
			d := &Selector{
				Datastorer: tt.fields.Datastorer,
			}
			got, err := d.FindByID(tt.args.ctx, tt.args.extlID)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindMovieByID() error = %v, wantErr %v", err, tt.wantErr)
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
