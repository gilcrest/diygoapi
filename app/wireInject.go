//+build wireinject

package main

import (
	"github.com/google/wire"
)

func setupApplication(flags *cliFlags) (*application, error) {
	wire.Build(provideLogLevel,
		provideLogger,
		provideName,
		provideAppDB,
		provideRouter,
		provideApplication)
	return &application{}, nil
}
