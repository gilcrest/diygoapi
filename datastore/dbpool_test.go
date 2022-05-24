package datastore

import (
	"context"
	"os"
	"testing"

	"github.com/gilcrest/diy-go-api/domain/logger"

	"github.com/rs/zerolog"
)

func Test_NewPostgreSQLPool(t *testing.T) {
	type args struct {
		ctx  context.Context
		pgds PostgreSQLDSN
		l    zerolog.Logger
	}

	ctx := context.Background()
	lgr := logger.NewLogger(os.Stdout, zerolog.DebugLevel, true)
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
			db, cleanup, err := NewPostgreSQLPool(tt.args.ctx, tt.args.pgds, tt.args.l)
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
