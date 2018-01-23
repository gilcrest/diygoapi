package env

import (
	"testing"

	"github.com/gilcrest/go-API-template/pkg/datastore"
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
			_, err := datastore.NewDatastore()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDatastore() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
