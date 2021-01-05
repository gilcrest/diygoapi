package datastore

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
)

func Test_NewDB(t *testing.T) {
	type args struct {
		pgds PGDatasourceName
		l    zerolog.Logger
	}

	// empty string for TimeFieldFormat will write logs with UNIX time
	zerolog.TimeFieldFormat = ""
	// set logging level based on input
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	// start a new logger with Stdout as the target
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	dsn := NewPGDatasourceName("localhost", "go_api_basic", "postgres", "", 5432)

	tests := []struct {
		name string
		args args
	}{
		{"App DB", args{dsn, logger}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _, err := NewDB(tt.args.pgds, tt.args.l)
			if err != nil {
				t.Errorf("Error from newDB = %v", err)
			}
			err = db.Ping()
			if err != nil {
				t.Errorf("Error pinging database = %v", err)
			}
		})
	}
}
