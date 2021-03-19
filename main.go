package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/logger"
)

const (
	// exitFail is the exit code if the program
	// fails.
	exitFail = 1
)

func main() {
	if err := run(os.Args); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(exitFail)
	}
}

func run(args []string) error {

	flgs, err := newFlags(args)
	if err != nil {
		return err
	}

	// setup logger with appropriate defaults
	lgr := logger.NewLogger(os.Stdout, true)

	// determine logging level
	loglevel := newLogLevel(flgs.loglvl)

	// set global logging level based on flag input
	zerolog.SetGlobalLevel(loglevel)
	lgr.Info().Msgf("logging level set to %s", loglevel)

	// set global logging time field format to Unix timestamp
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// validate port in acceptable range
	err = portRange(flgs.port)
	if err != nil {
		lgr.Fatal().Err(err).Msg("portRange() error")
	}

	//get struct holding PostgreSQL datasource name details
	dsn := datastore.NewPGDatasourceName(flgs.dbhost, flgs.dbname, flgs.dbuser, flgs.dbpassword, flgs.dbport)

	// initialize a non-nil, empty context
	ctx := context.Background()

	// newServer function returns a pointer to a gocloud server, a
	// cleanup function and an error
	srv, cleanup, err := newServer(ctx, lgr, dsn)
	if err != nil {
		lgr.Fatal().Err(err).Msg("Error returned from newServer")
	}
	defer cleanup()

	// Listen and serve HTTP
	lgr.Fatal().Err(srv.ListenAndServe(fmt.Sprintf(":%d", flgs.port))).Msg("Fatal Server Error")

	return nil
}

type flags struct {
	// log-level flag allows for setting logging level, e.g. to run the server
	// with level set to debug, it'd be: ./server -log-level=debug
	// If not set, defaults to error
	loglvl string

	// port flag is what http.ListenAndServe will listen on. default is 8080 if not set
	port int

	// dbhost is the database host
	dbhost string

	// dbport is the database port
	dbport int

	// dbname is the database name
	dbname string

	// dbuser is the database user
	dbuser string

	// dbpassword is the database user's password
	dbpassword string
}

// newFlags parses the command line flags using ff and returns
// a flags struct or an error
func newFlags(args []string) (flgs flags, err error) {
	// create new FlagSet using the program name being executed (args[0])
	// as the name of the FlagSet
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)

	var (
		loglvl     = fs.String("log-level", "info", "sets log level (debug, warn, error, fatal, panic, disabled), (also via LOG_LEVEL)")
		port       = fs.Int("port", 8080, "listen port for server (also via PORT)")
		dbhost     = fs.String("db-host", "", "postgresql database host (also via DB_HOST)")
		dbport     = fs.Int("db-port", 5432, "postgresql database port (also via DB_PORT)")
		dbname     = fs.String("db-name", "", "postgresql database name (also via DB_NAME)")
		dbuser     = fs.String("db-user", "", "postgresql database user (also via DB_USER)")
		dbpassword = fs.String("db-password", "", "postgresql database password (also via DB_PASSWORD)")
	)

	// Parse the command line flags from above
	err = ff.Parse(fs, args[1:], ff.WithEnvVarNoPrefix())
	if err != nil {
		return flgs, err
	}

	return flags{
		loglvl:     *loglvl,
		port:       *port,
		dbhost:     *dbhost,
		dbport:     *dbport,
		dbname:     *dbname,
		dbuser:     *dbuser,
		dbpassword: *dbpassword,
	}, nil
}

// newLogLevel sets up the logging level (e.g. Debug, Info, Error, etc.)
func newLogLevel(loglvl string) zerolog.Level {

	var lvl zerolog.Level

	switch loglvl {
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

// portRange validates the port be in an acceptable range
func portRange(port int) error {
	if port < 0 || port > 65535 {
		return errs.E(errors.New(fmt.Sprintf("port %d is not within valid port range (0 to 65535)", port)))
	}
	return nil
}
