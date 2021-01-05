package app

import (
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
