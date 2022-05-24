package logger

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"regexp"
	"testing"

	"github.com/gilcrest/diy-go-api/domain/errs"
	"github.com/rs/zerolog"
)

func TestNewLogger(t *testing.T) {
	// empty string for TimeFieldFormat will write logs with UNIX time
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	// start a new logger with Stdout as the target
	lgr := zerolog.New(os.Stdout).Level(zerolog.InfoLevel).With().Timestamp().Logger()
	lgr = lgr.Hook(GCPSeverityHook{})

	type args struct {
		w             io.Writer
		lvl           zerolog.Level
		withTimestamp bool
	}

	tests := []struct {
		name string
		args args
		want zerolog.Logger
	}{
		{"stdout", args{os.Stdout, zerolog.InfoLevel, true}, lgr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLogger(tt.args.w, tt.args.lvl, tt.args.withTimestamp); !reflect.DeepEqual(got, tt.want) {
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

func TestWriteErrorStackGlobal(t *testing.T) {
	t.Run("with stack", func(t *testing.T) {
		WriteErrorStackGlobal(true)
		out := &bytes.Buffer{}
		logger := zerolog.New(out)

		err := errs.E("some error")
		e := err.(*errs.Error)
		logger.Log().Stack().Err(e.Err).Msg("")

		got := out.String()
		want := `{"stack".*`
		if ok, _ := regexp.MatchString(want, got); !ok {
			t.Errorf("invalid log output:\ngot:  %v\nwant: %v", got, want)
		}
	})

	t.Run("without stack", func(t *testing.T) {
		WriteErrorStackGlobal(false)
		out := &bytes.Buffer{}
		logger := zerolog.New(out)

		err := errs.E("some error")
		e := err.(*errs.Error)
		logger.Log().Stack().Err(e.Err).Msg("")

		got := out.String()
		want := `{"error".*`
		if ok, _ := regexp.MatchString(want, got); !ok {
			t.Errorf("invalid log output:\ngot:  %v\nwant: %v", got, want)
		}
	})
}
