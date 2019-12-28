package datastore

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"

	"gocloud.dev/gcp"
	"gocloud.dev/postgres/gcppostgres"
)

// OpenGCPDatabase is a Wire provider function that connects to a GCP Cloud SQL
// MySQL database based on the command-line flags.
func OpenGCPDatabase(ctx context.Context, opener *gcppostgres.URLOpener, id gcp.ProjectID, n DSName) (*sql.DB, func(), error) {
	db, err := opener.OpenPostgresURL(ctx, &url.URL{
		Scheme: "gcppostgres",
		User:   url.UserPassword(os.Getenv(dbEnv(n, "user")), dbEnv(n, "password")),
		Host:   string(id),
		Path:   fmt.Sprintf("/%s/%s/%s", "us-east1", dbEnv(n, "host"), dbEnv(n, "dbname")),
	})
	if err != nil {
		return nil, nil, err
	}
	return db, func() { db.Close() }, nil
}
