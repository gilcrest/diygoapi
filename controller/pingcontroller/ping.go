package pingcontroller

import (
	"net/http"

	"github.com/gilcrest/go-api-basic/controller"

	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/gateway/authgateway"

	"golang.org/x/oauth2"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
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

func (ctl *PingController) newPingResponseData() *PingResponseData {
	return &PingResponseData{DBUp: true}
}

// Ping is a simple check to make sure the API is up and running
func (ctl *PingController) Ping(r *http.Request) (*controller.StandardResponse, error) {
	ctx := r.Context()

	accessToken, err := auth.FromRequest(r)
	if err != nil {
		return nil, err
	}

	oauthService, err := googleoauth2.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(accessToken.NewGoogleOauth2Token())))
	if err != nil {
		return nil, errs.E(err)
	}

	userInfo, err := oauthService.Userinfo.Get().Do()
	if err != nil {
		// "In summary, a 401 Unauthorized response should be used for missing or
		// bad authentication, and a 403 Forbidden response should be used afterwards,
		// when the user is authenticated but isn’t authorized to perform the
		// requested operation on the given resource."
		// In this case, we are getting a bad response from Google service, assume
		// they are not able to authenticate properly
		return nil, errs.E(errs.Unauthenticated, err)
	}

	u := authgateway.NewUser(userInfo)

	var authorizer auth.Authorizer = auth.Auth{}
	err = authorizer.Authorize(ctx, u, r.URL.Path, r.Method)
	if err != nil {
		return nil, err
	}

	// Populate the response
	pr := ctl.newPingResponseData()

	response, err := controller.NewStandardResponse(r, pr)
	if err != nil {
		return nil, err
	}

	return response, nil
}
