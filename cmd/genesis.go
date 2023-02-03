package cmd

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
	"golang.org/x/text/language"

	"github.com/gilcrest/diygoapi"
	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/gateway"
	"github.com/gilcrest/diygoapi/logger"
	"github.com/gilcrest/diygoapi/secure"
	"github.com/gilcrest/diygoapi/service"
	"github.com/gilcrest/diygoapi/sqldb"
)

// Genesis command runs the Genesis service and seeds the database.
func Genesis() (err error) {
	const op errs.Op = "cmd/Genesis"

	var (
		flgs        flags
		minlvl, lvl zerolog.Level
		ek          *[32]byte
	)

	// newFlags will retrieve the database info from the environment using ff
	flgs, err = newFlags([]string{"server"})
	if err != nil {
		return errs.E(op, err)
	}

	// determine minimum logging level based on flag input
	minlvl, err = zerolog.ParseLevel(flgs.logLvlMin)
	if err != nil {
		return errs.E(op, err)
	}

	// determine logging level based on flag input
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
	lgr.Info().Msgf("log error stack global set to %t", flgs.logErrorStack)

	if flgs.encryptkey == "" {
		lgr.Fatal().Msg("no encryption key found")
	}

	// decode and retrieve encryption key
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

	var supportedLangs = []language.Tag{
		language.AmericanEnglish,
	}

	matcher := language.NewMatcher(supportedLangs)

	s := service.GenesisService{
		Datastorer:      sqldb.NewDB(dbpool),
		APIKeyGenerator: secure.RandomGenerator{},
		EncryptionKey:   ek,
		TokenExchanger:  gateway.Oauth2TokenExchange{},
		LanguageMatcher: matcher,
	}

	var b []byte
	b, err = os.ReadFile(genesisRequestFile)
	if err != nil {
		return errs.E(op, err)
	}
	f := diygoapi.GenesisRequest{}
	err = json.Unmarshal(b, &f)
	if err != nil {
		return errs.E(op, err)
	}

	var response diygoapi.GenesisResponse
	response, err = s.Arche(ctx, &f)
	if err != nil {
		var e *errs.Error
		if errors.As(err, &e) {
			lgr.Error().Stack().Err(e.Err).
				Str("Kind", e.Kind.String()).
				Str("Parameter", string(e.Param)).
				Str("Code", string(e.Code)).
				Msg("Error Response Sent")
			return errs.E(op, err)
		} else {
			lgr.Error().Err(err).Send()
			return errs.E(op, err)
		}
	}

	var responseJSON []byte
	responseJSON, err = json.MarshalIndent(response, "", "  ")
	if err != nil {
		return errs.E(op, err)
	}

	err = os.WriteFile(service.LocalJSONGenesisResponseFile, responseJSON, 0644)
	if err != nil {
		return errs.E(op, err)
	}

	fmt.Println(string(responseJSON))

	return nil
}

// NewEncryptionKey generates a random 256-bit key and prints it to standard out.
// It will return an error if the system's secure random number generator fails
// to function correctly, in which case the caller should not continue.
// Taken from https://github.com/gtank/cryptopasta/blob/master/encrypt.go
func NewEncryptionKey() {
	lgr := logger.NewWithGCPHook(os.Stdout, zerolog.DebugLevel, true)

	keyBytes, err := secure.NewEncryptionKey()
	if err != nil {
		lgr.Fatal().Err(err).Msg("EncryptionKey() error")
	}

	fmt.Printf("Key Ciphertext:\t[%s]\n", hex.EncodeToString(keyBytes[:]))
}
