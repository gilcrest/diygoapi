//+build wireinject

package main

import (
	"database/sql"

	"github.com/rs/zerolog"

	"github.com/gilcrest/go-api-basic/datastore/moviestore"
	"github.com/gilcrest/go-api-basic/datastore/pingstore"
	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/random"
	"github.com/gilcrest/go-api-basic/gateway/authgateway"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/handler"
	"go.opencensus.io/trace"

	"github.com/google/wire"
	"gocloud.dev/server"
	"gocloud.dev/server/driver"
	"gocloud.dev/server/health"
	"gocloud.dev/server/health/sqlhealth"
)

var pingHandlerSet = wire.NewSet(
	pingstore.NewDefaultPinger,
	wire.Bind(new(pingstore.Pinger), new(pingstore.DefaultPinger)),
	wire.Struct(new(handler.DefaultPingHandler), "*"),
	handler.NewPingHandler,
)

var middlewareSet = wire.NewSet(
	wire.Struct(new(authgateway.GoogleAccessTokenConverter), "*"),
	wire.Bind(new(auth.AccessTokenConverter), new(authgateway.GoogleAccessTokenConverter)),
	wire.Struct(new(auth.DefaultAuthorizer), "*"),
	wire.Bind(new(auth.Authorizer), new(auth.DefaultAuthorizer)),
	wire.Struct(new(handler.Middleware), "*"),
)

var movieHandlerSet = wire.NewSet(
	wire.Struct(new(random.DefaultStringGenerator), "*"),
	wire.Bind(new(random.StringGenerator), new(random.DefaultStringGenerator)),
	moviestore.NewDefaultTransactor,
	wire.Bind(new(moviestore.Transactor), new(moviestore.DefaultTransactor)),
	moviestore.NewDefaultSelector,
	wire.Bind(new(moviestore.Selector), new(moviestore.DefaultSelector)),
	wire.Struct(new(handler.DefaultMovieHandlers), "*"),
	handler.NewCreateMovieHandler,
	handler.NewFindMovieByIDHandler,
	handler.NewFindAllMoviesHandler,
	handler.NewUpdateMovieHandler,
	handler.NewDeleteMovieHandler,
)

var loggerHandlerSet = wire.NewSet(
	handler.NewReadLoggerHandler,
	handler.NewUpdateLoggerHandler,
)

var datastoreSet = wire.NewSet(
	datastore.NewDefaultDatastore,
	wire.Bind(new(datastore.Datastorer), new(datastore.DefaultDatastore)),
)

// goCloudServerSet
var goCloudOptionSet = wire.NewSet(
	trace.AlwaysSample,
	server.NewDefaultDriver,
	wire.Bind(new(driver.Server), new(*server.DefaultDriver)),
	wire.InterfaceValue(new(trace.Exporter), trace.Exporter(nil)),
	appHealthChecks,
	wire.Struct(new(server.Options), "HealthChecks", "TraceExporter", "DefaultSamplingPolicy", "Driver"),
)

func newHandlers(db *sql.DB) (handler.Handlers, error) {
	wire.Build(
		datastoreSet,
		movieHandlerSet,
		loggerHandlerSet,
		pingHandlerSet,
		wire.Struct(new(handler.Handlers), "*"),
	)
	return handler.Handlers{}, nil
}

func newMiddleware(lgr zerolog.Logger) handler.Middleware {
	wire.Build(middlewareSet)

	return handler.Middleware{}
}

func newServerOptions(db *sql.DB) (*server.Options, func()) {
	wire.Build(goCloudOptionSet)

	return nil, nil
}

// appHealthChecks returns a health check for the database. This will signal
// to Kubernetes or other orchestrators that the server should not receive
// traffic until the server is able to connect to its database.
func appHealthChecks(db *sql.DB) ([]health.Checker, func()) {
	dbCheck := sqlhealth.New(db)
	list := []health.Checker{dbCheck}
	return list, func() {
		dbCheck.Stop()
	}
}
