package moviecontroller

import (
	"net/http"

	"github.com/gilcrest/go-api-basic/controller"

	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/gateway/authgateway"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"github.com/gilcrest/go-api-basic/datastore/moviestore"
	"github.com/gilcrest/go-api-basic/domain/movie"
)

// DeleteMovieResponse is the response struct for deleted Movies
type DeleteMovieResponse struct {
	ExtlID  string `json:"extl_id"`
	Deleted bool   `json:"deleted"`
}

func newDeleteMovieResponse(m *movie.Movie) *DeleteMovieResponse {
	return &DeleteMovieResponse{
		ExtlID:  m.ExternalID,
		Deleted: true,
	}
}

// Delete removes the movie given the id sent in
func (ctl *MovieController) Delete(r *http.Request) (*controller.StandardResponse, error) {
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

	// gorilla mux Vars function returns the route variables for the
	// current request, if any. id is the external id given for the
	// movie
	vars := mux.Vars(r)
	extlid := vars["id"]

	// declare variable as the Transactor interface
	var movieSelector moviestore.Selector
	movieSelector, err = moviestore.NewDB(ctl.App.Datastorer.DB())
	if err != nil {
		return nil, err
	}

	// Find the Movie by ID using the selector.FindByID method
	m, err := movieSelector.FindByID(ctx, extlid)
	if err != nil {
		return nil, err
	}

	// start a new database transaction
	tx, err := ctl.App.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	// declare variable as the Transactor interface
	var movieTransactor moviestore.Transactor
	movieTransactor, err = moviestore.NewTx(tx)
	if err != nil {
		return nil, err
	}

	// Delete method of Transactor physically deletes the record
	// from the DB, unless mocked
	err = movieTransactor.Delete(ctx, m)
	if err != nil {
		return nil, ctl.App.Datastorer.RollbackTx(tx, err)
	}

	// Commit the Transaction
	if err := ctl.App.Datastorer.CommitTx(tx); err != nil {
		return nil, err
	}

	// Populate the response
	response, err := controller.NewStandardResponse(r, newDeleteMovieResponse(m))
	if err != nil {
		return nil, err
	}

	return response, nil
}
