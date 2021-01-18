package datastore

import (
	"context"
	"database/sql"
	"os"
	"reflect"
	"testing"

	"github.com/gilcrest/go-api-basic/domain/errs"

	"github.com/pkg/errors"

	"github.com/gilcrest/go-api-basic/domain/logger"

	"github.com/matryer/is"
)

func TestNewPGDatasourceName(t *testing.T) {
	is := is.New(t)

	got := NewPGDatasourceName("localhost", "go_api_basic", "postgres", "", 5432)

	want := PGDatasourceName{
		Host:     "localhost",
		Port:     5432,
		DBName:   "go_api_basic",
		User:     "postgres",
		Password: "",
	}

	is.Equal(got, want)
}

func TestPGDatasourceName_String(t *testing.T) {
	type fields struct {
		Host     string
		Port     int
		DBName   string
		User     string
		Password string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"with password", fields{Host: "localhost", Port: 8080, DBName: "go_api_basic", User: "postgres", Password: "supahsecret"}, "host=localhost port=8080 dbname=go_api_basic user=postgres password=supahsecret sslmode=disable"},
		{"without password", fields{Host: "localhost", Port: 8080, DBName: "go_api_basic", User: "postgres", Password: ""}, "host=localhost port=8080 dbname=go_api_basic user=postgres sslmode=disable"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := PGDatasourceName{
				Host:     tt.fields.Host,
				Port:     tt.fields.Port,
				DBName:   tt.fields.DBName,
				User:     tt.fields.User,
				Password: tt.fields.Password,
			}
			if got := dsn.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatastore_DB(t *testing.T) {
	is := is.New(t)

	logger := logger.NewLogger(os.Stdout, true)

	ogdb, cleanup, err := NewDB(NewPGDatasourceName("localhost", "go_api_basic", "postgres", "", 5432), logger)
	defer cleanup()
	if err != nil {
		t.Fatal(err)
	}
	ds := Datastore{db: ogdb}
	db := ds.DB()

	is.Equal(db, ogdb)
}

func TestNewDatastore(t *testing.T) {
	is := is.New(t)

	logger := logger.NewLogger(os.Stdout, true)

	db, cleanup, err := NewDB(NewPGDatasourceName("localhost", "go_api_basic", "postgres", "", 5432), logger)
	defer cleanup()
	if err != nil {
		t.Fatal(err)
	}
	got := NewDatastore(db)

	want := &Datastore{db: db}

	is.Equal(got, want)
}

func TestNewNullInt64(t *testing.T) {
	type args struct {
		i int64
	}
	tests := []struct {
		name string
		args args
		want sql.NullInt64
	}{
		{"has value", args{i: 23}, sql.NullInt64{Int64: 23, Valid: true}},
		{"zero value", args{i: 0}, sql.NullInt64{Int64: 0, Valid: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNullInt64(tt.args.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNullInt64() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewNullString(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want sql.NullString
	}{
		{"has value", args{s: "foobar"}, sql.NullString{String: "foobar", Valid: true}},
		{"zero value", args{s: ""}, sql.NullString{String: "", Valid: false}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNullString(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewNullString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDatastore_BeginTx(t *testing.T) {
	type fields struct {
		db      *sql.DB
		cleanup func()
	}
	type args struct {
		ctx context.Context
	}

	dsn := NewPGDatasourceName("localhost", "go_api_basic", "postgres", "", 5432)
	lgr := logger.NewLogger(os.Stdout, true)

	db, cleanup, err := NewDB(dsn, lgr)
	if err != nil {
		t.Errorf("datastore.NewDB error = %v", err)
	}
	ctx := context.Background()
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"typical", fields{db, cleanup}, args{ctx}, false},
		{"closed db", fields{db, cleanup}, args{ctx}, true},
		{"nil db", fields{nil, cleanup}, args{ctx}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &Datastore{
				db: tt.fields.db,
			}
			if tt.wantErr == true {
				tt.fields.cleanup()
			}
			got, err := ds.BeginTx(tt.args.ctx)
			t.Logf("BeginTx error = %v", err)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeginTx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if ((err != nil) != tt.wantErr) && got == nil {
				t.Errorf("BeginTx() returned nil and should not")
			}
			tt.fields.cleanup()
		})
	}
}

func TestDatastore_RollbackTx(t *testing.T) {
	type fields struct {
		db *sql.DB
	}
	type args struct {
		tx  *sql.Tx
		err error
	}

	dsn := NewPGDatasourceName("localhost", "go_api_basic", "postgres", "", 5432)
	lgr := logger.NewLogger(os.Stdout, true)

	db, cleanup, err := NewDB(dsn, lgr)
	defer cleanup()
	if err != nil {
		t.Errorf("datastore.NewDB error = %v", err)
	}
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Errorf("db.BeginTx error = %v", err)
	}
	tx2, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Errorf("db.BeginTx error = %v", err)
	}
	err = tx2.Commit()
	if err != nil {
		t.Errorf("tx2.Commit() error = %v", err)
	}

	err = errors.New("some error")

	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{"typical", fields{db}, args{tx, err}},
		{"nil tx", fields{db}, args{nil, err}},
		{"already committed tx", fields{db}, args{tx2, err}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Log(tt.name)
			ds := &Datastore{
				db: tt.fields.db,
			}
			err := ds.RollbackTx(tt.args.tx, tt.args.err)
			// RollbackTx only returns an *errs.Error
			e, _ := err.(*errs.Error)
			t.Logf("error = %v", e)
			if tt.args.tx == nil && e.Code != "nil_tx" {
				t.Fatalf("ds.RollbackTx() tx was nil, but incorrect error returned = %v", e)
			}
			// I know this is weird, but it's the only way I could think to test this.
			if tt.name == "already committed tx" && e.Code != "rollback_err" {
				t.Fatalf("ds.RollbackTx() tx was already committed, but incorrect error returned = %v", e)
			}
		})
	}
}

func TestDatastore_CommitTx(t *testing.T) {
	type fields struct {
		db *sql.DB
	}
	type args struct {
		tx *sql.Tx
	}
	dsn := NewPGDatasourceName("localhost", "go_api_basic", "postgres", "", 5432)
	lgr := logger.NewLogger(os.Stdout, true)

	db, cleanup, err := NewDB(dsn, lgr)
	defer cleanup()
	if err != nil {
		t.Errorf("datastore.NewDB error = %v", err)
	}
	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Errorf("db.BeginTx error = %v", err)
	}
	tx2, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Errorf("db.BeginTx error = %v", err)
	}
	err = tx2.Commit()
	if err != nil {
		t.Errorf("tx2.Commit() error = %v", err)
	}
	tx3, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Errorf("db.BeginTx error = %v", err)
	}
	err = tx3.Rollback()
	if err != nil {
		t.Errorf("tx2.Commit() error = %v", err)
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"typical", fields{db}, args{tx}, false},
		{"already committed", fields{db}, args{tx2}, true},
		{"already rolled back", fields{db}, args{tx3}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &Datastore{
				db: tt.fields.db,
			}
			if err := ds.CommitTx(tt.args.tx); (err != nil) != tt.wantErr {
				t.Errorf("CommitTx() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
