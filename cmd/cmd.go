package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
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

	err = flgs.Validate()
	if err != nil {
		lgr.Fatal().Err(err).Msg("flags.Validate() error")
	}

	// decode and retrieve encryption key
	var ek *[32]byte
	ek, err = secure.ParseEncryptionKey(flgs.encryptkey)
	if err != nil {
		lgr.Fatal().Err(err).Msg("secure.ParseEncryptionKey() error")
	}

	// initialize Server enfolding a http.Server with default timeouts,
	// a mux router and a zerolog.Logger
	s := server.New(http.NewServeMux(), server.NewDriver(), lgr)

	// set Server listener address
	s.Addr = fmt.Sprintf(":%d", flgs.port)

	ctx := context.Background()

	// initialize PostgreSQL database
	var (
		dbpool  *pgxpool.Pool
		cleanup func()
	)
	dbpool, cleanup, err = sqldb.NewPgxPool(ctx, lgr, newPostgreSQLDSN(flgs))
	if err != nil {
		lgr.Fatal().Err(err).Msg("sqldb.NewPgxPool error")
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

// printFlags prints the flags to stdout
func printFlags(flgs flags) {
	fmt.Printf("Log Level: %s\n", flgs.loglvl)
	fmt.Printf("Log Level Min: %s\n", flgs.logLvlMin)
	fmt.Printf("Log Error Stack: %t\n", flgs.logErrorStack)
	fmt.Printf("Port: %d\n", flgs.port)
	fmt.Printf("DB Host: %s\n", flgs.dbhost)
	fmt.Printf("DB Port: %d\n", flgs.dbport)
	fmt.Printf("DB Name: %s\n", flgs.dbname)
	fmt.Printf("DB Search Path: %s\n", flgs.dbsearchpath)
	fmt.Printf("DB User: %s\n", flgs.dbuser)
	if flgs.dbpassword == "" {
		fmt.Println("DB Password: <empty>")
	} else {
		fmt.Println("DB Password: <not empty>")
	}
	if flgs.encryptkey == "" {
		fmt.Println("Encryption Key: <empty>")
	} else {
		fmt.Println("Encryption Key: <not empty>")
	}
}
