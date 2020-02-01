package datastore

import (
	"database/sql"
	"reflect"
	"testing"
)

func TestNewDatastorer(t *testing.T) {
	type args struct {
		n  Name
		db *sql.DB
	}

	db := new(sql.DB)
	tests := []struct {
		name    string
		args    args
		want    Datastorer
		wantErr bool
	}{
		{"Mock w non-nil db", args{MockedDatastore, db}, nil, true},
		{"Mock w nil db", args{MockedDatastore, nil}, &MockDatastore{}, false},
		{"Datastore w nil db", args{LocalDatastore, nil}, nil, true},
		{"Datastore w non-nil db", args{LocalDatastore, db}, &Datastore{DB: db}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewDatastorer(tt.args.n, tt.args.db)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDatastorer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewDatastorer() got = %v, want %v", got, tt.want)
			}
		})
	}
}
