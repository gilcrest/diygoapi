package moviecontroller

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
