// Package commands defines and implements command-line build
// commands and flags used by the application. The package name is
// inspired by Hugo and Cobra/Viper, but for now, neither of those
// are actually being used, opting instead for the simplicity of ff.
package commands

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/casbin/casbin/v2"
	"github.com/peterbourgon/ff/v3"
	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/datastore/pingstore"
	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/logger"
	"github.com/gilcrest/go-api-basic/domain/secure"
	"github.com/gilcrest/go-api-basic/gateway/authgateway"
	"github.com/gilcrest/go-api-basic/service"
)

const (
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
	// database search path environment variable name
	dbSearchPath string = "DB_SEARCH_PATH"
	// encryption key environment variable name
	encryptKey string = "ENCRYPT_KEY"
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

	// dbsearchpath is the database search path
	dbsearchpath string

	// encryptkey is the encryption key
	encryptkey string
}

// newFlags parses the command line flags using ff and returns
// a flags struct or an error
func newFlags(args []string) (flags, error) {
	// create new FlagSet using the program name being executed (args[0])
	// as the name of the FlagSet
	flagSet := flag.NewFlagSet(args[0], flag.ContinueOnError)

	var (
		logLvlMin     = flagSet.String("log-level-min", "trace", fmt.Sprintf("sets minimum log level (trace, debug, info, warn, error, fatal, panic, disabled), (also via %s)", logLevelMinEnv))
		loglvl        = flagSet.String("log-level", "info", fmt.Sprintf("sets log level (trace, debug, info, warn, error, fatal, panic, disabled), (also via %s)", loglevelEnv))
		logErrorStack = flagSet.Bool("log-error-stack", true, fmt.Sprintf("if true, log full error stacktrace, else just log error, (also via %s)", logErrorStackEnv))
		port          = flagSet.Int("port", 8080, fmt.Sprintf("listen port for server (also via %s)", portEnv))
		dbhost        = flagSet.String("db-host", "", fmt.Sprintf("postgresql database host (also via %s)", dbHostEnv))
		dbport        = flagSet.Int("db-port", 5432, fmt.Sprintf("postgresql database port (also via %s)", dbPortEnv))
		dbname        = flagSet.String("db-name", "", fmt.Sprintf("postgresql database name (also via %s)", dbNameEnv))
		dbuser        = flagSet.String("db-user", "", fmt.Sprintf("postgresql database user (also via %s)", dbUserEnv))
		dbpassword    = flagSet.String("db-password", "", fmt.Sprintf("postgresql database password (also via %s)", dbPasswordEnv))
		dbsearchpath  = flagSet.String("db-search-path", "", fmt.Sprintf("postgresql database search path (also via %s)", dbSearchPath))
		encryptkey    = flagSet.String("encrypt-key", "", fmt.Sprintf("encryption key (also via %s)", encryptKey))
	)

	// Parse the command line flags from above
	err := ff.Parse(flagSet, args[1:], ff.WithEnvVarNoPrefix())
	if err != nil {
		return flags{}, err
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
		dbsearchpath:  *dbsearchpath,
		encryptkey:    *encryptkey,
	}, nil
}

// Run parses command line flags and starts the server
func Run(args []string) error {

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
	rtr := app.NewMuxRouter()

	// initialize server driver
	serverDriver := app.NewDriver()

	// initialize server configuration parameters
	params := &app.ServerParams{
		Logger: lgr,
		Driver: serverDriver,
	}

	// initialize Server
	s := app.New(rtr, params)

	// set listener address
	s.Addr = fmt.Sprintf(":%d", flgs.port)

	if flgs.encryptkey == "" {
		lgr.Fatal().Msg("no encryption key found")
	}

	// decode and retrieve encryption key
	ek, err := secure.ParseEncryptionKey(flgs.encryptkey)
	if err != nil {
		lgr.Fatal().Err(err).Msg("secure.ParseEncryptionKey() error")
	}

	// initialize PostgreSQL database
	dbpool, cleanup, err := datastore.NewPostgreSQLPool(context.Background(), newPostgreSQLDSN(flgs), lgr)
	if err != nil {
		lgr.Fatal().Err(err).Msg("datastore.NewPostgreSQLPool error")
	}
	defer cleanup()

	// initialize Datastore
	ds := datastore.NewDatastore(dbpool)

	// initialize casbin enforcer (using config files for now, will migrate to db)
	casbinEnforcer, err := casbin.NewEnforcer("config/rbac_model.conf", "config/rbac_policy.csv")
	if err != nil {
		lgr.Fatal().Err(err).Msg("casbin.NewEnforcer error")
	}

	s.MiddlewareServices = app.MiddlewareServices{
		FindAppService:   service.FindAppService{Datastorer: ds},
		FindUserService:  service.FindUserService{GoogleOauth2TokenConverter: authgateway.GoogleOauth2TokenConverter{}, Datastorer: ds},
		AuthorizeService: service.AuthorizeService{Authorizer: auth.CasbinAuthorizer{Enforcer: casbinEnforcer}},
	}

	s.Services = app.Services{
		SeedService: service.SeedService{
			Datastorer:            ds,
			CryptoRandomGenerator: secure.CryptoRandomGenerator{},
			EncryptionKey:         ek,
		},
		PingService:      service.PingService{Pinger: pingstore.Pinger{Datastorer: ds}},
		LoggerService:    service.LoggerService{Logger: lgr},
		CreateOrgService: service.CreateOrgService{Datastorer: ds},
		UpdateOrgService: service.UpdateOrgService{Datastorer: ds},
		FindOrgService:   service.FindOrgService{Datastorer: ds},
		CreateAppService: service.CreateAppService{Datastorer: ds, CryptoRandomGenerator: secure.CryptoRandomGenerator{}, EncryptionKey: ek},
	}

	return s.ListenAndServe()
}

// newPostgreSQLDSN initializes a datastore.PostgreSQLDSN given a Flags struct
func newPostgreSQLDSN(flgs flags) datastore.PostgreSQLDSN {
	return datastore.PostgreSQLDSN{
		Host:       flgs.dbhost,
		Port:       flgs.dbport,
		DBName:     flgs.dbname,
		SearchPath: flgs.dbsearchpath,
		User:       flgs.dbuser,
		Password:   flgs.dbpassword,
	}
}

// portRange validates the port be in an acceptable range
func portRange(port int) error {
	if port < 0 || port > 65535 {
		return errs.E(fmt.Sprintf("port %d is not within valid port range (0 to 65535)", port))
	}
	return nil
}
