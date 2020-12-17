package pingcontroller

import (
	"net/http"

	"github.com/gilcrest/go-api-basic/controller"
	"github.com/rs/zerolog/hlog"

	"github.com/gilcrest/go-api-basic/app"
)

// NewPingController initializes SearchController
func NewPingController(app *app.Application) *PingController {
	return &PingController{App: app}
}

// PingController is used as the base controller for the Ping logic
type PingController struct {
	App *app.Application
}

// PingResponseData is the response struct for the ping service
type PingResponseData struct {
	DBUp bool `json:"db_up"`
}

func newPingResponseData(dbup bool) *PingResponseData {
	return &PingResponseData{DBUp: dbup}
}

// Ping is a simple check to make sure the API is up and running
func (ctl *PingController) Ping(r *http.Request) (*controller.StandardResponse, error) {
	// pull logger from request context
	logger := hlog.FromRequest(r)
	// pull the context from the http request
	ctx := r.Context()

	// check if db connection is still alive
	var dbok bool
	err := ctl.App.Datastorer.DB().PingContext(ctx)
	// if there is no error, db is up, set dbok to true,
	// if db is down, log the error
	if err != nil {
		logger.Error().Err(err).Msg("PingContext returned an error")
	} else {
		dbok = true
	}

	// if db is down, respond as such
	if !dbok {
		response, err := controller.NewStandardResponse(r, newPingResponseData(false))
		if err != nil {
			return nil, err
		}
		return response, nil
	}

	// if db is up, respond as such
	response, err := controller.NewStandardResponse(r, newPingResponseData(true))
	if err != nil {
		return nil, err
	}

	return response, nil
}
