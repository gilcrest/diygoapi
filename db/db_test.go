package db

import (
	"testing"

	_ "github.com/lib/pq"
)

func Test_newMainDB(t *testing.T) {

	tests := []struct {
		name    string
		wantErr bool
	}{
		{"Test newMainDB", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newMainDB()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMainDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestNewDatastore(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"Test NewDatastore", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDatastore()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDatastore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func Test_newCacheDb(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"Test newCacheDb", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cDb := newCacheDb()
			conn := cDb.Get()
			err := conn.Err()

			if err != nil {
				t.Errorf("Error in newCacheDB = %s", err)
				return
			}
		})
	}
}
