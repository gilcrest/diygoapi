// Package logger has helpers to setup a zerolog.Logger
//
//	https://github.com/rs/zerolog
package logger

import (
	"io"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// New is a convenience function to initialize a zerolog.Logger
// with an initial minimum accepted level and timestamp (if true)
// for a given io.Writer.
func New(w io.Writer, lvl zerolog.Level, withTimestamp bool) zerolog.Logger {
	// logger is initialized with the writer and level passed in.
	// All logs will be written at the given level (unless raised
	// using zerolog.SetGlobalLevel)
	lgr := zerolog.New(w).Level(lvl)
	if withTimestamp {
		lgr = lgr.With().Timestamp().Logger()
	}

	return lgr
}

// NewWithGCPHook is a convenience function to initialize a zerolog.Logger
// with an initial minimum accepted level and timestamp (if true) for a
// given io.Writer. In addition, it adds a Google Cloud Platform (GCP)
// Severity Hook. Zerolog by default outputs structured logs with
// "level":"error" as its leveling. Google Cloud, as an example, expects
// "severity","ERROR" for its leveling. This hook will add severity
// to each message.
func NewWithGCPHook(w io.Writer, lvl zerolog.Level, withTimestamp bool) zerolog.Logger {
	// logger is initialized with the writer and level passed in.
	// All logs will be written at the given level (unless raised
	// using zerolog.SetGlobalLevel)
	lgr := New(w, lvl, withTimestamp)

	lgr = lgr.Hook(GCPSeverityHook{})

	return lgr
}

// The GCPSeverityHook struct satisfies the zerolog.Hook interface
// as it has the Run method defined with the appropriate parameters
type GCPSeverityHook struct{}

// Run method satisfies zerolog.Hook interface and adds a severity
// level to all logs, given zerolog.Level passed in. Zerolog levels
// are mapped to log levels recognized by GCP
func (h GCPSeverityHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {

	const lvlKey string = "severity"

	switch level {
	case zerolog.PanicLevel:
		e.Str(lvlKey, "EMERGENCY")
	case zerolog.FatalLevel:
		e.Str(lvlKey, "EMERGENCY")
	case zerolog.ErrorLevel:
		e.Str(lvlKey, "ERROR")
	case zerolog.WarnLevel:
		e.Str(lvlKey, "WARNING")
	case zerolog.InfoLevel:
		e.Str(lvlKey, "INFO")
	case zerolog.DebugLevel:
		e.Str(lvlKey, "DEBUG")
	case zerolog.TraceLevel:
		e.Str(lvlKey, "DEBUG")
	default:
		e.Str(lvlKey, "DEFAULT")
	}
}

// LogErrorStackViaPkgErrors is a convenience function to set the zerolog
// ErrorStackMarshaler global variable.
// If true, writes error stacks for logs using "github.com/pkg/errors".
// If false, will use the internal errs.Op stack instead of "github.com/pkg/errors".
func LogErrorStackViaPkgErrors(p bool) {
	if !p {
		zerolog.ErrorStackMarshaler = nil
		return
	}
	// set ErrorStackMarshaler to pkgerrors.MarshalStack
	// to enable error stack traces
	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
}
