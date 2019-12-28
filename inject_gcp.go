//+build wireinject

package main


// // GCP is a Wire provider set that includes all Google Cloud Platform services
// // in this repository and authenticates using Application Default Credentials.
// var gcpSet = wire.NewSet(gcpServicesSet, gcp.DefaultIdentity)

// // Services is a Wire provider set that includes the default wiring for all
// // Google Cloud Platform services in this repository, but does not include
// // credentials. Individual services may require additional configuration.
// var gcpServicesSet = wire.NewSet(
// 	gcp.DefaultTransport,
// 	gcp.NewHTTPClient,
// 	gcpruntimeconfig.Set,
// 	gcpkms.Set,
// 	gcppubsub.Set,
// 	gcsblob.Set,
// 	cloudsql.CertSourceSet,
// 	gcpfirestore.Set,
// 	sdserver.Set,
// )

// // setupGCP is a Wire injector function that sets up the application using GCP.
// func setupGCP(ctx context.Context, envName app.EnvName, dsName datastore.DSName, loglvl zerolog.Level) (*server.Server, func(), error) {
// 	// This will be filled in by Wire with providers from the provider sets in
// 	// wire.Build.
// 	wire.Build(
// 		wire.InterfaceValue(new(trace.Exporter), trace.Exporter(nil)),
// 		goCloudServerSet,
// 		applicationSet,
// 		gcpSet,
// 		wire.Struct(new(gcpmysql.URLOpener), "CertSource"),
// 		applicationSet,
// 		datastore.OpenGCPDatabase,
// 	)
// 	return nil, nil, nil
// }
