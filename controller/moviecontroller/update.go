package moviecontroller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gilcrest/go-api-basic/domain/movie"
	"github.com/gilcrest/go-api-basic/domain/user"

	"github.com/gorilla/mux"

	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/gateway/authgateway"
	"golang.org/x/oauth2"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"github.com/gilcrest/go-api-basic/controller"
	"github.com/gilcrest/go-api-basic/datastore/moviestore"
)

// updateMovieRequestBody is the request struct for Update
type updateMovieRequestBody struct {
	Title    string `json:"title"`
	Rated    string `json:"rated"`
	Released string `json:"release_date"`
	RunTime  int    `json:"run_time"`
	Director string `json:"director"`
	Writer   string `json:"writer"`
}

// newCreateMovieResponseBody is an initializer for createMovieResponseBody
func newUpdateMovieResponse(m *movie.Movie) *UpdateMovieResponse {
	return &UpdateMovieResponse{
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

// UpdateMovieResponse is the response struct for a Movie
type UpdateMovieResponse struct {
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

// newMovie4Update is an initializer for the Movie struct for the
// update operation
func newMovie4Update(rb *updateMovieRequestBody, externalID string, u *user.User) (*movie.Movie, error) {
	var (
		m   = &movie.Movie{}
		err error
	)
	m.SetExternalID(externalID)
	m.SetTitle(rb.Title)
	m.SetRated(rb.Rated)
	m, err = m.SetReleased(rb.Released)
	if err != nil {
		return nil, err
	}
	m.SetRunTime(rb.RunTime)
	m.SetDirector(rb.Director)
	m.SetWriter(rb.Writer)
	m.SetUpdateUser(u)
	m.SetUpdateTime()

	return m, nil
}

// Update updates the movie given the external id sent in
func (ctl *MovieController) Update(r *http.Request) (*controller.StandardResponse, error) {
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

	// Declare rb as an instance of updateMovieRequestBody
	rb := new(updateMovieRequestBody)

	// Decode JSON HTTP request body into a Decoder type
	// and unmarshal that into requestData
	err = json.NewDecoder(r.Body).Decode(&rb)
	defer r.Body.Close()
	// Call DecoderErr to determine if body is nil, json is malformed
	// or any other error
	err = controller.DecoderErr(err)
	if err != nil {
		return nil, err
	}

	// Convert request into a Movie struct
	m, err := newMovie4Update(rb, extlid, u)
	if err != nil {
		return nil, err
	}

	// Begin a DB Tx, if the underlying struct is a MockDatastore then
	// the Tx will be nil
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

	// Call the Update method of the Transactor to update data on
	// the database (unless mocked, of course). If an error occurs,
	// rollback the transaction
	err = movieTransactor.Update(ctx, m)
	if err != nil {
		return nil, ctl.App.Datastorer.RollbackTx(tx, err)
	}

	// Commit the Transaction
	if err := ctl.App.Datastorer.CommitTx(tx); err != nil {
		return nil, err
	}

	// Populate the response
	response, err := controller.NewStandardResponse(r, newUpdateMovieResponse(m))
	if err != nil {
		return nil, err
	}

	return response, nil
}
