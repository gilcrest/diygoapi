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
)

func TestNewTx(t *testing.T) {
	type args struct {
		tx *sql.Tx
	}
	dsn := datastoretest.NewPGDatasourceName(t)
	lgr := logger.NewLogger(os.Stdout, true)

	db, cleanup, err := datastore.NewDB(dsn, lgr)
	defer cleanup()
	if err != nil {
		t.Errorf("datastore.NewDB error = %v", err)
	}
	ctx := context.Background()
	sqltx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Errorf("db.BeginTx() error = %v", err)
	}
	tx := &Tx{sqltx}

	tests := []struct {
		name    string
		args    args
		want    *Tx
		wantErr bool
	}{
		{"typical", args{sqltx}, tx, false},
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

func TestTx_Create(t *testing.T) {
	dsn := datastoretest.NewPGDatasourceName(t)
	lgr := logger.NewLogger(os.Stdout, true)

	db, cleanup, err := datastore.NewDB(dsn, lgr)
	defer cleanup()
	if err != nil {
		t.Errorf("datastore.NewDB error = %v", err)
	}
	ctx := context.Background()
	m := newMovie(t)
	type args struct {
		ctx context.Context
		m   *movie.Movie
	}
	tests := []struct {
		name string
		//fields  fields
		args    args
		wantErr bool
	}{
		{"typical", args{ctx, m}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			sqltx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Errorf("db.BeginTx() error = %v", err)
			}
			tx := &Tx{
				Tx: sqltx,
			}
			if err := tx.Create(tt.args.ctx, tt.args.m); (err != nil) != tt.wantErr {
				t1.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				t1.Errorf("tx.Rollback() error = %v", rollbackErr)
			}
		})
	}
}

func TestTx_Delete(t1 *testing.T) {
	type args struct {
		ctx context.Context
		m   *movie.Movie
	}
	dsn := datastoretest.NewPGDatasourceName(t1)
	lgr := logger.NewLogger(os.Stdout, true)

	// I am intentionally not using the cleanup function that is
	// returned from NewDB as I need the DB to stay open for the test
	// t.Cleanup function
	db, _, err := datastore.NewDB(dsn, lgr)
	if err != nil {
		t1.Errorf("datastore.NewDB error = %v", err)
	}
	ctx := context.Background()

	m := newMovieDBHelper(t1, ctx, db)

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"typical", args{ctx, m}, false},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			sqltx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t1.Errorf("db.BeginTx() error = %v", err)
			}
			t := &Tx{
				Tx: sqltx,
			}
			if err := t.Delete(tt.args.ctx, tt.args.m); (err != nil) != tt.wantErr {
				t1.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if rollbackErr := t.Rollback(); rollbackErr != nil {
				t1.Errorf("tx.Rollback() error = %v", rollbackErr)
			}
		})
	}
}

func TestTx_Update(t *testing.T) {
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
	m := newMovieDBHelper(t, ctx, db)

	type args struct {
		ctx context.Context
		m   *movie.Movie
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"typical", args{ctx, m}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t1 *testing.T) {
			sqltx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t1.Errorf("db.BeginTx() error = %v", err)
			}
			tx := &Tx{
				Tx: sqltx,
			}
			if err := tx.Update(tt.args.ctx, tt.args.m); (err != nil) != tt.wantErr {
				t1.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				t1.Errorf("tx.Rollback() error = %v", rollbackErr)
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
