package server

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/gilcrest/diy-go-api/domain/app"
	"github.com/gilcrest/diy-go-api/domain/audit"
	"github.com/gilcrest/diy-go-api/domain/auth"
	"github.com/gilcrest/diy-go-api/domain/user"
	"github.com/gilcrest/diy-go-api/service"
)

// CreateMovieService creates a Movie
type CreateMovieService interface {
	Create(ctx context.Context, r *service.CreateMovieRequest, adt audit.Audit) (service.MovieResponse, error)
}

// UpdateMovieService is a service for updating a Movie
type UpdateMovieService interface {
	Update(ctx context.Context, r *service.UpdateMovieRequest, adt audit.Audit) (service.MovieResponse, error)
}

// DeleteMovieService is a service for deleting a Movie
type DeleteMovieService interface {
	Delete(ctx context.Context, extlID string) (service.DeleteResponse, error)
}

// FindMovieService interface reads a Movie form the database
type FindMovieService interface {
	FindMovieByID(ctx context.Context, extlID string) (service.MovieResponse, error)
	FindAllMovies(ctx context.Context) ([]service.MovieResponse, error)
}

// CreateOrgService manages the creation of an Org (and optional app)
type CreateOrgService interface {
	Create(ctx context.Context, r *service.CreateOrgRequest, adt audit.Audit) (service.OrgResponse, error)
}

// OrgService manages the retrieval and manipulation of an Org
type OrgService interface {
	Update(ctx context.Context, r *service.UpdateOrgRequest, adt audit.Audit) (service.OrgResponse, error)
	Delete(ctx context.Context, extlID string) (service.DeleteResponse, error)
	FindAll(ctx context.Context) ([]service.OrgResponse, error)
	FindByExternalID(ctx context.Context, extlID string) (service.OrgResponse, error)
}

// AppService manages the retrieval and manipulation of an App
type AppService interface {
	Create(ctx context.Context, r *service.CreateAppRequest, adt audit.Audit) (service.AppResponse, error)
	Update(ctx context.Context, r *service.UpdateAppRequest, adt audit.Audit) (service.AppResponse, error)
}

// MiddlewareService are all the services uses by the various middleware functions
type MiddlewareService interface {
	// FindAppByAPIKey finds an app given its External ID and determines
	// if the given API key is a valid key for it
	FindAppByAPIKey(ctx context.Context, realm, appExtlID, apiKey string) (app.App, error)
	// FindUserByOauth2Token retrieves a User given an Oauth2 token
	FindUserByOauth2Token(ctx context.Context, params service.FindUserParams) (user.User, error)
	// Authorize determines whether an app/user (as part of an Audit
	// struct) can perform an action against a resource
	Authorize(lgr zerolog.Logger, r *http.Request, sub audit.Audit) error
}

// PermissionService allows for creating, updating, reading and deleting a Permission
type PermissionService interface {
	Create(ctx context.Context, r *service.PermissionRequest, adt audit.Audit) (auth.Permission, error)
	FindAll(ctx context.Context) ([]auth.Permission, error)
}

// RoleService allows for creating, updating, reading and deleting a Role
// as well as assigning permissions and users to it.
type RoleService interface {
	Create(ctx context.Context, r *auth.Role, adt audit.Audit) (auth.Role, error)
	//AddPermissions(ctx context.Context, r *[]auth.Permission, adt audit.Audit) error
}

// RegisterUserService registers a new user
type RegisterUserService interface {
	SelfRegister(ctx context.Context, adt audit.Audit) error
}

// LoggerService reads and updates the logger state
type LoggerService interface {
	Read() service.LoggerResponse
	Update(r *service.LoggerRequest) (service.LoggerResponse, error)
}

// PingService pings the database and responds whether it is up or down
type PingService interface {
	Ping(ctx context.Context, logger zerolog.Logger) service.PingResponse
}

// GenesisService initializes the database with dependent data
type GenesisService interface {
	// Seed initializes required dependent data in database
	Seed(ctx context.Context, r *service.GenesisRequest) (service.GenesisResponse, error)
	// ReadConfig reads the local config file generated as part of Seed (when run locally).
	// Is only a utility to help with local testing.
	ReadConfig() (service.GenesisResponse, error)
}

// Services are used by the application service handlers
type Services struct {
	CreateMovieService  CreateMovieService
	UpdateMovieService  UpdateMovieService
	DeleteMovieService  DeleteMovieService
	FindMovieService    FindMovieService
	CreateOrgService    CreateOrgService
	OrgService          OrgService
	AppService          AppService
	RegisterUserService RegisterUserService
	PingService         PingService
	LoggerService       LoggerService
	GenesisService      GenesisService
	MiddlewareService   MiddlewareService
	PermissionService   PermissionService
	RoleService         RoleService
}
