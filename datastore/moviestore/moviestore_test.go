package moviestore

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"github.com/gilcrest/go-api-basic/domain/movie"
)

func TestDB_FindAll(t *testing.T) {
	type fields struct {
		DB *sql.DB
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*movie.Movie
		wantErr bool
	}{
		// TODO: Add test cases.
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindAll() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDB_FindByID(t *testing.T) {
	type fields struct {
		DB *sql.DB
	}
	type args struct {
		ctx    context.Context
		extlID string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *movie.Movie
		wantErr bool
	}{
		// TODO: Add test cases.
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
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FindByID() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDB(t *testing.T) {
	type args struct {
		db *sql.DB
	}
	tests := []struct {
		name    string
		args    args
		want    *DB
		wantErr bool
	}{
		// TODO: Add test cases.
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

func TestNewTx(t *testing.T) {
	type args struct {
		tx *sql.Tx
	}
	tests := []struct {
		name    string
		args    args
		want    *Tx
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTx(tt.args.tx)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTx() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTx_Create(t1 *testing.T) {
	type fields struct {
		Tx *sql.Tx
	}
	type args struct {
		ctx context.Context
		m   *movie.Movie
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Tx{
				Tx: tt.fields.Tx,
			}
			if err := t.Create(tt.args.ctx, tt.args.m); (err != nil) != tt.wantErr {
				t1.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTx_Delete(t1 *testing.T) {
	type fields struct {
		Tx *sql.Tx
	}
	type args struct {
		ctx context.Context
		m   *movie.Movie
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Tx{
				Tx: tt.fields.Tx,
			}
			if err := t.Delete(tt.args.ctx, tt.args.m); (err != nil) != tt.wantErr {
				t1.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTx_Update(t1 *testing.T) {
	type fields struct {
		Tx *sql.Tx
	}
	type args struct {
		ctx context.Context
		m   *movie.Movie
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Tx{
				Tx: tt.fields.Tx,
			}
			if err := t.Update(tt.args.ctx, tt.args.m); (err != nil) != tt.wantErr {
				t1.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

//func NewMockTx() *MockTx {
//	return &MockTx{}
//}
//
//// MockMovieDB is the mock database implementation for CRUD operations for a movie
//type MockTx struct {
//	Log zerolog.Logger
//}
//
//// Create is a mock for creating a record
//func (t MockTx) Create(ctx context.Context, m *movie.Movie) error {
//	const op errs.Op = "movieDatastore/MockTx.Create"
//
//	// I would not recommend actually getting timestamps from the
//	// database on create, but I put in an example of doing it anyway
//	// Because of that, I have to set the timestamps in this mock
//	// as if they were being set by the DB procedure that is being
//	// called
//	m.CreateTime = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)
//	m.UpdateTime = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)
//
//	return nil
//}
//
//// Update is a mock for updating a record
//func (t MockTx) Update(ctx context.Context, m *movie.Movie) error {
//	const op errs.Op = "movieDatastore/MockTx.Update"
//
//	// Updates are a little different - on the non-mock, I am
//	// actually getting back data as part of the update of the
//	// original record that is helpful, so I am recreating that
//	m.ID = uuid.MustParse("b7f34380-386d-4142-b9a0-3834d6e2288e")
//	m.CreateTime = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)
//
//	return nil
//}
//
//// Delete mocks removing the Movie record from the table
//func (t MockTx) Delete(ctx context.Context, m *movie.Movie) error {
//	return nil
//}
//
//func NewMockDB() *MockDB {
//	return &MockDB{}
//}
//
//type MockDB struct {
//}
//
//// FindByID returns a Movie struct to populate the response
//func (d MockDB) FindByID(ctx context.Context, extlID string) (*movie.Movie, error) {
//	m1 := new(movie.Movie)
//	m1.ExternalID = extlID
//	m1.Title = "The Thing"
//	m1.Rated = "R"
//	m1.Released = time.Date(1982, time.June, 25, 0, 0, 0, 0, time.UTC)
//	m1.RunTime = 109
//	m1.Director = "John Carpenter"
//	m1.Writer = "Bill Lancaster"
//	m1.CreateTime = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)
//	m1.UpdateTime = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)
//
//	return m1, nil
//}
//
//// FindAll returns a slice of Movie structs to populate the response
//func (d MockDB) FindAll(ctx context.Context) ([]*movie.Movie, error) {
//	const op errs.Op = "movieDatastore/MockMovieDB.FindAll"
//
//	m1 := new(movie.Movie)
//	m1.ID = uuid.MustParse("4e58fa6a-5c4e-4e39-b6df-341087d1074b")
//	m1.ExternalID = "Z8MnDR5iw70Z-Q9OIUgH"
//	m1.Title = "The Thing"
//	m1.Rated = "R"
//	m1.Released = time.Date(1982, time.June, 25, 0, 0, 0, 0, time.UTC)
//	m1.RunTime = 109
//	m1.Director = "John Carpenter"
//	m1.Writer = "Bill Lancaster"
//	m1.CreateTime = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)
//	m1.UpdateTime = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)
//
//	m2 := new(movie.Movie)
//	m2.ID = uuid.MustParse("b7f34380-386d-4142-b9a0-3834d6e2288e")
//	m2.ExternalID = "QxKhsURZ08sBP68MufYu"
//	m2.Title = "Repo Man"
//	m2.Rated = "R"
//	m2.Released = time.Date(1984, time.March, 2, 0, 0, 0, 0, time.UTC)
//	m2.RunTime = 109
//	m2.Director = "Alex Cox"
//	m2.Writer = "Alex Cox"
//	m2.CreateTime = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)
//	m2.UpdateTime = time.Date(2020, time.February, 25, 0, 0, 0, 0, time.UTC)
//
//	s := []*movie.Movie{m1, m2}
//
//	return s, nil
//}
