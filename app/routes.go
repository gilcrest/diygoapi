package app

import (
	"net/http"
)

const (
	contentTypeHeaderKey        string = "Content-Type"
	appJSONContentTypeHeaderVal string = "application/json"
	defaultRealm                string = "go-api-basic"
	moviesV1PathRoot            string = "/v1/movies"
	orgsV1PathRoot              string = "/v1/orgs"
	appsV1PathRoot              string = "/v1/apps"
	extlIDPathDir               string = "/{extlID}"
	loggerV1PathRoot            string = "/v1/logger"
	pingV1PathRoot              string = "/v1/ping"
)

// register routes/middleware/handlers to the Server router
func (s *Server) registerRoutes() {

	//// Match only POST requests at /api/v1/movies
	//// with Content-Type header = application/json
	//s.router.Handle(moviesV1PathRoot,
	//	s.loggerChain().
	//		Append(s.appHandler).
	//		Append(s.userHandler).
	//		Append(s.authorizeUserHandler).
	//		Append(s.jsonContentTypeResponseHandler).
	//		ThenFunc(s.handleMovieCreate)).
	//	Methods(http.MethodPost).
	//	Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)
	//
	//// Match only PUT requests having an ID at /api/v1/movies/{extlID}
	//// with the Content-Type header = application/json
	//s.router.Handle(moviesV1PathRoot+extlIDPathDir,
	//	s.loggerChain().
	//		Append(s.appHandler).
	//		Append(s.userHandler).
	//		Append(s.authorizeUserHandler).
	//		Append(s.jsonContentTypeResponseHandler).
	//		ThenFunc(s.handleMovieUpdate)).
	//	Methods(http.MethodPut).
	//	Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)
	//
	//// Match only DELETE requests having an ID at /api/v1/movies/{extlID}
	//s.router.Handle(moviesV1PathRoot+extlIDPathDir,
	//	s.loggerChain().
	//		Append(s.appHandler).
	//		Append(s.userHandler).
	//		Append(s.authorizeUserHandler).
	//		Append(s.jsonContentTypeResponseHandler).
	//		ThenFunc(s.handleMovieDelete)).
	//	Methods(http.MethodDelete)
	//
	//// Match only GET requests having an ID at /api/v1/movies/{extlID}
	//s.router.Handle(moviesV1PathRoot+extlIDPathDir,
	//	s.loggerChain().
	//		Append(s.appHandler).
	//		Append(s.userHandler).
	//		Append(s.authorizeUserHandler).
	//		Append(s.jsonContentTypeResponseHandler).
	//		ThenFunc(s.handleFindMovieByID)).
	//	Methods(http.MethodGet)
	//
	//// Match only GET requests /api/v1/movies
	//s.router.Handle(moviesV1PathRoot,
	//	s.loggerChain().
	//		Append(s.appHandler).
	//		Append(s.userHandler).
	//		Append(s.authorizeUserHandler).
	//		Append(s.jsonContentTypeResponseHandler).
	//		ThenFunc(s.handleFindAllMovies)).
	//	Methods(http.MethodGet)

	// Match only GET requests at /api/v1/orgs
	s.router.Handle(orgsV1PathRoot,
		s.loggerChain().
			Append(s.appHandler).
			Append(s.userHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleOrgFindAll)).
		Methods(http.MethodGet)

	// Match only GET requests at /api/v1/orgs/{extlID}
	s.router.Handle(orgsV1PathRoot+extlIDPathDir,
		s.loggerChain().
			Append(s.appHandler).
			Append(s.userHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleOrgFindByExtlID)).
		Methods(http.MethodGet)

	// Match only POST requests at /api/v1/orgs
	// with Content-Type header = application/json
	s.router.Handle(orgsV1PathRoot,
		s.loggerChain().
			Append(s.appHandler).
			Append(s.userHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleOrgCreate)).
		Methods(http.MethodPost).
		Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

	// Match only PUT requests at /api/v1/orgs/{extlID}
	// with Content-Type header = application/json
	s.router.Handle(orgsV1PathRoot+extlIDPathDir,
		s.loggerChain().
			Append(s.appHandler).
			Append(s.userHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleOrgUpdate)).
		Methods(http.MethodPut).
		Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

	// Match only POST requests at /api/v1/apps
	// with Content-Type header = application/json
	s.router.Handle(appsV1PathRoot,
		s.loggerChain().
			Append(s.appHandler).
			Append(s.userHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleAppCreate)).
		Methods(http.MethodPost).
		Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

	// Match only GET requests /api/v1/logger
	s.router.Handle(loggerV1PathRoot,
		s.loggerChain().
			Append(s.appHandler).
			Append(s.userHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleLoggerRead)).
		Methods(http.MethodGet)

	// Match only PUT requests /api/v1/logger
	s.router.Handle(loggerV1PathRoot,
		s.loggerChain().
			Append(s.appHandler).
			Append(s.userHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleLoggerUpdate)).
		Methods(http.MethodPut).
		Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

	// Match only GET requests at /api/v1/ping
	s.router.Handle(pingV1PathRoot,
		s.loggerChain().
			Append(s.appHandler).
			Append(s.userHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handlePing)).
		Methods(http.MethodGet)

	// Match only POST requests at /api/v1/seed
	s.router.Handle("/v1/seed",
		s.loggerChain().
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleSeed)).
		Methods(http.MethodPost).
		Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)
}
