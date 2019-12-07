package app

import (
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/rs/zerolog"
)

// Application is the main server struct for Guestbook. It contains the state of
// the most recently read message of the day.
type Application struct {
	EnvName EnvName
	// PostgreSQL database
	DS datastore.Datastore
	// Logger
	Logger zerolog.Logger
}

// NewApplication creates a new application struct
func NewApplication(en EnvName, ds datastore.Datastore, log zerolog.Logger) *Application {
	return &Application{
		EnvName: en,
		DS:      ds,
		Logger:  log,
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
