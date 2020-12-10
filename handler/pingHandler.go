package handler

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/hlog"

	"github.com/gilcrest/go-api-basic/controller/pingcontroller"
	"github.com/gilcrest/go-api-basic/domain/errs"
)

// Ping handles GET requests for the /ping endpoint
func (ah *AppHandler) Ping(w http.ResponseWriter, r *http.Request) {
	logger := *hlog.FromRequest(r)

	// Initialize SearchController
	sc := pingcontroller.NewPingController(ah.App)

	// Send the request context and request struct to the controller
	// Receive a response or error in return
	response, err := sc.Ping(r)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
		return
	}
}
