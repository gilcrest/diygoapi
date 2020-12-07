package app

import (
	"bytes"
	"encoding/json"
	"os"
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
	lgr = lgr.Hook(GCPSeverityHook{})

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

func TestGCPSeverityHook_Run(t *testing.T) {
	// empty string for TimeFieldFormat will write logs with UNIX time
	zerolog.TimeFieldFormat = ""
	var b bytes.Buffer
	lgr := zerolog.New(&b).With().Timestamp().Logger().Hook(GCPSeverityHook{})

	type args struct {
		f     func()
		level zerolog.Level
		msg   string
		sev   string
	}
	tests := []struct {
		name string
		args args
	}{
		{zerolog.PanicLevel.String(), args{func() { lgr.Panic().Msg("") }, zerolog.PanicLevel, "", "EMERGENCY"}},
		//{zerolog.FatalLevel.String(), args{func() { lgr.Fatal().Msg("") }, zerolog.FatalLevel, "", "EMERGENCY"}},
		{zerolog.ErrorLevel.String(), args{func() { lgr.Error().Msg("") }, zerolog.ErrorLevel, "", "ERROR"}},
		{zerolog.WarnLevel.String(), args{func() { lgr.Warn().Msg("") }, zerolog.WarnLevel, "", "WARNING"}},
		{zerolog.InfoLevel.String(), args{func() { lgr.Info().Msg("") }, zerolog.InfoLevel, "", "INFO"}},
		{zerolog.DebugLevel.String(), args{func() { lgr.Debug().Msg("") }, zerolog.DebugLevel, "", "DEBUG"}},
		{zerolog.TraceLevel.String(), args{func() { lgr.Trace().Msg("") }, zerolog.TraceLevel, "", "DEBUG"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == zerolog.PanicLevel.String() {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Code should have panicked")
					}
				}()
			}
			b.Reset()
			tt.args.f()
			var dat map[string]interface{}
			if err := json.Unmarshal(b.Bytes(), &dat); err != nil {
				t.Fatalf("json Unmarshal error: %v", err)
			}
			got := dat["severity"].(string)
			want := tt.args.sev

			if got != want {
				t.Errorf("event.Msg() = %q, want %q", got, want)
			}
		})
	}
}
