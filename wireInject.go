//+build wireinject

package main

import (
	"github.com/gilcrest/go-api-basic/app"
	"github.com/gilcrest/go-api-basic/datastore"
	"github.com/gilcrest/go-api-basic/handler"
	"github.com/google/wire"
	"github.com/gorilla/mux"
)

func setupRouter(flags *cliFlags) (*mux.Router, error) {
	wire.Build(provideLogLevel,
		provideLogger,
		provideEnvName,
		provideDSName,
		datastore.NewDatastore,
		app.NewApplication,
		handler.NewAppHandler,
		provideRouter)
	return &mux.Router{}, nil
}
