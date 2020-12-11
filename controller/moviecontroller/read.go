package moviecontroller

import (
	"net/http"
	"time"

	"github.com/gilcrest/go-api-basic/domain/movie"

	"github.com/gorilla/mux"

	"github.com/gilcrest/go-api-basic/controller"

	"github.com/gilcrest/go-api-basic/datastore/moviestore"
	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/gateway/authgateway"
	"golang.org/x/oauth2"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

// neMovieResponse is an initializer for createMovieResponseBody
func newMovieResponse(m *movie.Movie) *MovieResponse {
	return &MovieResponse{
		ExternalID:      m.ExternalID,
		Title:           m.Title,
		Rated:           m.Rated,
		Released:        m.Released.Format(time.RFC3339),
		RunTime:         m.RunTime,
		Director:        m.Director,
		Writer:          m.Writer,
		CreateUsername:  m.CreateUser.Email,
		CreateTimestamp: m.CreateTime.Format(time.RFC3339),
		UpdateUsername:  m.UpdateUser.Email,
		UpdateTimestamp: m.UpdateTime.Format(time.RFC3339),
	}
}

// MovieResponse is the response struct for a Movie
type MovieResponse struct {
	ExternalID      string `json:"external_id"`
	Title           string `json:"title"`
	Rated           string `json:"rated"`
	Released        string `json:"release_date"`
	RunTime         int    `json:"run_time"`
	Director        string `json:"director"`
	Writer          string `json:"writer"`
	CreateUsername  string `json:"create_username"`
	CreateTimestamp string `json:"create_timestamp"`
	UpdateUsername  string `json:"update_username"`
	UpdateTimestamp string `json:"update_timestamp"`
}

// FindByID finds a movie given its' unique ID
func (ctl *MovieController) FindByID(r *http.Request) (*controller.StandardResponse, error) {
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

	// declare variable as the Transactor interface
	var movieSelector moviestore.Selector
	movieSelector, err = moviestore.NewDB(ctl.App.Datastorer.DB())
	if err != nil {
		return nil, err
	}

	// gorilla mux Vars function returns the route variables for the
	// current request, if any. id is the external id given for the
	// movie
	vars := mux.Vars(r)
	extlid := vars["id"]

	// Find the Movie by ID using the selector.FindByID method
	m, err := movieSelector.FindByID(ctx, extlid)
	if err != nil {
		return nil, err
	}

	// Populate the response
	response, err := controller.NewStandardResponse(r, newMovieResponse(m))
	if err != nil {
		return nil, err
	}

	return response, nil
}

// FindAll finds the entire set of Movies
func (ctl *MovieController) FindAll(r *http.Request) (*controller.StandardResponse, error) {
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

	// declare variable as the Transactor interface
	var movieSelector moviestore.Selector
	movieSelector, err = moviestore.NewDB(ctl.App.Datastorer.DB())
	if err != nil {
		return nil, err
	}

	// Find the list of all Movies using the selector.FindAll method
	movies, err := movieSelector.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var smr []*MovieResponse
	for _, m := range movies {
		mr := newMovieResponse(m)
		smr = append(smr, mr)
	}

	// Populate the response
	response, err := controller.NewStandardResponse(r, smr)
	if err != nil {
		return nil, err
	}

	return response, nil
}
