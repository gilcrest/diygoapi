package datastore

import (
	"testing"
)

func Test_ProvideDB(t *testing.T) {
	type args struct {
		n DBName
	}
	tests := []struct {
		name string
		args args
	}{
		{"App DB", args{AppDB}},
		{"Log DB", args{LogDB}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := ProvideDB(tt.args.n)
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
