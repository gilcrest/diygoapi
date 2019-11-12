//+build wireinject

package main

import (
	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/google/wire"
	"github.com/gorilla/mux"
)

func setupRouter(flags *cliFlags) (*mux.Router, error) {
	wire.Build(provideLogLevel,
		provideLogger,
		provideName,
		provideDBName,
		datastore.ProvideDB,
		datastore.ProvideDS,
		app.ProvideApplication,
		provideRouter)
	return &mux.Router{}, nil
}
