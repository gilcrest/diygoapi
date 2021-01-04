package moviecontroller

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gilcrest/go-api-basic/app"

	"github.com/gilcrest/go-api-basic/domain/random"
	"github.com/google/uuid"

	"github.com/gilcrest/go-api-basic/gateway/authgateway"

	"github.com/gilcrest/go-api-basic/controller"
	"github.com/gilcrest/go-api-basic/datastore/moviestore"
	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/gilcrest/go-api-basic/domain/movie"

	"golang.org/x/oauth2"
	googleoauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

// NewMovieController initializes MovieController
func NewMovieController(app *app.Application) *MovieController {
	return &MovieController{App: app}
}

// MovieController is used as the base controller for the Movie logic
type MovieController struct {
	App *app.Application
}

// createMovieRequestBody is the request struct for Create
type createMovieRequestBody struct {
	Title    string `json:"title"`
	Rated    string `json:"rated"`
	Released string `json:"release_date"`
	RunTime  int    `json:"run_time"`
	Director string `json:"director"`
	Writer   string `json:"writer"`
}

// CreateMovieResponse is the response struct for a Movie
type CreateMovieResponse struct {
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

// newCreateMovieResponseBody is an initializer for createMovieResponseBody
func newCreateMovieResponse(m *movie.Movie) *CreateMovieResponse {
	return &CreateMovieResponse{
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

// CreateMovie adds a movie to the datastore
func (ctl *MovieController) CreateMovie(r *http.Request) (*controller.StandardResponse, error) {
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

	// Declare requestBody as an instance of createMovieRequestBody
	rb := new(createMovieRequestBody)

	// Decode JSON HTTP request body into a Decoder type
	// and unmarshal that into the MovieRequest struct in the
	// AddMovieHandler
	err = json.NewDecoder(r.Body).Decode(&rb)
	defer r.Body.Close()
	// Call DecoderErr to determine if body is nil, json is malformed
	// or any other error
	err = controller.DecoderErr(err)
	if err != nil {
		return nil, err
	}

	extlID, err := random.CryptoString(15)
	if err != nil {
		return nil, err
	}

	// Call the Add method to perform domain business logic
	m, err := movie.NewMovie(uuid.New(), extlID, u)
	if err != nil {
		return nil, err
	}

	m.SetTitle(rb.Title)
	m.SetRated(rb.Rated)
	m, err = m.SetReleased(rb.Released)
	if err != nil {
		return nil, err
	}
	m.SetRunTime(rb.RunTime)
	m.SetDirector(rb.Director)
	m.SetWriter(rb.Writer)

	err = m.IsValid()
	if err != nil {
		return nil, err
	}

	// Begin a DB Tx
	tx, err := ctl.App.Datastorer.BeginTx(ctx)
	if err != nil {
		return nil, err
	}

	// declare variable as the Transactor interface
	var movieTransactor moviestore.Transactor
	// moviestore.Tx implements the Transactor interface
	movieTransactor, err = moviestore.NewTx(tx)
	if err != nil {
		return nil, err
	}

	// Call the Create method of the Transactor to insert data to
	// the database (unless mocked, of course). If an error occurs,
	// rollback the transaction
	err = movieTransactor.Create(ctx, m)
	if err != nil {
		return nil, ctl.App.Datastorer.RollbackTx(tx, err)
	}

	// Commit the Transaction
	if err := ctl.App.Datastorer.CommitTx(tx); err != nil {
		return nil, err
	}

	// Populate the response
	response, err := controller.NewStandardResponse(r, newCreateMovieResponse(m))
	if err != nil {
		return nil, err
	}

	return response, nil
}
