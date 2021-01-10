package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/pkg/errors"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/logger"

	"github.com/rs/zerolog"
)

// cliFlags are the command line flags parsed at startup
type cliFlags struct {
	logLevel   string
	port       int
	dbhost     string
	dbport     int
	dbname     string
	dbuser     string
	dbpassword string
}

func main() {
	// Initialize cliFlags and return a pointer to it
	cf := new(cliFlags)

	// loglvl flag allows for setting logging level, e.g. to run the server
	// with level set to debug, it'd be: ./server loglvl=debug
	// If not set, defaults to error
	flag.StringVar(&cf.logLevel, "loglvl", "info", "sets log level (debug, warn, error, fatal, panic, disabled)")

	// port flag is what http.ListenAndServe will listen on. default is 8080 if not set
	flag.IntVar(&cf.port, "port", 8080, "network port to listen on")

	// dbhost is the database host
	flag.StringVar(&cf.dbhost, "dbhost", "", "postgresql database host")

	// dbport is the database host
	flag.IntVar(&cf.dbport, "dbport", 0, "postgresql database port")

	// dbname is the database name
	flag.StringVar(&cf.dbname, "dbname", "", "postgresql database name")

	// dbname is the database name
	flag.StringVar(&cf.dbuser, "dbuser", "", "postgresql database user")

	// dbname is the database name
	flag.StringVar(&cf.dbpassword, "dbpassword", "", "postgresql database password")

	// Parse the command line flags from above
	flag.Parse()

	// setup logger with appropriate defaults
	logger := logger.NewLogger(os.Stdout, true)

	// determine logging level
	loglvl := newLogLevel(cf)

	// set global logging level based on flag input
	zerolog.SetGlobalLevel(loglvl)
	logger.Info().Msgf("logging level set to %s", loglvl)

	// set global logging time field format to Unix timestamp
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// validate port in acceptable range
	if cf.port < 0 || cf.port > 65535 {
		logger.Fatal().Msgf("port %d is not within valid port range (0 to 65535", cf.port)
	}

	dsn, err := newPGDatasourceName(cf)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error returned from newPGDatasourceName")
	}

	// initialize a non-nil, empty context
	ctx := context.Background()

	// newServer function returns a pointer to a gocloud server, a
	// cleanup function and an error
	srv, cleanup, err := newServer(ctx, logger, dsn)
	if err != nil {
		logger.Fatal().Err(err).Msg("Error returned from newServer")
	}
	defer cleanup()

	// Listen and serve HTTP
	logger.Fatal().Err(srv.ListenAndServe(fmt.Sprintf(":%d", cf.port))).Msg("Fatal Server Error")
}

// newLogLevel sets up the logging level (e.g. Debug, Info, Error, etc.)
// It takes a pointer to a string as that is how a parsed command line flag news
// and the intention is for the name to be set at run time
func newLogLevel(flags *cliFlags) zerolog.Level {

	var lvl zerolog.Level

	switch flags.logLevel {
	case "debug":
		lvl = zerolog.DebugLevel
	case "info":
		lvl = zerolog.InfoLevel
	case "warn":
		lvl = zerolog.WarnLevel
	case "error":
		lvl = zerolog.ErrorLevel
	case "fatal":
		lvl = zerolog.FatalLevel
	case "panic":
		lvl = zerolog.PanicLevel
	case "disabled":
		lvl = zerolog.Disabled
	default:
		lvl = zerolog.InfoLevel
	}

	return lvl
}

func newPGDatasourceName(flags *cliFlags) (datastore.PGDatasourceName, error) {

	// Constants for the PostgreSQL Database connection
	const (
		pgDBHost     string = "PG_APP_HOST"
		pgDBPort     string = "PG_APP_PORT"
		pgDBName     string = "PG_APP_DBNAME"
		pgDBUser     string = "PG_APP_USERNAME"
		pgDBPassword string = "PG_APP_PASSWORD"
	)

	var (
		ds         datastore.PGDatasourceName
		dbHost     string
		dbPort     int
		dbName     string
		dbUser     string
		dbPassword string
		ok         bool
		err        error
	)

	// check cli flag first, if has value, use it, otherwise
	// check environment variable
	dbHost = flags.dbhost
	if dbHost == "" {
		dbHost, ok = os.LookupEnv(pgDBHost)
		if !ok {
			return ds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", pgDBHost)))
		}
	}

	// check cli flag first, if has value, use it, otherwise
	// check environment variable
	dbPort = flags.dbport
	if dbPort == 0 {
		p, ok := os.LookupEnv(pgDBPort)
		if !ok {
			return ds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", pgDBPort)))
		}
		dbPort, err = strconv.Atoi(p)
		if err != nil {
			return ds, errs.E(errors.New(fmt.Sprintf("Unable to convert db port %s to int", p)))
		}
	}

	// check cli flag first, if has value, use it, otherwise
	// check environment variable
	dbName = flags.dbname
	if dbName == "" {
		dbName, ok = os.LookupEnv(pgDBName)
		if !ok {
			return ds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", pgDBName)))
		}
	}

	// check cli flag first, if has value, use it, otherwise
	// check environment variable
	dbUser = flags.dbuser
	if dbUser == "" {
		dbUser, ok = os.LookupEnv(pgDBUser)
		if !ok {
			return ds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", pgDBUser)))
		}
	}

	// check cli flag first, if has value, use it, otherwise
	// check environment variable
	dbPassword = flags.dbpassword
	if dbPassword == "" {
		dbPassword, ok = os.LookupEnv(pgDBPassword)
		if !ok {
			return ds, errs.E(errors.New(fmt.Sprintf("No environment variable found for %s", pgDBPassword)))
		}
	}

	return datastore.NewPGDatasourceName(dbHost, dbName, dbUser, dbPassword, dbPort), nil
}
