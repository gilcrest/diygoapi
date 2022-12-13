package cmd

import (
	"fmt"
	qt "github.com/frankban/quicktest"
	"github.com/google/go-cmp/cmp"
	"testing"

	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/sqldb"
)

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
		{"port < 0", args{port: -1}, errs.E(fmt.Sprintf("port %d is not within valid port range (0 to 65535)", -1))},
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

	setEnvFunc := func() {
		c.Log("setting environment variables for test")
		c.Setenv(loglevelEnv, "warn")
		c.Setenv(logLevelMinEnv, "debug")
		c.Setenv(logErrorStackEnv, "false")
		c.Setenv(portEnv, "8081")
		c.Setenv(sqldb.DBHostEnv, "hostwiththemost")
		c.Setenv(sqldb.DBPortEnv, "5150")
		c.Setenv(sqldb.DBNameEnv, "whatisinaname")
		c.Setenv(sqldb.DBUserEnv, "usersarelosers")
		c.Setenv(sqldb.DBPasswordEnv, "yeet")
		c.Setenv(sqldb.DBSearchPathEnv, "u2")
		c.Setenv(encryptKeyEnv, "reallyGoodKey")
		c.Log("Environment setup completed")
	}

	setEnv2EmptyFunc := func() {
		c.Log("setting environment variables for test")
		c.Setenv(loglevelEnv, "")
		c.Setenv(logLevelMinEnv, "")
		c.Setenv(logErrorStackEnv, "")
		c.Setenv(portEnv, "")
		c.Setenv(sqldb.DBHostEnv, "")
		c.Setenv(sqldb.DBPortEnv, "")
		c.Setenv(sqldb.DBNameEnv, "")
		c.Setenv(sqldb.DBUserEnv, "")
		c.Setenv(sqldb.DBPasswordEnv, "")
		c.Setenv(sqldb.DBSearchPathEnv, "")
		c.Setenv(encryptKeyEnv, "")
		c.Log("Environment setup completed")
	}

	a1 := args{args: []string{"server", "-log-level=info", "-log-level-min=debug", "-log-error-stack", "-port=8080", "-db-host=localhost", "-db-port=5432", "-db-name=go_api_basic", "-db-user=postgres", "-db-password=sosecret", "-db-search-path=demo", "-encrypt-key=reallyGoodKey"}}
	f1 := flags{
		loglvl:        "info",
		logLvlMin:     "debug",
		logErrorStack: true,
		port:          8080,
		dbhost:        "localhost",
		dbport:        5432,
		dbname:        "go_api_basic",
		dbuser:        "postgres",
		dbpassword:    "sosecret",
		dbsearchpath:  "demo",
		encryptkey:    "reallyGoodKey",
	}

	a2 := args{args: []string{"server"}}
	f2 := flags{
		loglvl:        "warn",
		logLvlMin:     "debug",
		logErrorStack: false,
		port:          8081,
		dbhost:        "hostwiththemost",
		dbport:        5150,
		dbname:        "whatisinaname",
		dbuser:        "usersarelosers",
		dbpassword:    "yeet",
		dbsearchpath:  "u2",
		encryptkey:    "reallyGoodKey",
	}

	a3 := args{args: []string{"server", "-log-level=error"}}
	f3 := flags{
		loglvl:        "error",
		logLvlMin:     "debug",
		logErrorStack: false,
		port:          8081,
		dbhost:        "hostwiththemost",
		dbport:        5150,
		dbname:        "whatisinaname",
		dbuser:        "usersarelosers",
		dbpassword:    "yeet",
		dbsearchpath:  "u2",
		encryptkey:    "reallyGoodKey",
	}

	a4 := args{args: []string{"server", "-badflag=true"}}
	f4 := flags{}

	a5 := args{args: []string{"server", "-log-level=debug", "-log-level-min=debug", "-log-error-stack", "-port=8080", "-db-host=localhost", "-db-port=5432", "-db-name=go_api_basic", "-db-user=postgres", "-db-password=sosecret"}}
	f5 := flags{
		loglvl:        "debug",
		logLvlMin:     "debug",
		logErrorStack: true,
		port:          8080,
		dbhost:        "localhost",
		dbport:        5432,
		dbname:        "go_api_basic",
		dbuser:        "postgres",
		dbpassword:    "sosecret",
	}

	tests := []struct {
		name     string
		args     args
		envFunc  func()
		wantFlgs flags
		wantErr  bool
	}{
		{"all flags", a1, func() {}, f1, false},
		{"min level flag", a5, setEnv2EmptyFunc, f5, false},
		{"invalid flag", a4, func() {}, f4, true},
		{"use environment", a2, setEnvFunc, f2, false},
		{"mix flags and env", a3, setEnvFunc, f3, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.envFunc()
			gotFlgs, err := newFlags(tt.args.args)
			c.Assert(gotFlgs, qt.Equals, tt.wantFlgs)
			if (err != nil) != tt.wantErr {
				t.Errorf("newFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
