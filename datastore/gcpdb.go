package datastore

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"gocloud.dev/gcp"
	"gocloud.dev/postgres/gcppostgres"
)

// OpenGCPDatabase is a Wire provider function that connects to a GCP Cloud SQL
// MySQL database based on the command-line flags.
func OpenGCPDatabase(ctx context.Context, opener *gcppostgres.URLOpener, id gcp.ProjectID, n DSName) (*sql.DB, func(), error) {
	const op errs.Op = "datastore/OpenGCPDatabase"

	dbEnvMap, err := dbEnv(n)
	if err != nil {
		return nil, nil, errs.E(op, err)
	}

	db, err := opener.OpenPostgresURL(ctx, &url.URL{
		Scheme: "gcppostgres",
		User:   url.UserPassword(dbEnvMap["user"], dbEnvMap["password"]),
		Host:   string(id),
		Path:   fmt.Sprintf("/%s/%s/%s", "us-east1", dbEnvMap["host"], dbEnvMap["dbname"]),
	})
	if err != nil {
		return nil, nil, err
	}
	return db, func() { db.Close() }, nil
}
