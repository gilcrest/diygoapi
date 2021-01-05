package app

import (
	"reflect"
	"testing"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/rs/zerolog"
)

func TestEnvName_String(t *testing.T) {
	tests := []struct {
		name string
		n    EnvName
		want string
	}{
		{"Production", Production, "Production"},
		{"Staging", Staging, "Staging"},
		{"QA", QA, "QA"},
		{"Local", Local, "Local"},
		{"Unknown Name", 99, "unknown_name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewApplication(t *testing.T) {
	type args struct {
		ds  datastore.Datastorer
		log zerolog.Logger
	}
	tests := []struct {
		name string
		args args
		want *Application
	}{
		{"New Application", args{
			ds:  nil,
			log: zerolog.Logger{},
		}, &Application{
			Datastorer: nil,
			Logger:     zerolog.Logger{},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewApplication(tt.args.ds, tt.args.log); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewApplication() = %v, want %v", got, tt.want)
			}
		})
	}
}
