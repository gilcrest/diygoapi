package moviecontroller

import (
	"github.com/gilcrest/go-api-basic/app"
)

// NewMovieController initializes MovieController
func NewMovieController(app *app.Application) *MovieController {
	return &MovieController{App: app}
}

// MovieController is used as the base controller for the Movie logic
type MovieController struct {
	App *app.Application
}

//// ListMovieResponse is the response struct for multiple Movies
//type ListMovieResponse struct {
//	controller.StandardResponseFields
//	Data []*ResponseData `json:"data"`
//}
//
//// SingleMovieResponse is the response struct for multiple Movies
//type SingleMovieResponse struct {
//	controller.StandardResponseFields
//	Data *ResponseData `json:"data"`
//}
//
//// DeleteMovieResponse is the response struct for deleted Movies
//type DeleteMovieResponse struct {
//	controller.StandardResponseFields
//	Data struct {
//		ExtlID  string `json:"extl_id"`
//		Deleted bool   `json:"deleted"`
//	} `json:"data"`
//}
//
//func newDeleteMovieResponse(m *movie.Movie, srf controller.StandardResponseFields) *DeleteMovieResponse {
//	return &DeleteMovieResponse{
//		StandardResponseFields: srf,
//		Data: struct {
//			ExtlID  string "json:\"extl_id\""
//			Deleted bool   "json:\"deleted\""
//		}{
//			ExtlID:  m.ExternalID,
//			Deleted: true,
//		},
//	}
//}

//// Delete removes the movie given the id sent in
//func (ctl *MovieController) Delete(ctx context.Context, id string, token string) (*DeleteMovieResponse, error) {
//	// authorize and get user from token
//	u, err := authcontroller.AuthorizeAccessToken(ctx, ctl.App, token)
//	if err != nil {
//		return nil, err
//	}
//
//	// TODO something to properly authorize Delete
//	ctl.App.Logger.Info().
//		Str("email", u.Email).
//		Str("first name", u.FirstName).
//		Str("last name", u.LastName).
//		Str("full name", u.FullName).
//		Msgf("Delete authorized for %s", u.Email)
//
//	// declare variable as the Transactor interface
//	var movieSelector moviestore.Selector
//
//	// If app is in Mock mode, use MockDB to satisfy the interface,
//	// otherwise use a true sql.DB for moviestore.DB
//	if ctl.App.Mock {
//		movieSelector = moviestore.NewMockDB()
//	} else {
//		movieSelector, err = moviestore.NewDB(ctl.App.Datastorer.DB())
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	// Find the Movie by ID using the selector.FindByID method
//	m, err := movieSelector.FindByID(ctx, id)
//	if err != nil {
//		return nil, err
//	}
//
//	// start a new database transaction
//	tx, err := ctl.App.Datastorer.BeginTx(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	// declare variable as the Transactor interface
//	var movieTransactor moviestore.Transactor
//
//	// If app is in Mock mode, use MockTx to satisfy the interface,
//	// otherwise use a true sql.Tx for moviestore.Tx
//	if ctl.App.Mock {
//		movieTransactor = moviestore.NewMockTx()
//	} else {
//		movieTransactor, err = moviestore.NewTx(tx)
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	// Delete method of Transactor physically deletes the record
//	// from the DB, unless mocked
//	err = movieTransactor.Delete(ctx, m)
//	if err != nil {
//		return nil, ctl.App.Datastorer.RollbackTx(tx, err)
//	}
//
//	// Commit the Transaction
//	if err := ctl.App.Datastorer.CommitTx(tx); err != nil {
//		return nil, err
//	}
//
//	// Populate the response
//	response := newDeleteMovieResponse(m, ctl.SRF)
//
//	return response, nil
//}
//
//// FindByID finds a movie given its' unique ID
//func (ctl *MovieController) FindByID(ctx context.Context, id string, token string) (*SingleMovieResponse, error) {
//	// authorize and get user from token
//	u, err := authcontroller.AuthorizeAccessToken(ctx, ctl.App, token)
//	if err != nil {
//		return nil, err
//	}
//
//	// TODO something to properly authorize FindByID
//	ctl.App.Logger.Info().
//		Str("email", u.Email).
//		Str("first name", u.FirstName).
//		Str("last name", u.LastName).
//		Str("full name", u.FullName).
//		Msgf("Delete authorized for %s", u.Email)
//
//	// declare variable as the Transactor interface
//	var movieSelector moviestore.Selector
//
//	// If app is in Mock mode, use MockDB to satisfy the interface,
//	// otherwise use a true sql.DB for moviestore.DB
//	if ctl.App.Mock {
//		movieSelector = moviestore.NewMockDB()
//	} else {
//		movieSelector, err = moviestore.NewDB(ctl.App.Datastorer.DB())
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	// Find the Movie by ID using the selector.FindByID method
//	m, err := movieSelector.FindByID(ctx, id)
//	if err != nil {
//		return nil, err
//	}
//
//	rd, err := newMovieResponse(m)
//	if err != nil {
//		return nil, err
//	}
//
//	// Populate the response
//	response := ctl.NewSingleMovieResponse(rd)
//
//	return response, nil
//}
//
//// FindAll finds the entire set of Movies
//func (ctl *MovieController) FindAll(ctx context.Context, token string) (*ListMovieResponse, error) {
//	// authorize and get user from token
//	u, err := authcontroller.AuthorizeAccessToken(ctx, ctl.App, token)
//	if err != nil {
//		return nil, err
//	}
//
//	// TODO something to properly authorize FindByID
//	ctl.App.Logger.Info().
//		Str("email", u.Email).
//		Str("first name", u.FirstName).
//		Str("last name", u.LastName).
//		Str("full name", u.FullName).
//		Msgf("Delete authorized for %s", u.Email)
//
//	// declare variable as the Transactor interface
//	var movieSelector moviestore.Selector
//
//	// If app is in Mock mode, use MockDB to satisfy the interface,
//	// otherwise use a true sql.DB for moviestore.DB
//	if ctl.App.Mock {
//		movieSelector = moviestore.NewMockDB()
//	} else {
//		movieSelector, err = moviestore.NewDB(ctl.App.Datastorer.DB())
//		if err != nil {
//			return nil, err
//		}
//	}
//
//	// Find the list of all Movies using the selector.FindAll method
//	movies, err := movieSelector.FindAll(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	// Populate the response
//	response, err := ctl.NewListMovieResponse(movies)
//	if err != nil {
//		return nil, err
//	}
//
//	return response, nil
//}
//
//// NewListMovieResponse is an initializer for ListMovieResponse
//func (ctl *MovieController) NewListMovieResponse(ms []*movie.Movie) (*ListMovieResponse, error) {
//	var s []*ResponseData
//
//	for _, m := range ms {
//		mr, err := newMovieResponse(m)
//		if err != nil {
//			return nil, err
//		}
//		s = append(s, mr)
//	}
//
//	return &ListMovieResponse{StandardResponseFields: ctl.SRF, Data: s}, nil
//}
//
//// NewSingleMovieResponse is an initializer for SingleMovieResponse
//func (ctl *MovieController) NewSingleMovieResponse(mr *ResponseData) *SingleMovieResponse {
//	return &SingleMovieResponse{StandardResponseFields: ctl.SRF, Data: mr}
//}
//
//
