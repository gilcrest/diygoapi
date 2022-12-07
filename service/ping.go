package service

import (
	"context"

	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog"

	"github.com/gilcrest/saaswhip"
)

// PingService pings the database.
type PingService struct {
	Datastorer saaswhip.Datastorer
}

// Ping method pings the database
func (s *PingService) Ping(ctx context.Context, lgr zerolog.Logger) saaswhip.PingResponse {
	// start db txn using pgxpool
	var (
		tx  pgx.Tx
		err error
	)
	tx, err = s.Datastorer.BeginTx(ctx)
	if err != nil {
		// if error from PingDB, log the error, set DBUp to false
		lgr.Error().Stack().Err(err).Msg("PingService.Ping BeginTx error")
		return saaswhip.PingResponse{DBUp: false}
	}
	// defer transaction rollback and handle error, if any
	defer func() {
		err = s.Datastorer.RollbackTx(ctx, tx, err)
	}()

	err = s.Datastorer.Ping(ctx)
	if err != nil {
		// if error from PingDB, log the error, set DBUp to false
		lgr.Error().Stack().Err(err).Msg("s.Datastorer.Ping error")
		return saaswhip.PingResponse{DBUp: false}
	}

	return saaswhip.PingResponse{DBUp: true}
}
