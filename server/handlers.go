package server

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rs/zerolog/hlog"

	"github.com/gilcrest/diy-go-api/domain/audit"
	"github.com/gilcrest/diy-go-api/domain/auth"
	"github.com/gilcrest/diy-go-api/domain/errs"
	"github.com/gilcrest/diy-go-api/service"
)

// CreateMovie is a HandlerFunc used to create a Movie
func (s *Server) handleMovieCreate(w http.ResponseWriter, r *http.Request) {
	logger := *hlog.FromRequest(r)

	adt, err := audit.FromRequest(r)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// Declare request body (rb) as an instance of service.MovieRequest
	rb := new(service.CreateMovieRequest)

	// Decode JSON HTTP request body into a Decoder type
	// and unmarshal that into the CreateMovieRequest struct (rb)
	err = json.NewDecoder(r.Body).Decode(&rb)
	defer r.Body.Close()
	// Call decoderErr to determine if body is nil, json is malformed
	// or any other error
	err = decoderErr(err)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	response, err := s.CreateMovieService.Create(r.Context(), rb, adt)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
		return
	}
}

// handleMovieUpdate handles PUT requests for the /movies/{id} endpoint
// and updates the given movie
func (s *Server) handleMovieUpdate(w http.ResponseWriter, r *http.Request) {

	logger := *hlog.FromRequest(r)

	adt, err := audit.FromRequest(r)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// gorilla mux Vars function returns the route variables for the
	// current request, if any. id is the external id given for the
	// movie
	vars := mux.Vars(r)
	extlid := vars["extlID"]

	// Declare request body (rb) as an instance of service.MovieRequest
	rb := new(service.UpdateMovieRequest)

	// Decode JSON HTTP request body into a Decoder type
	// and unmarshal that into requestData
	err = json.NewDecoder(r.Body).Decode(&rb)
	defer r.Body.Close()
	// Call DecoderErr to determine if body is nil, json is malformed
	// or any other error
	err = decoderErr(err)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// External ID is from path variable, need to set separate
	// from decoding response body
	rb.ExternalID = extlid

	response, err := s.UpdateMovieService.Update(r.Context(), rb, adt)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
		return
	}
}

// handleMovieDelete handles DELETE requests for the /movies/{id} endpoint
// and updates the given movie
func (s *Server) handleMovieDelete(w http.ResponseWriter, r *http.Request) {

	logger := *hlog.FromRequest(r)

	// gorilla mux Vars function returns the route variables for the
	// current request, if any. id is the external id given for the
	// movie
	vars := mux.Vars(r)
	extlID := vars["extlID"]

	response, err := s.DeleteMovieService.Delete(r.Context(), extlID)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
		return
	}
}

// handleFindMovieByID handles GET requests for the /movies/{id} endpoint
// and finds a movie by its ID
func (s *Server) handleFindMovieByID(w http.ResponseWriter, r *http.Request) {

	logger := *hlog.FromRequest(r)

	// gorilla mux Vars function returns the route variables for the
	// current request, if any. id is the external id given for the
	// movie
	vars := mux.Vars(r)
	extlID := vars["extlID"]

	response, err := s.FindMovieService.FindMovieByID(r.Context(), extlID)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
		return
	}
}

// handleFindAllMovies handles GET requests for the /movies endpoint and finds
// all movies
func (s *Server) handleFindAllMovies(w http.ResponseWriter, r *http.Request) {

	logger := *hlog.FromRequest(r)

	response, err := s.FindMovieService.FindAllMovies(r.Context())
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
		return
	}
}

// handleOrgCreate is a HandlerFunc used to create an Org
func (s *Server) handleOrgCreate(w http.ResponseWriter, r *http.Request) {
	lgr := *hlog.FromRequest(r)

	adt, err := audit.FromRequest(r)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	// Declare request body (rb) as an instance of service.MovieRequest
	rb := new(service.CreateOrgRequest)

	// Decode JSON HTTP request body into a Decoder type
	// and unmarshal that into the MovieRequest struct in the
	// AddMovieHandler
	err = json.NewDecoder(r.Body).Decode(&rb)
	defer r.Body.Close()
	// Call decoderErr to determine if body is nil, json is malformed
	// or any other error
	err = decoderErr(err)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	var response service.OrgResponse
	response, err = s.CreateOrgService.Create(r.Context(), rb, adt)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, errs.E(errs.Internal, err))
		return
	}
}

// handleOrgUpdate is a HandlerFunc used to update an Org
func (s *Server) handleOrgUpdate(w http.ResponseWriter, r *http.Request) {
	lgr := *hlog.FromRequest(r)

	adt, err := audit.FromRequest(r)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	// Declare request body (rb) as an instance of service.MovieRequest
	rb := new(service.UpdateOrgRequest)

	// Decode JSON HTTP request body into a Decoder type
	// and unmarshal that into the MovieRequest struct in the
	// AddMovieHandler
	err = json.NewDecoder(r.Body).Decode(&rb)
	defer r.Body.Close()
	// Call decoderErr to determine if body is nil, json is malformed
	// or any other error
	err = decoderErr(err)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	// gorilla mux Vars function returns the route variables for the
	// current request, if any. ID is the external id given for the resource
	vars := mux.Vars(r)
	rb.ExternalID = vars["extlID"]

	var response service.OrgResponse
	response, err = s.OrgService.Update(r.Context(), rb, adt)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, errs.E(errs.Internal, err))
		return
	}
}

// handleOrgDelete is a HandlerFunc used to delete an Org
func (s *Server) handleOrgDelete(w http.ResponseWriter, r *http.Request) {
	logger := *hlog.FromRequest(r)

	// gorilla mux Vars function returns the route variables for the
	// current request, if any.
	vars := mux.Vars(r)
	// extlID is the external id given for the resource
	extlID := vars["extlID"]

	response, err := s.OrgService.Delete(r.Context(), extlID)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
		return
	}
}

// handleOrgFindAll is a HandlerFunc used to find a list of Orgs
func (s *Server) handleOrgFindAll(w http.ResponseWriter, r *http.Request) {
	logger := *hlog.FromRequest(r)

	response, err := s.OrgService.FindAll(r.Context())
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
		return
	}
}

// handleOrgFindByExtlID is a HandlerFunc used to find a specific Org by External ID
func (s *Server) handleOrgFindByExtlID(w http.ResponseWriter, r *http.Request) {
	logger := *hlog.FromRequest(r)

	// gorilla mux Vars function returns the route variables for the
	// current request, if any. ID is the external id given for the resource
	vars := mux.Vars(r)
	extlID := vars["extlID"]

	response, err := s.OrgService.FindByExternalID(r.Context(), extlID)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
		return
	}
}

// handleAppCreate is a HandlerFunc used to create an App
func (s *Server) handleAppCreate(w http.ResponseWriter, r *http.Request) {
	lgr := *hlog.FromRequest(r)

	adt, err := audit.FromRequest(r)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	// Declare request body (rb)
	rb := new(service.CreateAppRequest)

	// Decode JSON HTTP request body into a Decoder type
	// and unmarshal that into the MovieRequest struct in the
	// AddMovieHandler
	err = json.NewDecoder(r.Body).Decode(&rb)
	defer r.Body.Close()
	// Call decoderErr to determine if body is nil, json is malformed
	// or any other error
	err = decoderErr(err)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	var response service.AppResponse
	response, err = s.AppService.Create(r.Context(), rb, adt)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, errs.E(errs.Internal, err))
		return
	}
}

// handleRegister is a HandlerFunc used to register a User
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	lgr := *hlog.FromRequest(r)

	adt, err := audit.FromRequest(r)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	err = s.RegisterUserService.SelfRegister(r.Context(), adt)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}
}

// handleLoggerRead handles GET requests for the /logger endpoint
func (s *Server) handleLoggerRead(w http.ResponseWriter, r *http.Request) {
	lgr := *hlog.FromRequest(r)

	response := s.LoggerService.Read()

	// Encode response struct to JSON for the response body
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, errs.E(errs.Internal, err))
		return
	}
}

// handleLoggerUpdate handles PUT requests for the /logger endpoint
// and updates the logger globals
func (s *Server) handleLoggerUpdate(w http.ResponseWriter, r *http.Request) {
	lgr := *hlog.FromRequest(r)

	// Declare rb as an instance of service.LoggerRequest
	rb := new(service.LoggerRequest)

	// Decode JSON HTTP request body into a json.Decoder type
	// and unmarshal that into rb
	err := json.NewDecoder(r.Body).Decode(&rb)
	defer r.Body.Close()
	// Call DecoderErr to determine if body is nil, json is malformed
	// or any other error
	err = decoderErr(err)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	var response service.LoggerResponse
	response, err = s.LoggerService.Update(rb)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, errs.E(errs.Internal, err))
		return
	}
}

// Ping handles GET requests for the /ping endpoint
func (s *Server) handlePing(w http.ResponseWriter, r *http.Request) {
	// pull logger from request context
	logger := *hlog.FromRequest(r)

	// pull the context from the http request
	ctx := r.Context()

	response := s.PingService.Ping(ctx, logger)

	// Encode response struct to JSON for the response body
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, logger, errs.E(errs.Internal, err))
		return
	}
}

// handleGenesis handles POST requests for the /genesis endpoint
func (s *Server) handleGenesis(w http.ResponseWriter, r *http.Request) {
	lgr := *hlog.FromRequest(r)

	// Declare rb as an instance of service.LoggerRequest
	rb := new(service.GenesisRequest)

	// Decode JSON HTTP request body into a json.Decoder type
	// and unmarshal that into rb
	err := json.NewDecoder(r.Body).Decode(&rb)
	defer r.Body.Close()
	// Call DecoderErr to determine if body is nil, json is malformed
	// or any other error
	err = decoderErr(err)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	var response service.GenesisResponse
	response, err = s.GenesisService.Seed(r.Context(), rb)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, errs.E(errs.Internal, err))
		return
	}
}

// handleGenesis handles GET requests for the /genesis endpoint
func (s *Server) handleGenesisRead(w http.ResponseWriter, r *http.Request) {
	lgr := *hlog.FromRequest(r)

	var (
		response service.GenesisResponse
		err      error
	)
	response, err = s.GenesisService.ReadConfig()
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, errs.E(errs.Internal, err))
		return
	}
}

// handlePermissionCreate handles POST requests for the /permission endpoint
func (s *Server) handlePermissionCreate(w http.ResponseWriter, r *http.Request) {
	lgr := *hlog.FromRequest(r)

	var (
		err error
		adt audit.Audit
	)
	adt, err = audit.FromRequest(r)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	// Declare rb as an instance of auth.Permission
	rb := new(service.PermissionRequest)

	// Decode JSON HTTP request body into a json.Decoder type
	// and unmarshal that into rb
	err = json.NewDecoder(r.Body).Decode(&rb)
	defer r.Body.Close()
	// Call DecoderErr to determine if body is nil, json is malformed
	// or any other error
	err = decoderErr(err)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	var response auth.Permission
	response, err = s.PermissionService.Create(r.Context(), rb, adt)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, errs.E(errs.Internal, err))
		return
	}
}

// handlePermissionFindAll handles GET requests for the /permission endpoint
func (s *Server) handlePermissionFindAll(w http.ResponseWriter, r *http.Request) {
	lgr := *hlog.FromRequest(r)

	response, err := s.PermissionService.FindAll(r.Context())
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, err)
		return
	}

	// Encode response struct to JSON for the response body
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		errs.HTTPErrorResponse(w, lgr, errs.E(errs.Internal, err))
		return
	}
}
