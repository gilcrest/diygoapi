package env

import (
	"reflect"
	"testing"

	"github.com/rs/zerolog"
)

func TestNewEnv(t *testing.T) {
	tests := []struct {
		name    string
		want    *Env
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewEnv(zerolog.DebugLevel)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}
