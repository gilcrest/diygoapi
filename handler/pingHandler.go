package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gilcrest/go-api-basic/datastore/pingstore"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/rs/zerolog/hlog"
)

// PingHandler is a Handler that gives app status, such as db ping, etc.
type PingHandler http.Handler

// NewPingHandler is a provider for PingHandler
func NewPingHandler(h DefaultPingHandler) PingHandler {
	return http.HandlerFunc(h.Ping)
}

// DefaultPingHandler is a handler to allow for general health checks
type DefaultPingHandler struct {
	Pinger pingstore.Pinger
}

// Ping handles GET requests for the /ping endpoint
func (h DefaultPingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	type pingResponseData struct {
		DBUp bool `json:"db_up"`
	}

	// pull logger from request context
	logger := *hlog.FromRequest(r)

	// pull the context from the http request
	ctx := r.Context()

	dbok := true
	err := h.Pinger.PingDB(ctx)
	if err != nil {
		// if error from PingDB, log the error, set dbok to false
		logger.Error().Err(err).Msg("PingDB error")
		dbok = false
	}

	response := pingResponseData{DBUp: dbok}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
		return
	}
}
