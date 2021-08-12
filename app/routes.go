package app

import (
	"net/http"
)

const (
	contentTypeHeaderKey        string = "Content-Type"
	appJSONContentTypeHeaderVal string = "application/json"
	moviesV1PathRoot            string = "/v1/movies"
	extlIDPathDir               string = "/{extlID}"
	loggerV1PathRoot            string = "/v1/logger"
	pingV1PathRoot              string = "/v1/ping"
)

// register routes/middleware/handlers to the Server router
func (s *Server) routes() {

	// Match only POST requests at /api/v1/movies
	// with Content-Type header = application/json
	s.router.Handle(moviesV1PathRoot,
		s.loggerChain().Extend(s.ctxWithUserChain()).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleMovieCreate)).
		Methods(http.MethodPost).
		Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

	// Match only PUT requests having an ID at /api/v1/movies/{extlID}
	// with the Content-Type header = application/json
	s.router.Handle(moviesV1PathRoot+extlIDPathDir,
		s.loggerChain().Extend(s.ctxWithUserChain()).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleMovieUpdate)).
		Methods(http.MethodPut).
		Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

	// Match only DELETE requests having an ID at /api/v1/movies/{extlID}
	s.router.Handle(moviesV1PathRoot+extlIDPathDir,
		s.loggerChain().Extend(s.ctxWithUserChain()).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleMovieDelete)).
		Methods(http.MethodDelete)

	// Match only GET requests having an ID at /api/v1/movies/{extlID}
	s.router.Handle(moviesV1PathRoot+extlIDPathDir,
		s.loggerChain().Extend(s.ctxWithUserChain()).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleFindMovieByID)).
		Methods(http.MethodGet)

	// Match only GET requests /api/v1/movies
	s.router.Handle(moviesV1PathRoot,
		s.loggerChain().Extend(s.ctxWithUserChain()).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleFindAllMovies)).
		Methods(http.MethodGet)

	// Match only GET requests /api/v1/logger
	s.router.Handle(loggerV1PathRoot,
		s.loggerChain().Extend(s.ctxWithUserChain()).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleLoggerRead)).
		Methods(http.MethodGet)

	// Match only PUT requests /api/v1/logger
	s.router.Handle(loggerV1PathRoot,
		s.loggerChain().Extend(s.ctxWithUserChain()).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleLoggerUpdate)).
		Methods(http.MethodPut).
		Headers(contentTypeHeaderKey, appJSONContentTypeHeaderVal)

	// Match only GET requests at /api/v1/ping
	s.router.Handle(pingV1PathRoot,
		s.loggerChain().
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handlePing)).
		Methods(http.MethodGet)

}
