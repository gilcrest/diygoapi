package app

import (
	"context"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/domain/app"
	"github.com/gilcrest/go-api-basic/domain/audit"
	"github.com/gilcrest/go-api-basic/domain/user"
	"github.com/gilcrest/go-api-basic/service"
)

//// CreateMovieService creates a Movie
//type CreateMovieService interface {
//	Create(ctx context.Context, r *service.CreateMovieRequest, u user.User) (service.MovieResponse, error)
//}
//
//// UpdateMovieService is a service for updating a Movie
//type UpdateMovieService interface {
//	Update(ctx context.Context, r *service.UpdateMovieRequest, u user.User) (service.MovieResponse, error)
//}
//
//// DeleteMovieService is a service for deleting a Movie
//type DeleteMovieService interface {
//	Delete(ctx context.Context, extlID string) (service.DeleteMovieResponse, error)
//}
//
//// FindMovieService interface reads a Movie form the database
//type FindMovieService interface {
//	FindMovieByID(ctx context.Context, extlID string) (service.MovieResponse, error)
//	FindAllMovies(ctx context.Context) ([]service.MovieResponse, error)
//}

// CreateOrgService creates an Org
type CreateOrgService interface {
	Create(ctx context.Context, r *service.CreateOrgRequest, adt audit.Audit) (service.OrgResponse, error)
}

// UpdateOrgService updates an Org
type UpdateOrgService interface {
	Update(ctx context.Context, r *service.UpdateOrgRequest, adt audit.Audit) (service.OrgResponse, error)
}

// FindOrgService retrieves Org information from the datastore
type FindOrgService interface {
	FindAll(ctx context.Context) ([]service.OrgResponse, error)
	FindByExternalID(ctx context.Context, extlID string) (service.OrgResponse, error)
}

// CreateAppService creates an App
type CreateAppService interface {
	Create(ctx context.Context, r *service.CreateAppRequest, adt audit.Audit) (service.AppResponse, error)
}

// FindAppService retrieves an App
type FindAppService interface {
	// FindAppByAPIKey finds an app given its External ID and determines
	// if the given API key is a valid key for it
	FindAppByAPIKey(ctx context.Context, realm, appExtlID, apiKey string) (app.App, error)
}

// FindUserService retrieves a User
type FindUserService interface {
	FindUserByOauth2Token(ctx context.Context, params service.FindUserParams) (user.User, error)
}

// AuthorizeService determines whether an app and a user can perform
// an action against a resource
type AuthorizeService interface {
	Authorize(lgr zerolog.Logger, r *http.Request, sub audit.Audit) error
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

// SeedService initializes the database with dependent data
type SeedService interface {
	Seed(ctx context.Context, r *service.SeedRequest) (service.SeedResponse, error)
}

// Services are used by the application service handlers
type Services struct {
	//CreateMovieService CreateMovieService
	//UpdateMovieService UpdateMovieService
	//DeleteMovieService DeleteMovieService
	//FindMovieService   FindMovieService
	SeedService      SeedService
	PingService      PingService
	LoggerService    LoggerService
	CreateOrgService CreateOrgService
	UpdateOrgService UpdateOrgService
	FindOrgService   FindOrgService
	CreateAppService CreateAppService
}

// MiddlewareServices are used by middleware handlers
type MiddlewareServices struct {
	FindAppService   FindAppService
	FindUserService  FindUserService
	AuthorizeService AuthorizeService
}
