package datastore

import (
	"database/sql"
	"reflect"
	"testing"

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
	type fields struct {
		db *sql.DB
	}
	tests := []struct {
		name   string
		fields fields
		want   *sql.DB
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := &Datastore{
				db: tt.fields.db,
			}
			if got := ds.DB(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DB() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewDatastore(t *testing.T) {
	type args struct {
		db *sql.DB
	}
	tests := []struct {
		name string
		args args
		want *Datastore
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewDatastore(tt.args.db); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDatastore() = %v, want %v", got, tt.want)
			}
		})
	}
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
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
