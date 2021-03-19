package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/pkg/errors"

	qt "github.com/frankban/quicktest"

	"github.com/rs/zerolog"
)

func Test_newLogLevel(t *testing.T) {
	c := qt.New(t)

	type args struct {
		loglvl string
	}
	tests := []struct {
		name string
		args args
		want zerolog.Level
	}{
		{"debug", args{loglvl: "debug"}, zerolog.DebugLevel},
		{"info", args{loglvl: "info"}, zerolog.InfoLevel},
		{"warn", args{loglvl: "warn"}, zerolog.WarnLevel},
		{"error", args{loglvl: "error"}, zerolog.ErrorLevel},
		{"fatal", args{loglvl: "fatal"}, zerolog.FatalLevel},
		{"panic", args{loglvl: "panic"}, zerolog.PanicLevel},
		{"disabled", args{loglvl: "disabled"}, zerolog.Disabled},
		{"default", args{loglvl: ""}, zerolog.InfoLevel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newLogLevel(tt.args.loglvl)
			c.Assert(got, qt.Equals, tt.want)
		})
	}
}

func Test_portRange(t *testing.T) {
	c := qt.New(t)

	type args struct {
		port int
	}

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{"valid port", args{port: 5432}, nil},
		{"port < 0", args{port: -1}, errs.E(errors.New(fmt.Sprintf("port %d is not within valid port range (0 to 65535)", -1)))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := portRange(tt.args.port)
			c.Assert(err, qt.CmpEquals(cmp.Comparer(errs.Match)), tt.wantErr)
		})
	}
}

func Test_newFlags(t *testing.T) {
	c := qt.New(t)

	type args struct {
		args []string
	}

	a1 := args{args: []string{"server", "-log-level=debug", "-port=8080", "-db-host=localhost", "-db-port=5432", "-db-name=go_api_basic", "-db-user=postgres", "-db-password=sosecret"}}

	f1 := flags{
		loglvl:     "debug",
		port:       8080,
		dbhost:     "localhost",
		dbport:     5432,
		dbname:     "go_api_basic",
		dbuser:     "postgres",
		dbpassword: "sosecret",
	}

	type envLookup struct {
		value string
		ok    bool
	}

	type originalEnvs struct {
		logLevel   envLookup
		port       envLookup
		dbhost     envLookup
		dbport     envLookup
		dbname     envLookup
		dbuser     envLookup
		dbpassword envLookup
	}

	const (
		loglevelEnv   string = "LOG_LEVEL"
		portEnv       string = "PORT"
		dbHostEnv     string = "DB_HOST"
		dbPortEnv     string = "DB_PORT"
		dbNameEnv     string = "DB_NAME"
		dbUserEnv     string = "DB_USER"
		dbPasswordEnv string = "DB_PASSWORD"
	)

	ogEnvs := new(originalEnvs)
	ogEnvs.logLevel.value, ogEnvs.logLevel.ok = os.LookupEnv(loglevelEnv)
	ogEnvs.port.value, ogEnvs.port.ok = os.LookupEnv(portEnv)
	ogEnvs.dbhost.value, ogEnvs.dbhost.ok = os.LookupEnv(dbHostEnv)
	ogEnvs.dbport.value, ogEnvs.dbport.ok = os.LookupEnv(dbPortEnv)
	ogEnvs.dbname.value, ogEnvs.dbname.ok = os.LookupEnv(dbNameEnv)
	ogEnvs.dbuser.value, ogEnvs.dbuser.ok = os.LookupEnv(dbUserEnv)
	ogEnvs.dbpassword.value, ogEnvs.dbpassword.ok = os.LookupEnv(dbPasswordEnv)

	t.Logf("original %s = %s", loglevelEnv, ogEnvs.logLevel.value)
	t.Logf("original %s = %s", portEnv, ogEnvs.port.value)
	t.Logf("original %s = %s", dbHostEnv, ogEnvs.dbhost.value)
	t.Logf("original %s = %s", dbPortEnv, ogEnvs.dbport.value)
	t.Logf("original %s = %s", dbNameEnv, ogEnvs.dbname.value)
	t.Logf("original %s = %s", dbUserEnv, ogEnvs.dbuser.value)
	t.Logf("original %s = %s", dbPasswordEnv, ogEnvs.dbpassword.value)

	os.Setenv(loglevelEnv, "warn")
	os.Setenv(portEnv, "8081")
	os.Setenv(dbHostEnv, "hostwiththemost")
	os.Setenv(dbPortEnv, "5150")
	os.Setenv(dbNameEnv, "whatisinaname")
	os.Setenv(dbUserEnv, "usersarelosers")
	os.Setenv(dbPasswordEnv, "yeet")

	cleanup := func() {
		if ogEnvs.logLevel.ok {
			os.Setenv(loglevelEnv, ogEnvs.logLevel.value)
		}
		if ogEnvs.port.ok {
			os.Setenv(portEnv, ogEnvs.port.value)
		}
		if ogEnvs.dbhost.ok {
			os.Setenv(dbHostEnv, ogEnvs.dbhost.value)
		}
		if ogEnvs.dbport.ok {
			os.Setenv(dbPortEnv, ogEnvs.dbport.value)
		}
		if ogEnvs.dbname.ok {
			os.Setenv(dbNameEnv, ogEnvs.dbname.value)
		}
		if ogEnvs.dbuser.ok {
			os.Setenv(dbUserEnv, ogEnvs.dbuser.value)
		}
		if ogEnvs.dbpassword.ok {
			os.Setenv(dbPasswordEnv, ogEnvs.dbpassword.value)
		}
	}
	t.Cleanup(cleanup)

	a2 := args{args: []string{"server"}}
	f2 := flags{
		loglvl:     "warn",
		port:       8081,
		dbhost:     "hostwiththemost",
		dbport:     5150,
		dbname:     "whatisinaname",
		dbuser:     "usersarelosers",
		dbpassword: "yeet",
	}

	a3 := args{args: []string{"server", "-log-level=error"}}
	f3 := flags{
		loglvl:     "error",
		port:       8081,
		dbhost:     "hostwiththemost",
		dbport:     5150,
		dbname:     "whatisinaname",
		dbuser:     "usersarelosers",
		dbpassword: "yeet",
	}

	a4 := args{args: []string{"server", "-badflag=true"}}
	f4 := flags{}

	tests := []struct {
		name     string
		args     args
		wantFlgs flags
		wantErr  bool
	}{
		{"all flags", a1, f1, false},
		{"use environment", a2, f2, false},
		{"mix flags and env", a3, f3, false},
		{"invalid flag", a4, f4, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFlgs, err := newFlags(tt.args.args)
			c.Assert(gotFlgs, qt.Equals, tt.wantFlgs)
			if (err != nil) != tt.wantErr {
				t.Errorf("newFlags() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
}

func Test_run(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := run(tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
