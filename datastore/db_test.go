package datastore

import (
	"os"
	"testing"

	"github.com/gilcrest/go-api-basic/domain/logger"

	"github.com/rs/zerolog"
)

func Test_NewDB(t *testing.T) {
	type args struct {
		pgds PGDatasourceName
		l    zerolog.Logger
	}

	lgr := logger.NewLogger(os.Stdout, true)
	dsn := NewPGDatasourceName("localhost", "go_api_basic", "postgres", "", 5432)
	baddsn := NewPGDatasourceName("badhost", "go_api_basic", "postgres", "", 5432)

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"App DB", args{dsn, lgr}, false},
		{"Bad DSN", args{baddsn, lgr}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, cleanup, err := NewDB(tt.args.pgds, tt.args.l)
			defer cleanup()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				err = db.Ping()
				if err != nil {
					t.Errorf("Error pinging database = %v", err)
				}
			}
		})
	}
}
