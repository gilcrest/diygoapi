//+build wireinject

package main

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/gilcrest/go-api-basic/domain/random"

	"github.com/gilcrest/go-api-basic/datastore/moviestore"
	"github.com/gilcrest/go-api-basic/datastore/pingstore"

	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/gateway/authgateway"

	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/handler"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
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
	handler.ProvidePingHandler,
)

var movieHandlerSet = wire.NewSet(
	wire.Struct(new(random.DefaultStringGenerator), "*"),
	wire.Bind(new(random.StringGenerator), new(random.DefaultStringGenerator)),
	wire.Struct(new(authgateway.GoogleAccessTokenConverter), "*"),
	wire.Bind(new(auth.AccessTokenConverter), new(authgateway.GoogleAccessTokenConverter)),
	wire.Struct(new(auth.DefaultAuthorizer), "*"),
	wire.Bind(new(auth.Authorizer), new(auth.DefaultAuthorizer)),
	moviestore.NewDefaultTransactor,
	wire.Bind(new(moviestore.Transactor), new(moviestore.DefaultTransactor)),
	moviestore.NewDefaultSelector,
	wire.Bind(new(moviestore.Selector), new(moviestore.DefaultSelector)),
	wire.Struct(new(handler.DefaultMovieHandlers), "*"),
	handler.ProvideCreateMovieHandler,
	handler.ProvideFindMovieByIDHandler,
	handler.ProvideFindAllMoviesHandler,
	handler.ProvideUpdateMovieHandler,
	handler.ProvideDeleteMovieHandler,
)

var loggerHandlerSet = wire.NewSet(
	wire.Struct(new(handler.DefaultLoggerHandlers), "*"),
	handler.NewReadLoggerHandler,
	handler.NewUpdateLoggerHandler,
)

var datastoreSet = wire.NewSet(
	datastore.NewDB,
	datastore.NewDefaultDatastore,
	wire.Bind(new(datastore.Datastorer), new(datastore.DefaultDatastore)),
)

// goCloudServerSet
var goCloudServerSet = wire.NewSet(
	trace.AlwaysSample,
	server.New,
	server.NewDefaultDriver,
	wire.Bind(new(driver.Server), new(*server.DefaultDriver)),
)

var routerSet = wire.NewSet(
	handler.NewMuxRouter,
	wire.Bind(new(http.Handler), new(*mux.Router)),
)

// newServer is a Wire injector function that sets up the
// application using a PostgreSQL implementation
func newServer(ctx context.Context, logger zerolog.Logger, dsn datastore.PGDatasourceName) (*server.Server, func(), error) {
	// This will be filled in by Wire with providers from the provider sets in
	// wire.Build.
	wire.Build(
		wire.InterfaceValue(new(trace.Exporter), trace.Exporter(nil)),
		goCloudServerSet,
		appHealthChecks,
		wire.Struct(new(server.Options), "HealthChecks", "TraceExporter", "DefaultSamplingPolicy", "Driver"),
		datastoreSet,
		movieHandlerSet,
		loggerHandlerSet,
		pingHandlerSet,
		wire.Struct(new(handler.Handlers), "*"),
		routerSet,
	)
	return nil, nil, nil
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
