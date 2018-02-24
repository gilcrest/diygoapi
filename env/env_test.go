package env

import (
	"testing"
)

func TestNewDatastore(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"Test Datastore struct creation", false},
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
