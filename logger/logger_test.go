package logger

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/rs/zerolog"
)

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

//func TestWriteErrorStackGlobal(t *testing.T) {
//	t.Run("with stack", func(t *testing.T) {
//		WriteErrorStack(true)
//		out := &bytes.Buffer{}
//		logger := zerolog.New(out)
//
//		err := errs.E("some error")
//		e := err.(*errs.Error)
//		logger.Log().Stack().Err(e.Err).Msg("")
//
//		got := out.String()
//		want := `{"stack".*`
//		if ok, _ := regexp.MatchString(want, got); !ok {
//			t.Errorf("invalid log output:\ngot:  %v\nwant: %v", got, want)
//		}
//	})
//
//	t.Run("without stack", func(t *testing.T) {
//		WriteErrorStack(false)
//		out := &bytes.Buffer{}
//		logger := zerolog.New(out)
//
//		err := errs.E("some error")
//		e := err.(*errs.Error)
//		logger.Log().Stack().Err(e.Err).Msg("")
//
//		got := out.String()
//		want := `{"error".*`
//		if ok, _ := regexp.MatchString(want, got); !ok {
//			t.Errorf("invalid log output:\ngot:  %v\nwant: %v", got, want)
//		}
//	})
//}

func ExampleNewWithGCPHook() {
	lgr := NewWithGCPHook(os.Stdout, zerolog.DebugLevel, false)
	lgr.Trace().Msg("Trace is lower than Debug, this message is filtered out")
	lgr.Debug().Msg("This is a log at the Debug level")

	zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	lgr.Debug().Msg("Logging level raised to Error, Debug is lower than Error, this message is filtered out")
	lgr.Error().Msg("This is a log at the Error level")

	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	lgr.Trace().Msg("Setting Global level will not impact minimum set to logger, this trace message will still be filtered out")
	lgr.Debug().Msg("Logging level raised all the way down to Trace level, Debug is higher than Trace, this will log")

	// Output:
	// {"level":"debug","severity":"DEBUG","message":"This is a log at the Debug level"}
	// {"level":"error","severity":"ERROR","message":"This is a log at the Error level"}
	// {"level":"debug","severity":"DEBUG","message":"Logging level raised all the way down to Trace level, Debug is higher than Trace, this will log"}
}
