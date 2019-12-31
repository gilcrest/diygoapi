package datastore

import (
	"testing"
)

func Test_NewLocalDB(t *testing.T) {
	type args struct {
		n DSName
	}
	tests := []struct {
		name string
		args args
	}{
		{"App DB", args{LocalDatastore}},
		// {"Log DB", args{LogDatastore}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _, err := NewDB(tt.args.n)
			if err != nil {
				t.Errorf("Error from newDB = %v", err)
			}
			err = db.Ping()
			if err != nil {
				t.Errorf("Error pinging database = %v", err)
			}
		})
	}
}
