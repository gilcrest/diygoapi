package main

import (
	"database/sql"
	"os"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// application is the main server struct for Guestbook. It contains the state of
// the most recently read message of the day.
type application struct {
	envName envName
	// multiplex router
	router *mux.Router
	// PostgreSQL database
	db *sql.DB
	// Logger
	logger zerolog.Logger
}

// newApplication creates a new application struct
func provideApplication(nm envName, rtr *mux.Router, db *sql.DB, logger zerolog.Logger) *application {
	return &application{
		envName: nm,
		router:  rtr,
		db:      db,
		logger:  logger,
	}
}

// envName is the environment Name int representation
// Using iota, 1 (Production) is the lowest,
// 2 (Staging) is 2nd lowest, and so on...
type envName uint8

// EnvName of environment.
const (
	production envName = iota + 1 // Production (1)
	staging                       // Staging (2)
	qa                            // QA (3)
	dev                           // Dev (4)
)

func (n envName) String() string {
	switch n {
	case production:
		return "Production"
	case staging:
		return "Staging"
	case qa:
		return "QA"
	case dev:
		return "Dev"
	}
	return "unknown_name"
}

// provideName sets up the environment (e.g. Production, Staging, QA, etc.)
// It takes a pointer to a string as that is how a parsed command line flag provides
// and the intention is for the name to be set at run time
func provideName(flags *cliFlags) envName {

	var name envName

	switch flags.envName {
	case "dev":
		name = dev
	case "qa":
		name = qa
	case "stg":
		name = staging
	case "prod":
		name = production
	default:
		name = dev
	}

	log.Log().Msgf("Environment set to %s", name)

	return name
}

// ProvideLogger sets up the zerolog.Logger
func provideLogger(lvl zerolog.Level) zerolog.Logger {
	// empty string for TimeFieldFormat will write logs with UNIX time
	zerolog.TimeFieldFormat = ""
	// set logging level based on input
	zerolog.SetGlobalLevel(lvl)
	// start a new logger with Stdout as the target
	lgr := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return lgr
}

// ProvideLogLevel sets up the logging level (e.g. Debug, Info, Error, etc.)
// It takes a pointer to a string as that is how a parsed command line flag provides
// and the intention is for the name to be set at run time
func provideLogLevel(flags *cliFlags) zerolog.Level {

	var lvl zerolog.Level

	switch flags.logLevel {
	case "debug":
		lvl = zerolog.DebugLevel
	case "info":
		lvl = zerolog.InfoLevel
	case "warn":
		lvl = zerolog.WarnLevel
	case "fatal":
		lvl = zerolog.FatalLevel
	case "panic":
		lvl = zerolog.PanicLevel
	case "disabled":
		lvl = zerolog.Disabled
	default:
		lvl = zerolog.ErrorLevel
	}

	log.Log().Msgf("Logging Level set to %s", lvl)

	return lvl
}
