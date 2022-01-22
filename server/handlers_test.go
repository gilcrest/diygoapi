package server

// TODO - these tests all need to be refactored after sqlc changes

//// MockTransactor is a mock which satisfies the moviestore.Transactor
//// interface
//type mockTransactor struct {
//	t *testing.T
//}
//
//// NewMockTransactor is an initializer for MockTransactor
//func newMockTransactor(t *testing.T) mockTransactor {
//	return mockTransactor{t: t}
//}
//
//func (mt mockTransactor) Create(ctx context.Context, m *movie.Movie) error {
//	return nil
//}
//
//func (mt mockTransactor) Update(ctx context.Context, m *movie.Movie) error {
//	return nil
//}
//
//func (mt mockTransactor) Delete(ctx context.Context, m *movie.Movie) error {
//	return nil
//}
//
//// MockSelector is a mock which satisfies the moviestore.Selector
//// interface
//type mockSelector struct {
//	t *testing.T
//}
//
//// NewMockSelector is an initializer for MockSelector
//func newMockSelector(t *testing.T) mockSelector {
//	return mockSelector{t: t}
//}
//
//// FindByID mocks finding a movie by External ID
//func (ms mockSelector) FindByID(ctx context.Context, s string) (*movie.Movie, error) {
//
//	// get test user
//	u := usertest.NewUser(ms.t)
//
//	// mock create/update timestamp
//	cuTime := time.Date(2008, 1, 8, 06, 54, 0, 0, time.UTC)
//
//	return &movie.Movie{
//		ID:         uuid.MustParse("f118f4bb-b345-4517-b463-f237630b1a07"),
//		ExternalID: "kCBqDtyAkZIfdWjRDXQG",
//		Title:      "Repo Man",
//		Rated:      "R",
//		Released:   time.Date(1984, 3, 2, 0, 0, 0, 0, time.UTC),
//		RunTime:    92,
//		Director:   "Alex Cox",
//		Writer:     "Alex Cox",
//		CreateUser: u,
//		CreateTime: cuTime,
//		UpdateUser: u,
//		UpdateTime: cuTime,
//	}, nil
//}
//
//// FindAll mocks finding multiple movies by External ID
//func (ms mockSelector) FindAll(ctx context.Context) ([]*movie.Movie, error) {
//	// get test user
//	u := usertest.NewUser(ms.t)
//
//	// mock create/update timestamp
//	cuTime := time.Date(2008, 1, 8, 06, 54, 0, 0, time.UTC)
//
//	m1 := &movie.Movie{
//		ID:         uuid.MustParse("f118f4bb-b345-4517-b463-f237630b1a07"),
//		ExternalID: "kCBqDtyAkZIfdWjRDXQG",
//		Title:      "Repo Man",
//		Rated:      "R",
//		Released:   time.Date(1984, 3, 2, 0, 0, 0, 0, time.UTC),
//		RunTime:    92,
//		Director:   "Alex Cox",
//		Writer:     "Alex Cox",
//		CreateUser: u,
//		CreateTime: cuTime,
//		UpdateUser: u,
//		UpdateTime: cuTime,
//	}
//
//	m2 := &movie.Movie{
//		ID:         uuid.MustParse("e883ebbb-c021-423b-954a-e94edb8b85b8"),
//		ExternalID: "RWn8zcaTA1gk3ybrBdQV",
//		Title:      "The Return of the Living Dead",
//		Rated:      "R",
//		Released:   time.Date(1985, 8, 16, 0, 0, 0, 0, time.UTC),
//		RunTime:    91,
//		Director:   "Dan O'Bannon",
//		Writer:     "Russell Streiner",
//		CreateUser: u,
//		CreateTime: cuTime,
//		UpdateUser: u,
//		UpdateTime: cuTime,
//	}
//
//	return []*movie.Movie{m1, m2}, nil
//}
//
//func TestHandleMovieCreate(t *testing.T) {
//	t.Run("typical", func(t *testing.T) {
//		// set environment variable NO_DB to true if you don't
//		// have database connectivity and this test will be skipped
//		if os.Getenv("NO_DB") == "true" {
//			t.Skip("skipping db dependent test")
//		}
//
//		// initialize quickest checker
//		c := qt.New(t)
//
//		// setup Server
//		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)
//		rtr := NewMuxRouter()
//		driver := NewDriver()
//		params := NewServerParams(lgr, driver)
//		s, err := NewServer(rtr, params)
//		c.Assert(err, qt.IsNil)
//		s.AccessTokenConverter = authtest.NewMockAccessTokenConverter(t)
//		s.Authorizer = auth.CasbinAuthorizer{Enforcer: casbin.NewEnforcer("../config/rbac_model.conf", "../config/rbac_policy.csv")}
//
//		// initialize Datastore
//		ds, cleanup := datastoretest.NewDatastore(t)
//
//		// defer cleanup of the database until after the test is completed
//		t.Cleanup(cleanup)
//
//		// initialize the Transactor for the moviestore
//		movieTransactor := moviestore.NewTransactor(ds)
//
//		// initialize random.StringGenerator
//		randomStringGenerator := random.StringGenerator{}
//		s.CreateMovieService = service.NewCreateMovieService(randomStringGenerator, movieTransactor)
//
//		// setup request body using anonymous struct
//		requestBody := struct {
//			Title    string `json:"title"`
//			Rated    string `json:"rated"`
//			Released string `json:"release_date"`
//			RunTime  int    `json:"run_time"`
//			Director string `json:"director"`
//			Writer   string `json:"writer"`
//		}{
//			Title:    "Repo Man",
//			Rated:    "R",
//			Released: "1984-03-02T00:00:00Z",
//			RunTime:  92,
//			Director: "Alex Cox",
//			Writer:   "Alex Cox",
//		}
//
//		// encode request body into buffer variable
//		var buf bytes.Buffer
//		err = json.NewEncoder(&buf).Encode(requestBody)
//		if err != nil {
//			t.Fatalf("Encode() error = %v", err)
//		}
//
//		// setup path
//		path := pathPrefix + moviesV1PathRoot
//
//		// form request using httptest
//		req := httptest.NewRequest(http.MethodPost, path, &buf)
//
//		// add test access token
//		req.Header.Add("Authorization", auth.BearerTokenType+" abc123def1")
//
//		// add application/JSON header to request
//		req.Header.Add(contentTypeHeaderKey, appJSONContentTypeHeaderVal)
//
//		// initialize ResponseRecorder to use with ServeHTTP as it
//		// satisfies ResponseWriter interface and records the response
//		// for testing
//		rr := httptest.NewRecorder()
//
//		// call the router ServeHTTP method to execute the request
//		// and record the response
//		s.router.ServeHTTP(rr, req)
//
//		// Assert that Response Status Code equals 200 (StatusOK)
//		c.Assert(rr.Code, qt.Equals, http.StatusOK)
//
//		// createMovieResponse is the response struct for a Movie
//		// the response struct is tucked inside the handler, so we
//		// have to recreate it here
//		type createMovieResponse struct {
//			ExternalID      string `json:"external_id"`
//			Title           string `json:"title"`
//			Rated           string `json:"rated"`
//			Released        string `json:"release_date"`
//			RunTime         int    `json:"run_time"`
//			Director        string `json:"director"`
//			Writer          string `json:"writer"`
//			CreateUsername  string `json:"create_username"`
//			CreateTimestamp string `json:"create_timestamp"`
//			UpdateUsername  string `json:"update_username"`
//			UpdateTimestamp string `json:"update_timestamp"`
//		}
//
//		// retrieve the mock User that is used for testing
//		u, _ := s.AccessTokenConverter.Convert(req.Context(), authtest.NewAccessToken(t))
//
//		// setup the expected response data
//		wantBody := createMovieResponse{
//			ExternalID:      "superRandomString",
//			Title:           "Repo Man",
//			Rated:           "R",
//			Released:        "1984-03-02T00:00:00Z",
//			RunTime:         92,
//			Director:        "Alex Cox",
//			Writer:          "Alex Cox",
//			CreateUsername:  u.Email,
//			CreateTimestamp: "",
//			UpdateUsername:  u.Email,
//			UpdateTimestamp: "",
//		}
//
//		// initialize createMovieResponse
//		gotBody := createMovieResponse{}
//
//		// decode the response body into gotBody
//		err = decoderErr(json.NewDecoder(rr.Result().Body).Decode(&gotBody))
//		defer rr.Result().Body.Close()
//
//		// Assert that there is no error after decoding the response body
//		c.Assert(err, qt.IsNil)
//
//		// quicktest uses Google's cmp library for DeepEqual comparisons. It
//		// has some great options included with it. Below is an example of
//		// ignoring certain fields...
//		ignoreFields := cmpopts.IgnoreFields(createMovieResponse{},
//			"ExternalID", "CreateTimestamp", "UpdateTimestamp")
//
//		// Assert that the response body (gotBody) is as expected (wantBody).
//		// The External ID needs to be unique as the database unique index
//		// requires it. As a result, the ExternalID field is ignored as part
//		// of the comparison. The Create/Update timestamps are ignored as
//		// well, as they are always unique.
//		// I could put another interface into the domain logic to solve
//		// for the timestamps and may do so later, but it's probably not
//		// necessary
//		c.Assert(gotBody, qt.CmpEquals(ignoreFields), wantBody)
//	})
//
//	t.Run("mock DB", func(t *testing.T) {
//		// initialize quickest checker
//		c := qt.New(t)
//
//		// setup Server
//		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)
//		rtr := NewMuxRouter()
//		driver := NewDriver()
//		params := NewServerParams(lgr, driver)
//		s, err := NewServer(rtr, params)
//		c.Assert(err, qt.IsNil)
//		s.AccessTokenConverter = authtest.NewMockAccessTokenConverter(t)
//		s.Authorizer = auth.CasbinAuthorizer{Enforcer: casbin.NewEnforcer("../config/rbac_model.conf", "../config/rbac_policy.csv")}
//
//		// initialize a mock Transactor
//		movieTransactor := newMockTransactor(t)
//
//		// initialize random.StringGenerator
//		randomStringGenerator := random.StringGenerator{}
//		s.CreateMovieService = service.NewCreateMovieService(randomStringGenerator, movieTransactor)
//
//		// setup request body using anonymous struct
//		requestBody := struct {
//			Title    string `json:"title"`
//			Rated    string `json:"rated"`
//			Released string `json:"release_date"`
//			RunTime  int    `json:"run_time"`
//			Director string `json:"director"`
//			Writer   string `json:"writer"`
//		}{
//			Title:    "Repo Man",
//			Rated:    "R",
//			Released: "1984-03-02T00:00:00Z",
//			RunTime:  92,
//			Director: "Alex Cox",
//			Writer:   "Alex Cox",
//		}
//
//		// encode request body into buffer variable
//		var buf bytes.Buffer
//		err = json.NewEncoder(&buf).Encode(requestBody)
//		if err != nil {
//			t.Fatalf("Encode() error = %v", err)
//		}
//
//		// setup path
//		path := pathPrefix + moviesV1PathRoot
//
//		// form request using httptest
//		req := httptest.NewRequest(http.MethodPost, path, &buf)
//
//		// add test access token
//		req.Header.Add("Authorization", auth.BearerTokenType+" abc123def1")
//
//		// add application/JSON header to request
//		req.Header.Add(contentTypeHeaderKey, appJSONContentTypeHeaderVal)
//
//		// initialize ResponseRecorder to use with ServeHTTP as it
//		// satisfies ResponseWriter interface and records the response
//		// for testing
//		rr := httptest.NewRecorder()
//
//		// call the router ServeHTTP method to execute the request
//		// and record the response
//		s.router.ServeHTTP(rr, req)
//
//		// Assert that Response Status Code equals 200 (StatusOK)
//		c.Assert(rr.Code, qt.Equals, http.StatusOK)
//
//		// createMovieResponse is the response struct for a Movie
//		// the response struct is tucked inside the handler, so we
//		// have to recreate it here
//		type createMovieResponse struct {
//			ExternalID      string `json:"external_id"`
//			Title           string `json:"title"`
//			Rated           string `json:"rated"`
//			Released        string `json:"release_date"`
//			RunTime         int    `json:"run_time"`
//			Director        string `json:"director"`
//			Writer          string `json:"writer"`
//			CreateUsername  string `json:"create_username"`
//			CreateTimestamp string `json:"create_timestamp"`
//			UpdateUsername  string `json:"update_username"`
//			UpdateTimestamp string `json:"update_timestamp"`
//		}
//
//		// retrieve the mock User that is used for testing
//		u, _ := s.AccessTokenConverter.Convert(req.Context(), authtest.NewAccessToken(t))
//
//		// setup the expected response data
//		wantBody := createMovieResponse{
//			ExternalID:      "superRandomString",
//			Title:           "Repo Man",
//			Rated:           "R",
//			Released:        "1984-03-02T00:00:00Z",
//			RunTime:         92,
//			Director:        "Alex Cox",
//			Writer:          "Alex Cox",
//			CreateUsername:  u.Email,
//			CreateTimestamp: "",
//			UpdateUsername:  u.Email,
//			UpdateTimestamp: "",
//		}
//
//		// initialize createMovieResponse
//		gotBody := createMovieResponse{}
//
//		// decode the response body into gotBody
//		err = decoderErr(json.NewDecoder(rr.Result().Body).Decode(&gotBody))
//		defer rr.Result().Body.Close()
//
//		// Assert that there is no error after decoding the response body
//		c.Assert(err, qt.IsNil)
//
//		// quicktest uses Google's cmp library for DeepEqual comparisons. It
//		// has some great options included with it. Below is an example of
//		// ignoring certain fields...
//		ignoreFields := cmpopts.IgnoreFields(createMovieResponse{},
//			"ExternalID", "CreateTimestamp", "UpdateTimestamp")
//
//		// Assert that the response body (gotBody) is as expected (wantBody).
//		// The External ID needs to be unique as the database unique index
//		// requires it. As a result, the ExternalID field is ignored as part
//		// of the comparison. The Create/Update timestamps are ignored as
//		// well, as they are always unique.
//		// I could put another interface into the domain logic to solve
//		// for the timestamps and may do so later, but it's probably not
//		// necessary
//		c.Assert(gotBody, qt.CmpEquals(ignoreFields), wantBody)
//	})
//}
//
//func TestHandleMovieUpdate(t *testing.T) {
//	t.Run("typical", func(t *testing.T) {
//		// set environment variable NO_DB to skip database
//		// dependent tests
//		if os.Getenv("NO_DB") == "true" {
//			t.Skip("skipping db dependent test")
//		}
//
//		// initialize quickest checker
//		c := qt.New(t)
//
//		// setup Server
//		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)
//		rtr := NewMuxRouter()
//		driver := NewDriver()
//		params := NewServerParams(lgr, driver)
//		s, err := NewServer(rtr, params)
//		c.Assert(err, qt.IsNil)
//		s.AccessTokenConverter = authtest.NewMockAccessTokenConverter(t)
//		s.Authorizer = auth.CasbinAuthorizer{Enforcer: casbin.NewEnforcer("../config/rbac_model.conf", "../config/rbac_policy.csv")}
//
//		// initialize Datastore
//		ds, cleanup := datastoretest.NewDatastore(t)
//
//		// defer cleanup of the database until after the test is completed
//		t.Cleanup(cleanup)
//
//		// create a test movie in the database
//		m, movieCleanup := moviestore.NewMovieDBHelper(context.Background(), t, ds)
//
//		// defer cleanup of movie record until after the test is completed
//		t.Cleanup(movieCleanup)
//
//		// initialize the Transactor for the moviestore
//		transactor := moviestore.NewTransactor(ds)
//
//		s.UpdateMovieService = service.NewUpdateMovieService(transactor)
//
//		// setup request body using anonymous struct
//		requestBody := struct {
//			Title    string `json:"title"`
//			Rated    string `json:"rated"`
//			Released string `json:"release_date"`
//			RunTime  int    `json:"run_time"`
//			Director string `json:"director"`
//			Writer   string `json:"writer"`
//		}{
//			Title:    "Repo Man",
//			Rated:    "R",
//			Released: "1984-03-02T00:00:00Z",
//			RunTime:  92,
//			Director: "Alex Cox",
//			Writer:   "Alex Cox",
//		}
//
//		// encode request body into buffer variable
//		var buf bytes.Buffer
//		err = json.NewEncoder(&buf).Encode(requestBody)
//		if err != nil {
//			t.Fatalf("Encode() error = %v", err)
//		}
//
//		// setup path
//		path := pathPrefix + moviesV1PathRoot + "/" + m.ExternalID
//
//		// form request using httptest
//		req := httptest.NewRequest(http.MethodPut, path, &buf)
//
//		// add test access token
//		req.Header.Add("Authorization", auth.BearerTokenType+" abc123def1")
//
//		// add application/JSON header to request
//		req.Header.Add(contentTypeHeaderKey, appJSONContentTypeHeaderVal)
//
//		// initialize ResponseRecorder to use with ServeHTTP as it
//		// satisfies ResponseWriter interface and records the response
//		// for testing
//		rr := httptest.NewRecorder()
//
//		// call the router ServeHTTP method to execute the request
//		// and record the response
//		s.router.ServeHTTP(rr, req)
//
//		// Assert that Response Status Code equals 200 (StatusOK)
//		c.Assert(rr.Code, qt.Equals, http.StatusOK)
//
//		// retrieve the mock User that is used for testing
//		u, _ := s.AccessTokenConverter.Convert(req.Context(), authtest.NewAccessToken(t))
//
//		// setup the expected response data
//		wantBody := service.MovieResponse{
//			//ExternalID:      "superRandomString",
//			Title:          "Repo Man",
//			Rated:          "R",
//			Released:       "1984-03-02T00:00:00Z",
//			RunTime:        92,
//			Director:       "Alex Cox",
//			Writer:         "Alex Cox",
//			CreateUsername: u.Email,
//			//CreateTimestamp: "",
//			UpdateUsername: u.Email,
//			//UpdateTimestamp: "",
//		}
//
//		// initialize updateMovieResponse
//		gotBody := service.MovieResponse{}
//
//		// decode the response body into gotBody
//		err = decoderErr(json.NewDecoder(rr.Result().Body).Decode(&gotBody))
//		defer rr.Result().Body.Close()
//
//		// Assert that there is no error after decoding the response body
//		c.Assert(err, qt.IsNil)
//
//		// quicktest uses Google's cmp library for DeepEqual comparisons. It
//		// has some great options included with it. Below is an example of
//		// ignoring certain fields...
//		ignoreFields := cmpopts.IgnoreFields(service.MovieResponse{},
//			"ExternalID", "CreateTimestamp", "UpdateTimestamp")
//
//		// Assert that the response body (gotBody) is as expected (wantBody).
//		// The External ID needs to be unique as the database unique index
//		// requires it. As a result, the ExternalID field is ignored as part
//		// of the comparison. The Create/Update timestamps are ignored as
//		// well, as they are always unique.
//		// I could put another interface into the domain logic to solve
//		// for the timestamps and may do so later, but it's probably not
//		// necessary
//		c.Assert(gotBody, qt.CmpEquals(ignoreFields), wantBody)
//	})
//}
//
//func TestHandleMovieDelete(t *testing.T) {
//	t.Run("typical", func(t *testing.T) {
//		// set environment variable NO_DB to skip database
//		// dependent tests
//		if os.Getenv("NO_DB") == "true" {
//			t.Skip("skipping db dependent test")
//		}
//
//		// initialize quickest checker
//		c := qt.New(t)
//
//		// setup Server
//		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)
//		rtr := NewMuxRouter()
//		driver := NewDriver()
//		params := NewServerParams(lgr, driver)
//		s, err := NewServer(rtr, params)
//		c.Assert(err, qt.IsNil)
//		s.AccessTokenConverter = authtest.NewMockAccessTokenConverter(t)
//		s.Authorizer = auth.CasbinAuthorizer{Enforcer: casbin.NewEnforcer("../config/rbac_model.conf", "../config/rbac_policy.csv")}
//
//		// initialize Datastore
//		ds, cleanup := datastoretest.NewDatastore(t)
//
//		// defer cleanup of the database until after the test is completed
//		t.Cleanup(cleanup)
//
//		// create a test movie in the database, do not use cleanup
//		// function as this test should delete the movie
//		m, _ := moviestore.NewMovieDBHelper(context.Background(), t, ds)
//
//		// initialize the Transactor for the moviestore
//		transactor := moviestore.NewTransactor(ds)
//
//		// initialize the Selector for the moviestore
//		selector := moviestore.NewSelector(ds)
//
//		s.DeleteMovieService = service.NewDeleteMovieService(selector, transactor)
//
//		// setup path
//		path := pathPrefix + moviesV1PathRoot + "/" + m.ExternalID
//
//		// form request using httptest
//		req := httptest.NewRequest(http.MethodDelete, path, nil)
//
//		// add test access token
//		req.Header.Add("Authorization", auth.BearerTokenType+" abc123def1")
//
//		// initialize ResponseRecorder to use with ServeHTTP as it
//		// satisfies ResponseWriter interface and records the response
//		// for testing
//		rr := httptest.NewRecorder()
//
//		// call the router ServeHTTP method to execute the request
//		// and record the response
//		s.router.ServeHTTP(rr, req)
//
//		// Assert that Response Status Code equals 200 (StatusOK)
//		c.Assert(rr.Code, qt.Equals, http.StatusOK)
//
//		// setup the expected response data
//		wantBody := service.DeleteMovieResponse{
//			ExternalID: m.ExternalID,
//			Deleted:    true,
//		}
//
//		// initialize deleteMovieResponse
//		gotBody := service.DeleteMovieResponse{}
//
//		// decode the response body into gotBody
//		err = decoderErr(json.NewDecoder(rr.Result().Body).Decode(&gotBody))
//		defer rr.Result().Body.Close()
//
//		// Assert that there is no error after decoding the response body
//		c.Assert(err, qt.IsNil)
//
//		// Assert that the response body (gotBody) is as expected (wantBody).
//		// The External ID needs to be unique as the database unique index
//		// requires it. As a result, the ExternalID field is ignored as part
//		// of the comparison. The Create/Update timestamps are ignored as
//		// well, as they are always unique.
//		// I could put another interface into the domain logic to solve
//		// for the timestamps and may do so later, but it's probably not
//		// necessary
//		c.Assert(gotBody, qt.Equals, wantBody)
//	})
//}
//
//func TestHandleFindMovieByID(t *testing.T) {
//	t.Run("typical", func(t *testing.T) {
//		// set environment variable NO_DB to skip database
//		// dependent tests
//		if os.Getenv("NO_DB") == "true" {
//			t.Skip("skipping db dependent test")
//		}
//
//		// initialize quickest checker
//		c := qt.New(t)
//
//		// setup Server
//		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)
//		rtr := NewMuxRouter()
//		driver := NewDriver()
//		params := NewServerParams(lgr, driver)
//		s, err := NewServer(rtr, params)
//		c.Assert(err, qt.IsNil)
//		s.AccessTokenConverter = authtest.NewMockAccessTokenConverter(t)
//		s.Authorizer = auth.CasbinAuthorizer{Enforcer: casbin.NewEnforcer("../config/rbac_model.conf", "../config/rbac_policy.csv")}
//
//		// initialize Datastore
//		ds, cleanup := datastoretest.NewDatastore(t)
//
//		// defer cleanup of the database until after the test is completed
//		t.Cleanup(cleanup)
//
//		// create a test movie in the database
//		m, movieCleanup := moviestore.NewMovieDBHelper(context.Background(), t, ds)
//
//		// defer cleanup of movie record until after the test is completed
//		t.Cleanup(movieCleanup)
//
//		// initialize the Selector for the moviestore
//		selector := moviestore.NewSelector(ds)
//
//		s.FindMovieService = service.NewFindMovieService(selector)
//
//		// setup request body using anonymous struct
//		requestBody := struct {
//			Title    string `json:"title"`
//			Rated    string `json:"rated"`
//			Released string `json:"release_date"`
//			RunTime  int    `json:"run_time"`
//			Director string `json:"director"`
//			Writer   string `json:"writer"`
//		}{
//			Title:    "Repo Man",
//			Rated:    "R",
//			Released: "1984-03-02T00:00:00Z",
//			RunTime:  92,
//			Director: "Alex Cox",
//			Writer:   "Alex Cox",
//		}
//
//		// encode request body into buffer variable
//		var buf bytes.Buffer
//		err = json.NewEncoder(&buf).Encode(requestBody)
//		if err != nil {
//			t.Fatalf("Encode() error = %v", err)
//		}
//
//		// setup path
//		path := pathPrefix + moviesV1PathRoot + "/" + m.ExternalID
//
//		// form request using httptest
//		req := httptest.NewRequest(http.MethodGet, path, &buf)
//
//		// add test access token
//		req.Header.Add("Authorization", auth.BearerTokenType+" abc123def1")
//
//		// initialize ResponseRecorder to use with ServeHTTP as it
//		// satisfies ResponseWriter interface and records the response
//		// for testing
//		rr := httptest.NewRecorder()
//
//		// call the router ServeHTTP method to execute the request
//		// and record the response
//		s.router.ServeHTTP(rr, req)
//
//		// Assert that Response Status Code equals 200 (StatusOK)
//		c.Assert(rr.Code, qt.Equals, http.StatusOK)
//
//		// retrieve the mock User that is used for testing
//		u, _ := s.AccessTokenConverter.Convert(req.Context(), authtest.NewAccessToken(t))
//
//		// setup the expected response data
//		wantBody := service.MovieResponse{
//			ExternalID:     m.ExternalID,
//			Title:          "Repo Man",
//			Rated:          "R",
//			Released:       "1984-03-02T00:00:00Z",
//			RunTime:        92,
//			Director:       "Alex Cox",
//			Writer:         "Alex Cox",
//			CreateUsername: u.Email,
//			//CreateTimestamp: "",
//			UpdateUsername: u.Email,
//			//UpdateTimestamp: "",
//		}
//
//		// initialize movieResponse
//		gotBody := service.MovieResponse{}
//
//		// decode the response body into gotBody
//		err = decoderErr(json.NewDecoder(rr.Result().Body).Decode(&gotBody))
//		defer rr.Result().Body.Close()
//
//		// Assert that there is no error after decoding the response body
//		c.Assert(err, qt.IsNil)
//
//		// quicktest uses Google's cmp library for DeepEqual comparisons. It
//		// has some great options included with it. Below is an example of
//		// ignoring certain fields...
//		ignoreFields := cmpopts.IgnoreFields(service.MovieResponse{},
//			"CreateTimestamp", "UpdateTimestamp")
//
//		// Assert that the response body (gotBody) is as expected (wantBody).
//		// The External ID needs to be unique as the database unique index
//		// requires it. As a result, the ExternalID field is ignored as part
//		// of the comparison. The Create/Update timestamps are ignored as
//		// well, as they are always unique.
//		// I could put another interface into the domain logic to solve
//		// for the timestamps and may do so later, but it's probably not
//		// necessary
//		c.Assert(gotBody, qt.CmpEquals(ignoreFields), wantBody)
//	})
//}
//
//func TestHandleFindAllMovies(t *testing.T) {
//	t.Run("typical", func(t *testing.T) {
//		// set environment variable NO_DB to skip database
//		// dependent tests
//		if os.Getenv("NO_DB") == "true" {
//			t.Skip("skipping db dependent test")
//		}
//
//		// initialize quickest checker
//		c := qt.New(t)
//
//		// setup Server
//		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)
//		rtr := NewMuxRouter()
//		driver := NewDriver()
//		params := NewServerParams(lgr, driver)
//		s, err := NewServer(rtr, params)
//		c.Assert(err, qt.IsNil)
//		s.AccessTokenConverter = authtest.NewMockAccessTokenConverter(t)
//		s.Authorizer = auth.CasbinAuthorizer{Enforcer: casbin.NewEnforcer("../config/rbac_model.conf", "../config/rbac_policy.csv")}
//
//		// initialize MockSelector for the moviestore
//		mockSelector := newMockSelector(t)
//
//		s.FindMovieService = service.NewFindMovieService(mockSelector)
//
//		// setup path
//		path := pathPrefix + moviesV1PathRoot
//
//		// form request using httptest
//		req := httptest.NewRequest(http.MethodGet, path, nil)
//
//		// add test access token
//		req.Header.Add("Authorization", auth.BearerTokenType+" abc123def1")
//
//		// initialize ResponseRecorder to use with ServeHTTP as it
//		// satisfies ResponseWriter interface and records the response
//		// for testing
//		rr := httptest.NewRecorder()
//
//		// call the router ServeHTTP method to execute the request
//		// and record the response
//		s.router.ServeHTTP(rr, req)
//
//		// Assert that Response Status Code equals 200 (StatusOK)
//		c.Assert(rr.Code, qt.Equals, http.StatusOK)
//
//		// get mocked slice of movies that should be returned
//		movies, err := mockSelector.FindAll(req.Context())
//		if err != nil {
//			t.Fatalf("mockSelector.FindAll error = %v", err)
//		}
//
//		var smr []service.MovieResponse
//		for _, m := range movies {
//			mr := service.MovieResponse{
//				ExternalID:      m.ExternalID,
//				Title:           m.Title,
//				Rated:           m.Rated,
//				Released:        m.Released.Format(time.RFC3339),
//				RunTime:         m.RunTime,
//				Director:        m.Director,
//				Writer:          m.Writer,
//				CreateUsername:  m.CreateUser.Email,
//				CreateTimestamp: m.CreateTime.Format(time.RFC3339),
//				UpdateUsername:  m.UpdateUser.Email,
//				UpdateTimestamp: m.UpdateTime.Format(time.RFC3339),
//			}
//			smr = append(smr, mr)
//		}
//
//		// setup the expected response data
//		wantBody := smr
//
//		// initialize a slice of movieResponse{}
//		gotBody := []service.MovieResponse{}
//
//		// decode the response body into gotBody
//		err = decoderErr(json.NewDecoder(rr.Result().Body).Decode(&gotBody))
//		defer rr.Result().Body.Close()
//
//		// Assert that there is no error after decoding the response body
//		c.Assert(err, qt.IsNil)
//
//		// Assert that the response body (gotBody) is as expected (wantBody).
//		c.Assert(gotBody, qt.DeepEquals, wantBody)
//	})
//}
//
//func TestHandleLoggerRead(t *testing.T) {
//	t.Run("typical", func(t *testing.T) {
//		// initialize quickest checker
//		c := qt.New(t)
//
//		// initialize a zerolog Logger
//		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)
//
//		// set global logging level to Info
//		zerolog.SetGlobalLevel(zerolog.InfoLevel)
//
//		// set Error stack trace to true
//		logger.WriteErrorStackGlobal(true)
//
//		rtr := NewMuxRouter()
//		driver := NewDriver()
//		params := NewServerParams(lgr, driver)
//
//		s, err := NewServer(rtr, params)
//		c.Assert(err, qt.IsNil)
//
//		s.AccessTokenConverter = authtest.NewMockAccessTokenConverter(t)
//		s.Authorizer = auth.CasbinAuthorizer{Enforcer: casbin.NewEnforcer("../config/rbac_model.conf", "../config/rbac_policy.csv")}
//
//		s.LoggerService = service.NewLoggerService(lgr)
//
//		// setup path
//		path := pathPrefix + loggerV1PathRoot
//
//		// form request using httptest
//		req := httptest.NewRequest(http.MethodGet, path, nil)
//
//		// add test access token
//		req.Header.Add("Authorization", auth.BearerTokenType+" abc123def1")
//
//		// initialize ResponseRecorder to use with ServeHTTP as it
//		// satisfies ResponseWriter interface and records the response
//		// for testing
//		rr := httptest.NewRecorder()
//
//		s.router.ServeHTTP(rr, req)
//
//		// Assert that Response Status Code equals 200 (StatusOK)
//		c.Assert(rr.Code, qt.Equals, http.StatusOK)
//
//		// setup the expected response data
//		wantBody := service.LoggerResponse{
//			LoggerMinimumLevel: zerolog.DebugLevel.String(),
//			GlobalLogLevel:     zerolog.InfoLevel.String(),
//			LogErrorStack:      true,
//		}
//
//		// initialize readLoggerResponse
//		gotBody := service.LoggerResponse{}
//
//		// decode the response body into gotBody
//		err = decoderErr(json.NewDecoder(rr.Result().Body).Decode(&gotBody))
//		defer rr.Result().Body.Close()
//
//		// Assert that there is no error after decoding the response body
//		c.Assert(err, qt.IsNil)
//
//		// Assert that the response body (gotBody) is as expected (wantBody).
//		c.Assert(gotBody, qt.DeepEquals, wantBody)
//	})
//}
//
//func TestHandleLoggerUpdate(t *testing.T) {
//	t.Run("typical", func(t *testing.T) {
//		// initialize quickest checker
//		c := qt.New(t)
//
//		// initialize a zerolog Logger
//		lgr := logger.NewLogger(os.Stdout, zerolog.TraceLevel, true)
//
//		// set global logging level to Info
//		zerolog.SetGlobalLevel(zerolog.InfoLevel)
//
//		// set Error stack to true
//		logger.WriteErrorStackGlobal(false)
//
//		t.Logf("Minimum accepted log level set to %s", lgr.GetLevel().String())
//		t.Logf("Initial global log level set to %s", zerolog.GlobalLevel())
//		var logErrorStack bool
//		if zerolog.ErrorStackMarshaler != nil {
//			logErrorStack = true
//		}
//		t.Logf("Initial Write Error Stack global set to %t", logErrorStack)
//
//		rtr := NewMuxRouter()
//		driver := NewDriver()
//		params := NewServerParams(lgr, driver)
//
//		s, err := NewServer(rtr, params)
//		c.Assert(err, qt.IsNil)
//
//		s.AccessTokenConverter = authtest.NewMockAccessTokenConverter(t)
//		s.Authorizer = auth.CasbinAuthorizer{Enforcer: casbin.NewEnforcer("../config/rbac_model.conf", "../config/rbac_policy.csv")}
//
//		s.LoggerService = service.NewLoggerService(lgr)
//
//		// setup request body using anonymous struct
//		requestBody := struct {
//			GlobalLogLevel string `json:"global_log_level,omitempty"`
//			LogErrorStack  string `json:"log_error_stack,omitempty"`
//		}{
//			GlobalLogLevel: "debug",
//			LogErrorStack:  "true",
//		}
//
//		// encode request body into buffer variable
//		var buf bytes.Buffer
//		err = json.NewEncoder(&buf).Encode(requestBody)
//		if err != nil {
//			t.Fatalf("Encode() error = %v", err)
//		}
//
//		// setup path
//		path := pathPrefix + loggerV1PathRoot
//
//		// form request using httptest
//		req := httptest.NewRequest(http.MethodPut, path, &buf)
//
//		// add application/JSON header to request
//		req.Header.Add(contentTypeHeaderKey, appJSONContentTypeHeaderVal)
//
//		// add test access token
//		req.Header.Add("Authorization", auth.BearerTokenType+" abc123def1")
//
//		// initialize ResponseRecorder to use with ServeHTTP as it
//		// satisfies ResponseWriter interface and records the response
//		// for testing
//		rr := httptest.NewRecorder()
//
//		// call the router ServeHTTP method to execute the request
//		// and record the response
//		s.router.ServeHTTP(rr, req)
//
//		// Assert that Response Status Code equals 200 (StatusOK)
//		c.Assert(rr.Code, qt.Equals, http.StatusOK)
//
//		// setup the expected response data
//		wantBody := service.LoggerResponse{
//			LoggerMinimumLevel: zerolog.TraceLevel.String(),
//			GlobalLogLevel:     zerolog.DebugLevel.String(),
//			LogErrorStack:      true,
//		}
//
//		// initialize readLoggerResponse
//		gotBody := service.LoggerResponse{}
//
//		// decode the response body into gotBody
//		err = decoderErr(json.NewDecoder(rr.Result().Body).Decode(&gotBody))
//		defer rr.Result().Body.Close()
//
//		// Assert that there is no error after decoding the response body
//		c.Assert(err, qt.IsNil)
//
//		// Assert that the response body (gotBody) is as expected (wantBody).
//		c.Assert(gotBody, qt.DeepEquals, wantBody)
//	})
//}
//
//func TestHandlePing(t *testing.T) {
//	t.Run("typical", func(t *testing.T) {
//
//		c := qt.New(t)
//		var emptyBody []byte
//
//		// initialize a zerolog Logger
//		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)
//
//		// set global logging level to Info
//		zerolog.SetGlobalLevel(zerolog.InfoLevel)
//
//		// set Error stack trace to true
//		logger.WriteErrorStackGlobal(true)
//
//		rtr := NewMuxRouter()
//		driver := NewDriver()
//		params := NewServerParams(lgr, driver)
//
//		s, err := NewServer(rtr, params)
//		c.Assert(err, qt.IsNil)
//
//		s.AccessTokenConverter = authtest.NewMockAccessTokenConverter(t)
//		s.Authorizer = auth.CasbinAuthorizer{Enforcer: casbin.NewEnforcer("../config/rbac_model.conf", "../config/rbac_policy.csv")}
//
//		// initialize Datastore
//		ds, cleanup := datastoretest.NewDatastore(t)
//
//		// defer cleanup of the database until after the test is completed
//		t.Cleanup(cleanup)
//
//		pinger := pingstore.NewPinger(ds)
//		s.PingService = service.NewPingService(pinger)
//
//		path := "/api/v1/ping"
//		req := httptest.NewRequest(http.MethodGet, path, bytes.NewBuffer(emptyBody))
//		rr := httptest.NewRecorder()
//
//		// call the router ServeHTTP method to execute the request
//		// and record the response
//		s.router.ServeHTTP(rr, req)
//
//		wantBody := service.PingResponse{DBUp: true}
//
//		gotBody := service.PingResponse{}
//		err = decoderErr(json.NewDecoder(rr.Result().Body).Decode(&gotBody))
//		defer rr.Result().Body.Close()
//		c.Assert(err, qt.IsNil)
//
//		// Response Status Code should be 200
//		c.Assert(rr.Code, qt.Equals, http.StatusOK)
//
//		// Assert that the response body equals the body we want
//		c.Assert(gotBody, qt.DeepEquals, wantBody)
//	})
//
//	t.Run("mock", func(t *testing.T) {
//		c := qt.New(t)
//		var emptyBody []byte
//
//		// initialize a zerolog Logger
//		lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)
//
//		// set global logging level to Info
//		zerolog.SetGlobalLevel(zerolog.InfoLevel)
//
//		// set Error stack trace to true
//		logger.WriteErrorStackGlobal(true)
//
//		rtr := NewMuxRouter()
//		driver := NewDriver()
//		params := NewServerParams(lgr, driver)
//
//		s, err := NewServer(rtr, params)
//		c.Assert(err, qt.IsNil)
//
//		s.AccessTokenConverter = authtest.NewMockAccessTokenConverter(t)
//		s.Authorizer = auth.CasbinAuthorizer{Enforcer: casbin.NewEnforcer("../config/rbac_model.conf", "../config/rbac_policy.csv")}
//
//		// use mockPinger instead of a real db
//		pinger := mockPinger{}
//		s.PingService = service.NewPingService(pinger)
//
//		path := "/api/v1/ping"
//		req := httptest.NewRequest(http.MethodGet, path, bytes.NewBuffer(emptyBody))
//		rr := httptest.NewRecorder()
//
//		// call the router ServeHTTP method to execute the request
//		// and record the response
//		s.router.ServeHTTP(rr, req)
//
//		wantBody := service.PingResponse{DBUp: true}
//
//		gotBody := service.PingResponse{}
//		err = decoderErr(json.NewDecoder(rr.Result().Body).Decode(&gotBody))
//		defer rr.Result().Body.Close()
//		c.Assert(err, qt.IsNil)
//
//		// Response Status Code should be 200
//		c.Assert(rr.Code, qt.Equals, http.StatusOK)
//
//		// Assert that the response body equals the body we want
//		c.Assert(gotBody, qt.DeepEquals, wantBody)
//	})
//
//}
//
//type mockPinger struct{}
//
//func (m mockPinger) PingDB(ctx context.Context) error {
//	return nil
//}
