package datastore

import (
	"database/sql"
	"reflect"
	"testing"

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

func TestDatastore_DB(t *testing.T) {
	is := is.New(t)

	logger := logger.NewLogger()

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

	logger := logger.NewLogger()

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
