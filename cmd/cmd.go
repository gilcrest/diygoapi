package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/peterbourgon/ff/v3"
	"github.com/rs/zerolog"
	"golang.org/x/text/language"

	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/gateway"
	"github.com/gilcrest/diygoapi/logger"
	"github.com/gilcrest/diygoapi/secure"
	"github.com/gilcrest/diygoapi/server"
	"github.com/gilcrest/diygoapi/service"
	"github.com/gilcrest/diygoapi/sqldb"
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
	// encryption key environment variable name
	encryptKeyEnv string = "ENCRYPT_KEY"
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
	const op errs.Op = "cmd/newFlags"
	// create new FlagSet using the program name being executed (args[0])
	// as the name of the FlagSet
	fs := flag.NewFlagSet(args[0], flag.ContinueOnError)
	var (
		logLvlMin     = fs.String("log-level-min", "trace", fmt.Sprintf("sets minimum log level (trace, debug, info, warn, error, fatal, panic, disabled), (also via %s)", logLevelMinEnv))
		loglvl        = fs.String("log-level", "info", fmt.Sprintf("sets log level (trace, debug, info, warn, error, fatal, panic, disabled), (also via %s)", loglevelEnv))
		logErrorStack = fs.Bool("log-error-stack", false, fmt.Sprintf("if true, log full error stacktrace using github.com/pkg/errors, else just log error, (also via %s)", logErrorStackEnv))
		port          = fs.Int("port", 8080, fmt.Sprintf("listen port for server (also via %s)", portEnv))
		dbhost        = fs.String("db-host", "", fmt.Sprintf("postgresql database host (also via %s)", sqldb.DBHostEnv))
		dbport        = fs.Int("db-port", 5432, fmt.Sprintf("postgresql database port (also via %s)", sqldb.DBPortEnv))
		dbname        = fs.String("db-name", "", fmt.Sprintf("postgresql database name (also via %s)", sqldb.DBNameEnv))
		dbuser        = fs.String("db-user", "", fmt.Sprintf("postgresql database user (also via %s)", sqldb.DBUserEnv))
		dbpassword    = fs.String("db-password", "", fmt.Sprintf("postgresql database password (also via %s)", sqldb.DBPasswordEnv))
		dbsearchpath  = fs.String("db-search-path", "", fmt.Sprintf("postgresql database search path (also via %s)", sqldb.DBSearchPathEnv))
		encryptkey    = fs.String("encrypt-key", "", fmt.Sprintf("encryption key (also via %s)", encryptKeyEnv))
	)

	// Parse the command line flags from above
	err := ff.Parse(fs, args[1:], ff.WithEnvVars())
	if err != nil {
		return flags{}, errs.E(op, err)
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
func Run(args []string) (err error) {
	const op errs.Op = "cmd/Run"

	var flgs flags
	flgs, err = newFlags(args)
	if err != nil {
		return errs.E(op, err)
	}

	// determine minimum logging level based on flag input
	var minlvl zerolog.Level
	minlvl, err = zerolog.ParseLevel(flgs.logLvlMin)
	if err != nil {
		return errs.E(op, err)
	}

	// determine logging level based on flag input
	var lvl zerolog.Level
	lvl, err = zerolog.ParseLevel(flgs.loglvl)
	if err != nil {
		return errs.E(op, err)
	}

	// setup logger with appropriate defaults
	lgr := logger.NewWithGCPHook(os.Stdout, minlvl, true)

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
	logger.LogErrorStackViaPkgErrors(flgs.logErrorStack)
	lgr.Info().Msgf("log error stack via github.com/pkg/errors set to %t", flgs.logErrorStack)

	// validate port in acceptable range
	err = portRange(flgs.port)
	if err != nil {
		lgr.Fatal().Err(err).Msg("portRange() error")
	}

	// initialize Server enfolding a http.Server with default timeouts
	// a Gorilla mux router with /api subroute and a zerolog.Logger
	s := server.New(server.NewMuxRouter(), server.NewDriver(), lgr)

	// set listener address
	s.Addr = fmt.Sprintf(":%d", flgs.port)

	if flgs.encryptkey == "" {
		lgr.Fatal().Msg("no encryption key found")
	}

	// decode and retrieve encryption key
	var ek *[32]byte
	ek, err = secure.ParseEncryptionKey(flgs.encryptkey)
	if err != nil {
		lgr.Fatal().Err(err).Msg("secure.ParseEncryptionKey() error")
	}

	ctx := context.Background()

	// initialize PostgreSQL database
	var (
		dbpool  *pgxpool.Pool
		cleanup func()
	)
	dbpool, cleanup, err = sqldb.NewPostgreSQLPool(ctx, lgr, newPostgreSQLDSN(flgs))
	if err != nil {
		lgr.Fatal().Err(err).Msg("sqldb.NewPostgreSQLPool error")
	}
	defer cleanup()

	// create a new DB using the pool and established connection
	db := sqldb.NewDB(dbpool)

	err = db.ValidatePool(ctx, lgr)
	if err != nil {
		lgr.Fatal().Err(err).Msg("db.ValidatePool error")
	}

	var supportedLangs = []language.Tag{
		language.AmericanEnglish,
	}

	matcher := language.NewMatcher(supportedLangs)

	s.Services = server.Services{
		OrgServicer: &service.OrgService{
			Datastorer:      db,
			APIKeyGenerator: secure.RandomGenerator{},
			EncryptionKey:   ek},
		AppServicer: &service.AppService{
			Datastorer:      db,
			APIKeyGenerator: secure.RandomGenerator{},
			EncryptionKey:   ek},
		PingService:   &service.PingService{Datastorer: db},
		LoggerService: &service.LoggerService{Logger: lgr},
		GenesisServicer: &service.GenesisService{
			Datastorer:      db,
			APIKeyGenerator: secure.RandomGenerator{},
			EncryptionKey:   ek,
			TokenExchanger:  gateway.Oauth2TokenExchange{},
			LanguageMatcher: matcher,
		},
		AuthenticationServicer: service.DBAuthenticationService{
			Datastorer:      db,
			TokenExchanger:  gateway.Oauth2TokenExchange{},
			EncryptionKey:   ek,
			LanguageMatcher: matcher,
		},
		AuthorizationServicer: &service.DBAuthorizationService{Datastorer: db},
		PermissionServicer:    &service.PermissionService{Datastorer: db},
		RoleServicer:          &service.RoleService{Datastorer: db},
		MovieServicer:         &service.MovieService{Datastorer: db},
	}

	return s.ListenAndServe()
}

// newPostgreSQLDSN initializes a sqldb.PostgreSQLDSN given a Flags struct
func newPostgreSQLDSN(flgs flags) sqldb.PostgreSQLDSN {
	return sqldb.PostgreSQLDSN{
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
	const op errs.Op = "cmd/portRange"

	if port < 0 || port > 65535 {
		return errs.E(op, fmt.Sprintf("port %d is not within valid port range (0 to 65535)", port))
	}
	return nil
}
