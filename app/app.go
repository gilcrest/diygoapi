package app

import (
	"os"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/rs/zerolog"
)

// Application contains the app configurations and Datastore
type Application struct {
	// Datastorer is an interface type meant to be the
	// persistence mechanism. It can be a
	// SQL database (PostgreSQL) or a mock database
	Datastorer datastore.Datastorer
	// Logger
	Logger zerolog.Logger
}

// NewApplication initializes an Application struct
func NewApplication(datastorer datastore.Datastorer, logger zerolog.Logger) *Application {
	return &Application{
		Datastorer: datastorer,
		Logger:     logger,
	}
}

// EnvName is the environment Name int representation
// Using iota, 1 (Production) is the lowest,
// 2 (Staging) is 2nd lowest, and so on...
type EnvName uint8

// EnvName of environment.
const (
	Production EnvName = iota + 1 // Production (1)
	Staging                       // Staging (2)
	QA                            // QA (3)
	Local                         // Local (4)
)

func (n EnvName) String() string {
	switch n {
	case Production:
		return "Production"
	case Staging:
		return "Staging"
	case QA:
		return "QA"
	case Local:
		return "Local"
	}
	return "unknown_name"
}

// NewLogger sets up the zerolog.Logger
func NewLogger(lvl zerolog.Level) zerolog.Logger {

	// empty string for TimeFieldFormat will write logs with UNIX time
	zerolog.TimeFieldFormat = ""

	// set logging level based on input
	zerolog.SetGlobalLevel(lvl)

	// start a new logger with Stdout as the target
	lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// Add Severity Hook. Zerolog by default outputs structured logs
	// with "level":"error" as its leveling. Google Cloud as an
	// example expects "severity","ERROR" for its leveling. This
	// hook will add severity to each message
	lgr = lgr.Hook(GCPSeverityHook{})

	lgr.Info().Msgf("logging level set to %s", lvl)

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
