package service

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Pinger pings the database
type Pinger interface {
	PingDB(context.Context) error
}

// PingResponse is the response struct for the PingService
type PingResponse struct {
	DBUp bool `json:"db_up"`
}

// PingService pings the database.
type PingService struct {
	Pinger Pinger
}

// NewPingService is an initializer for PingService
func NewPingService(p Pinger) *PingService {
	return &PingService{Pinger: p}
}

// Ping method pings the database
func (p PingService) Ping(ctx context.Context, logger zerolog.Logger) PingResponse {
	dbok := true
	err := p.Pinger.PingDB(ctx)
	if err != nil {
		pingErr := errors.WithStack(err)
		// if error from PingDB, log the error, set dbok to false
		logger.Error().Stack().Err(pingErr).Msg("PingDB error")
		dbok = false
	}

	response := PingResponse{DBUp: dbok}

	return response
}
