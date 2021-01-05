package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/rs/zerolog"
)

func TestNewLogger(t *testing.T) {
	// empty string for TimeFieldFormat will write logs with UNIX time
	zerolog.TimeFieldFormat = ""
	// start a new logger with Stdout as the target
	lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()
	lgr = lgr.Hook(GCPSeverityHook{})

	tests := []struct {
		name string
		want zerolog.Logger
	}{
		{"New Logger", lgr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLogger(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLogger() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGCPSeverityHook_Run(t *testing.T) {
	// empty string for TimeFieldFormat will write logs with UNIX time
	zerolog.TimeFieldFormat = ""
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	var b bytes.Buffer
	lgr := zerolog.New(&b).With().Timestamp().Logger().Hook(GCPSeverityHook{})

	type args struct {
		f func()
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"default", args{func() { lgr.Log().Msg("") }}, "DEFAULT"},
		{zerolog.PanicLevel.String(), args{func() { lgr.Panic().Msg("") }}, "EMERGENCY"},
		//{zerolog.FatalLevel.String(), args{func() { lgr.Fatal().Msg("") }, zerolog.FatalLevel, "", "EMERGENCY"}},
		{zerolog.ErrorLevel.String(), args{func() { lgr.Error().Msg("") }}, "ERROR"},
		{zerolog.WarnLevel.String(), args{func() { lgr.Warn().Msg("") }}, "WARNING"},
		{zerolog.InfoLevel.String(), args{func() { lgr.Info().Msg("") }}, "INFO"},
		{zerolog.DebugLevel.String(), args{func() { lgr.Debug().Msg("") }}, "DEBUG"},
		{zerolog.TraceLevel.String(), args{func() { lgr.Trace().Msg("") }}, "DEBUG"},
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

			if got != tt.want {
				t.Errorf("event.Msg() = %q, want %q", got, tt.want)
			}
		})
	}
}
