package main

import (
	"testing"
)

func Test_newDB(t *testing.T) {
	type args struct {
		n dbName
	}
	tests := []struct {
		name string
		args args
	}{
		{"App DB", args{appDB}},
		{"Log DB", args{logDB}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := newDB(tt.args.n)
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
