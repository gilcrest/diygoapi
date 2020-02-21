package app

import (
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/rs/zerolog"
	"os"
)

// Application contains the app configurations and Datastore
type Application struct {
	// Environment Name
	EnvName EnvName
	// Datastorer is an interface type meant to be the
	// persistence mechanism. It can be a
	// SQL database (PostgreSQL) or a mock database
	Datastorer datastore.Datastorer
	// Logger
	Logger zerolog.Logger
}

// NewApplication creates a new application struct
func NewApplication(en EnvName, ds datastore.Datastorer, log zerolog.Logger) *Application {
	return &Application{
		EnvName:    en,
		Datastorer: ds,
		Logger:     log,
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

	lgr.Log().Msgf("Logging Level set to %s", lvl)

	return lgr
}
