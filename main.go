package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/peterbourgon/ff/v3"
	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/datastore/moviestore"
	"github.com/gilcrest/go-api-basic/datastore/pingstore"
	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/logger"
	"github.com/gilcrest/go-api-basic/domain/random"
	"github.com/gilcrest/go-api-basic/gateway/authgateway"
	"github.com/gilcrest/go-api-basic/service"
)

const (
	// exitFail is the exit code if the program
	// fails.
	exitFail = 1
	// log level environment variable name
	loglevelEnv string = "LOG_LEVEL"
	// minimum accepted log level environment variable name
	logLevelMinEnv string = "LOG_LEVEL_MIN"
	// log error stack environment variable name
	logErrorStackEnv string = "LOG_ERROR_STACK"
	// server port environment variable name
	portEnv string = "PORT"
	// database host environment variable name
	dbHostEnv string = "DB_HOST"
	// database port environment variable name
	dbPortEnv string = "DB_PORT"
	// database name environment variable name
	dbNameEnv string = "DB_NAME"
	// database user environment variable name
	dbUserEnv string = "DB_USER"
	// database user password environment variable name
	dbPasswordEnv string = "DB_PASSWORD"
)

type flags struct {
	// log-level flag allows for setting logging level, e.g. to run the server
	// with level set to debug, it'd be: ./server -log-level=debug
	// If not set, defaults to error
	loglvl string

	// log-level-min flag sets the minimum accepted logging level
	// - e.g. in production, you may have a policy to never allow logs at
	// trace level. You could set the minimum log level to Debug. Even
	// if the Global log level is set to Trace, only logs at Debug
	// and above would be logged. Default level is trace.
	logLvlMin string

	// logErrorStack flag determines whether or not a full error stack
	// should be logged. If true, error stacks are logged, if false,
	// just the error is logged
	logErrorStack bool

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
		logLvlMin     = fs.String("log-level-min", "trace", fmt.Sprintf("sets minimum log level (trace, debug, info, warn, error, fatal, panic, disabled), (also via %s)", logLevelMinEnv))
		loglvl        = fs.String("log-level", "info", fmt.Sprintf("sets log level (trace, debug, info, warn, error, fatal, panic, disabled), (also via %s)", loglevelEnv))
		logErrorStack = fs.Bool("log-error-stack", true, fmt.Sprintf("if true, log full error stacktrace, else just log error, (also via %s)", logErrorStackEnv))
		port          = fs.Int("port", 8080, fmt.Sprintf("listen port for server (also via %s)", portEnv))
		dbhost        = fs.String("db-host", "", fmt.Sprintf("postgresql database host (also via %s)", dbHostEnv))
		dbport        = fs.Int("db-port", 5432, fmt.Sprintf("postgresql database port (also via %s)", dbPortEnv))
		dbname        = fs.String("db-name", "", fmt.Sprintf("postgresql database name (also via %s)", dbNameEnv))
		dbuser        = fs.String("db-user", "", fmt.Sprintf("postgresql database user (also via %s)", dbUserEnv))
		dbpassword    = fs.String("db-password", "", fmt.Sprintf("postgresql database password (also via %s)", dbPasswordEnv))
	)

	// Parse the command line flags from above
	err = ff.Parse(fs, args[1:], ff.WithEnvVarNoPrefix())
	if err != nil {
		return flgs, err
	}

	return flags{
		loglvl:        *loglvl,
		logLvlMin:     *logLvlMin,
		logErrorStack: *logErrorStack,
		port:          *port,
		dbhost:        *dbhost,
		dbport:        *dbport,
		dbname:        *dbname,
		dbuser:        *dbuser,
		dbpassword:    *dbpassword,
	}, nil
}

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error from main.run(): %s\n", err)
		os.Exit(exitFail)
	}
}

func run(args []string) error {

	flgs, err := newFlags(args)
	if err != nil {
		return err
	}

	// determine minimum logging level based on flag input
	minlvl, err := zerolog.ParseLevel(flgs.logLvlMin)
	if err != nil {
		return err
	}

	// determine logging level based on flag input
	lvl, err := zerolog.ParseLevel(flgs.loglvl)
	if err != nil {
		return err
	}

	// setup logger with appropriate defaults
	lgr := logger.NewLogger(os.Stdout, minlvl, true)

	// logs will be written at the level set in NewLogger (which is
	// also the minimum level). If the logs are to be written at a
	// different level than the minimum, use SetGlobalLevel to set
	// the global logging level to that. Minimum rules will still
	// apply.
	if minlvl != lvl {
		zerolog.SetGlobalLevel(lvl)
	}

	// set global logging time field format to Unix timestamp
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	lgr.Info().Msgf("minimum accepted logging level set to %s", minlvl)
	lgr.Info().Msgf("logging level set to %s", lvl)

	// set global to log errors with stack (or not) based on flag
	logger.WriteErrorStackGlobal(flgs.logErrorStack)
	lgr.Info().Msgf("log error stack global set to %t", flgs.logErrorStack)

	// validate port in acceptable range
	err = portRange(flgs.port)
	if err != nil {
		lgr.Fatal().Err(err).Msg("portRange() error")
	}

	// initialize Gorilla mux router with /api subroute
	mr := app.NewMuxRouter()

	// initialize server driver
	serverDriver := app.NewDriver()

	// initialize server configuration parameters
	params := app.NewServerParams(lgr, serverDriver)

	// initialize Server
	s, err := app.NewServer(mr, params)
	if err != nil {
		lgr.Fatal().Err(err).Msg("Error from app.NewServer")
	}

	// set listener address
	s.Addr = fmt.Sprintf(":%d", flgs.port)

	// initialize auth structs
	s.AccessTokenConverter = authgateway.GoogleAccessTokenConverter{}
	s.Authorizer = auth.Authorizer{}

	// initialize struct with PostgreSQL datasource name details
	dsn := datastore.NewPostgreSQLDSN(flgs.dbhost, flgs.dbname, flgs.dbuser, flgs.dbpassword, flgs.dbport)

	// initialize PostgreSQL database
	db, cleanup, err := datastore.NewPostgreSQLDB(dsn, lgr)
	if err != nil {
		lgr.Fatal().Err(err).Msg("Error from datastore.NewDB")
	}
	defer cleanup()

	// initialize Datastore
	pgDatastore := datastore.NewDatastore(db)

	// initialize services

	pinger := pingstore.NewPinger(pgDatastore)
	s.PingService = service.NewPingService(pinger)

	s.LoggerService = service.NewLoggerService(lgr)

	randomStringGenerator := random.StringGenerator{}
	movieTransactor := moviestore.NewTransactor(pgDatastore)
	movieSelector := moviestore.NewSelector(pgDatastore)

	s.CreateMovieService = service.NewCreateMovieService(randomStringGenerator, movieTransactor)
	s.UpdateMovieService = service.NewUpdateMovieService(movieTransactor)
	s.DeleteMovieService = service.NewDeleteMovieService(movieSelector, movieTransactor)
	s.FindMovieService = service.NewFindMovieService(movieSelector)

	return s.ListenAndServe()
}

// portRange validates the port be in an acceptable range
func portRange(port int) error {
	if port < 0 || port > 65535 {
		return errs.E(fmt.Sprintf("port %d is not within valid port range (0 to 65535)", port))
	}
	return nil
}
