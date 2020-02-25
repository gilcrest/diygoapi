package app

import (
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/rs/zerolog"
	"os"
	"reflect"
	"testing"
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
		en  EnvName
		ds  datastore.Datastorer
		log zerolog.Logger
	}
	tests := []struct {
		name string
		args args
		want *Application
	}{
		{"New Application", args{
			en:  Local,
			ds:  nil,
			log: zerolog.Logger{},
		}, &Application{
			EnvName:    Local,
			Mock:       false,
			Datastorer: nil,
			Logger:     zerolog.Logger{},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewApplication(tt.args.en, tt.args.ds, tt.args.log); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewApplication() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	type args struct {
		lvl zerolog.Level
	}

	// empty string for TimeFieldFormat will write logs with UNIX time
	zerolog.TimeFieldFormat = ""
	// set logging level based on input
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	// start a new logger with Stdout as the target
	lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()

	tests := []struct {
		name string
		args args
		want zerolog.Logger
	}{
		{"New Logger", args{zerolog.DebugLevel}, lgr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLogger(tt.args.lvl); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLogger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMockedApplication(t *testing.T) {
	type args struct {
		en  EnvName
		ds  datastore.Datastorer
		log zerolog.Logger
	}
	tests := []struct {
		name string
		args args
		want *Application
	}{
		{"New Mocked Application", args{
			en:  Local,
			ds:  nil,
			log: zerolog.Logger{},
		}, &Application{
			EnvName:    Local,
			Mock:       true,
			Datastorer: nil,
			Logger:     zerolog.Logger{},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMockedApplication(tt.args.en, tt.args.ds, tt.args.log); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMockedApplication() = %v, want %v", got, tt.want)
			}
		})
	}
}
