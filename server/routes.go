package server

// register routes/middleware/handlers to the Server ServeMux
func (s *Server) registerRoutes() {

	// Match only POST requests at /api/v1/movies
	// with Content-Type header = application/json
	s.mux.Handle("POST /api/v1/movies",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.enforceJSONContentTypeHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleMovieCreate))

	// Match only PUT requests having an ID at /api/v1/movies/{extlID}
	// with the Content-Type header = application/json
	s.mux.Handle("PUT /api/v1/movies/{extlID}",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.enforceJSONContentTypeHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleMovieUpdate))

	// Match only DELETE requests having an ID at /api/v1/movies/{extlID}
	s.mux.Handle("DELETE /api/v1/movies/{extlID}",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleMovieDelete))

	// Match only GET requests having an ID at /api/v1/movies/{extlID}
	s.mux.Handle("GET /api/v1/movies/{extlID}",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleFindMovieByID))

	// Match only GET requests /api/v1/movies
	s.mux.Handle("GET /api/v1/movies",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleFindAllMovies))

	// Match only POST requests at /api/v1/orgs
	// with Content-Type header = application/json
	s.mux.Handle("POST /api/v1/orgs",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.enforceJSONContentTypeHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleOrgCreate))

	// Match only PUT requests at /api/v1/orgs/{extlID}
	// with Content-Type header = application/json
	s.mux.Handle("PUT /api/v1/orgs/{extlID}",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.enforceJSONContentTypeHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleOrgUpdate))

	// Match only DELETE requests at /api/v1/orgs/{extlID}
	s.mux.Handle("DELETE /api/v1/orgs/{extlID}",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleOrgDelete))

	// Match only GET requests at /api/v1/orgs
	s.mux.Handle("GET /api/v1/orgs",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleOrgFindAll))

	// Match only GET requests at /api/v1/orgs/{extlID}
	s.mux.Handle("GET /api/v1/orgs/{extlID}",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleOrgFindByExtlID))

	// Match only POST requests at /api/v1/apps
	// with Content-Type header = application/json
	s.mux.Handle("POST /api/v1/apps",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.enforceJSONContentTypeHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleAppCreate))

	// Match only POST requests at /api/v1/users
	s.mux.Handle("POST /api/v1/users",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.enforceJSONContentTypeHandler).
			Append(s.appHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleNewUser))

	// Match only GET requests /api/v1/logger
	s.mux.Handle("GET /api/v1/logger",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleLoggerRead))

	// Match only PUT requests /api/v1/logger
	s.mux.Handle("PUT /api/v1/logger",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.enforceJSONContentTypeHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleLoggerUpdate))

	// Match only GET requests at /api/v1/ping
	s.mux.Handle("GET /api/v1/ping",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.authorizeUserHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handlePing))

	// Match only POST requests at /api/v1/permissions
	s.mux.Handle("POST /api/v1/permissions",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.enforceJSONContentTypeHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handlePermissionCreate))

	// Match only GET requests at /api/v1/permissions
	s.mux.Handle("GET /api/v1/permissions",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handlePermissionFindAll))

	// Match only DELETE requests at /api/v1/permissions/{extlID}
	s.mux.Handle("DELETE /api/v1/permissions/{extlID}",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.appHandler).
			Append(s.authHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handlePermissionDelete))

	// Match only POST requests at /api/v1/genesis
	s.mux.Handle("POST /api/v1/genesis",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.genesisAuthHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleGenesis))

	// Match only GET requests at /api/v1/genesis
	s.mux.Handle("GET /api/v1/genesis",
		s.loggerChain().
			Append(s.addRequestHandlerPatternContextHandler).
			Append(s.jsonContentTypeResponseHandler).
			ThenFunc(s.handleGenesisRead))
}
