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
	wire.Build(
		newRouter,
		handler.NewAppHandler,
		app.NewApplication,
		newLogger,
		newLogLevel,
		datastore.NewDatastore,
		newDSName,
		newEnvName)
	return &mux.Router{}, nil
}
