package sqldb

import (
	"context"
	"os"
	"strconv"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/rs/zerolog"

	"github.com/gilcrest/diygoapi/logger"
)

func Test_NewPostgreSQLPool(t *testing.T) {
	type args struct {
		ctx  context.Context
		pgds PostgreSQLDSN
		l    zerolog.Logger
	}

	ctx := context.Background()
	lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)
	dsn := newPostgreSQLDSN(t)
	baddsn := PostgreSQLDSN{
		Host:   "badhost",
		Port:   5432,
		DBName: "go_api_basic",
		User:   "postgres",
	}

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"App DB", args{ctx, dsn, lgr}, false},
		{"Bad DSN", args{ctx, baddsn, lgr}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanup, err := NewPostgreSQLPool(tt.args.ctx, tt.args.l, tt.args.pgds)
			t.Cleanup(cleanup)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				err = db.Ping(ctx)
				if err != nil {
					t.Errorf("Error pinging database = %v", err)
				}
			}
		})
	}
}

func TestNewDB(t *testing.T) {
	t.Run("typical", func(t *testing.T) {
		c := qt.New(t)

		ctx := context.Background()
		lgr := logger.New(os.Stdout, zerolog.DebugLevel, true)

		dsn := newPostgreSQLDSN(t)

		dbpool, cleanup, err := NewPostgreSQLPool(ctx, lgr, dsn)
		c.Assert(err, qt.IsNil)
		t.Cleanup(cleanup)

		got := NewDB(dbpool)
		want := DB{pool: dbpool}

		c.Assert(*got, qt.Equals, want)
	})
}

// newPGDatasourceName is a test helper to get a PGDatasourceName
// from environment variables
func newPostgreSQLDSN(t *testing.T) PostgreSQLDSN {
	t.Helper()

	var (
		dbHost       string
		dbPort       int
		dbName       string
		dbUser       string
		dbPassword   string
		dbSearchPath string
		ok           bool
		err          error
	)

	dbHost, ok = os.LookupEnv(DBHostEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", DBHostEnv)
	}

	var p string
	p, ok = os.LookupEnv(DBPortEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", DBPortEnv)
	}
	dbPort, err = strconv.Atoi(p)
	if err != nil {
		t.Fatalf("Unable to convert db port %s to int", p)
	}

	dbName, ok = os.LookupEnv(DBNameEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", DBNameEnv)
	}

	dbUser, ok = os.LookupEnv(DBUserEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", DBUserEnv)
	}

	dbPassword, ok = os.LookupEnv(DBPasswordEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", DBPasswordEnv)
	}

	dbSearchPath, ok = os.LookupEnv(DBSearchPathEnv)
	if !ok {
		t.Fatalf("No environment variable found for %s", DBSearchPathEnv)
	}

	return PostgreSQLDSN{
		Host:       dbHost,
		Port:       dbPort,
		DBName:     dbName,
		SearchPath: dbSearchPath,
		User:       dbUser,
		Password:   dbPassword,
	}
}
