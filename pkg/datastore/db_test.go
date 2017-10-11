package datastore

import (
	"testing"

	_ "github.com/lib/pq"
)

func TestNewMainDB(t *testing.T) {

	tests := []struct {
		name    string
		wantErr bool
	}{
		{"Test MainDB connection", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMainDB()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewMainDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
