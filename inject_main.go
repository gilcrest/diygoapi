//+build wireinject

package main

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/handler"
	"github.com/google/wire"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog"
	"go.opencensus.io/trace"
	"gocloud.dev/server"
	"gocloud.dev/server/driver"
	"gocloud.dev/server/health"
	"gocloud.dev/server/health/sqlhealth"
)

// applicationSet is the Wire provider set for the application
var applicationSet = wire.NewSet(
	app.NewApplication,
	newRouter,
	wire.Bind(new(http.Handler), new(*mux.Router)),
	handler.NewAppHandler,
	app.NewLogger,
)

// goCloudServerSet
var goCloudServerSet = wire.NewSet(
	trace.AlwaysSample,
	server.New,
	server.NewDefaultDriver,
	wire.Bind(new(driver.Server), new(*server.DefaultDriver)),
)

// newServer is a Wire injector function that sets up the
// application using a PostgreSQL implementation
func newServer(ctx context.Context, loglvl zerolog.Level) (*server.Server, func(), error) {
	// This will be filled in by Wire with providers from the provider sets in
	// wire.Build.
	wire.Build(
		wire.InterfaceValue(new(trace.Exporter), trace.Exporter(nil)),
		goCloudServerSet,
		applicationSet,
		appHealthChecks,
		wire.Struct(new(server.Options), "HealthChecks", "TraceExporter", "DefaultSamplingPolicy", "Driver"),
		datastore.NewPGDatasourceName,
		datastore.NewDB,
		wire.Bind(new(datastore.Datastorer), new(*datastore.Datastore)),
		datastore.NewDatastore)
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
