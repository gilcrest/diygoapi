package service

import (
	"context"

	"github.com/gilcrest/diy-go-api/datastore/pingstore"

	"github.com/rs/zerolog"
)

// PingResponse is the response struct for the PingService
type PingResponse struct {
	DBUp bool `json:"db_up"`
}

// PingService pings the database.
type PingService struct {
	Datastorer Datastorer
}

// Ping method pings the database
func (p PingService) Ping(ctx context.Context, lgr zerolog.Logger) PingResponse {
	err := pingstore.PingDB(ctx, p.Datastorer.Pool())
	if err != nil {
		// if error from PingDB, log the error, set dbok to false
		lgr.Error().Stack().Err(err).Msg("PingDB error")
		return PingResponse{DBUp: false}
	}

	return PingResponse{DBUp: true}
}
