package datastore

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
)

func Test_NewLocalDB(t *testing.T) {
	type args struct {
		n Name
		l zerolog.Logger
	}

	// empty string for TimeFieldFormat will write logs with UNIX time
	zerolog.TimeFieldFormat = ""
	// set logging level based on input
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	// start a new logger with Stdout as the target
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	tests := []struct {
		name string
		args args
	}{
		{"App DB", args{LocalDatastore, logger}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, _, err := NewDB(tt.args.n, tt.args.l)
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
