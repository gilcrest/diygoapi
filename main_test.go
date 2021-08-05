package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	qt "github.com/frankban/quicktest"
	"github.com/gilcrest/go-api-basic/domain/errs"
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

	type envLookup struct {
		value string
		ok    bool
	}

	type originalEnvs struct {
		logLevel      envLookup
		minLogLevel   envLookup
		logErrorStack envLookup
		port          envLookup
		dbhost        envLookup
		dbport        envLookup
		dbname        envLookup
		dbuser        envLookup
		dbpassword    envLookup
	}

	// get original state of environment variables before any test
	// modifies them
	ogEnvs := originalEnvs{}
	ogEnvs.logLevel.value, ogEnvs.logLevel.ok = os.LookupEnv(loglevelEnv)
	ogEnvs.minLogLevel.value, ogEnvs.minLogLevel.ok = os.LookupEnv(logLevelMinEnv)
	ogEnvs.logErrorStack.value, ogEnvs.logErrorStack.ok = os.LookupEnv(logErrorStackEnv)
	ogEnvs.port.value, ogEnvs.port.ok = os.LookupEnv(portEnv)
	ogEnvs.dbhost.value, ogEnvs.dbhost.ok = os.LookupEnv(dbHostEnv)
	ogEnvs.dbport.value, ogEnvs.dbport.ok = os.LookupEnv(dbPortEnv)
	ogEnvs.dbname.value, ogEnvs.dbname.ok = os.LookupEnv(dbNameEnv)
	ogEnvs.dbuser.value, ogEnvs.dbuser.ok = os.LookupEnv(dbUserEnv)
	ogEnvs.dbpassword.value, ogEnvs.dbpassword.ok = os.LookupEnv(dbPasswordEnv)

	if !ogEnvs.logLevel.ok {
		t.Logf("%s is not set to the environment", loglevelEnv)
	} else {
		t.Logf("original %s = %s", loglevelEnv, ogEnvs.logLevel.value)
	}
	if !ogEnvs.minLogLevel.ok {
		t.Logf("%s is not set to the environment", logLevelMinEnv)
	} else {
		t.Logf("original %s = %s", logLevelMinEnv, ogEnvs.logLevel.value)
	}
	if !ogEnvs.logErrorStack.ok {
		t.Logf("%s is not set to the environment", logErrorStackEnv)
	} else {
		t.Logf("original %s = %s", logErrorStackEnv, ogEnvs.logErrorStack.value)
	}
	if !ogEnvs.port.ok {
		t.Logf("%s is not set to the environment", portEnv)
	} else {
		t.Logf("original %s = %s", portEnv, ogEnvs.port.value)
	}
	if !ogEnvs.dbhost.ok {
		t.Logf("%s is not set to the environment", dbHostEnv)
	} else {
		t.Logf("original %s = %s", dbHostEnv, ogEnvs.dbhost.value)
	}
	if !ogEnvs.dbport.ok {
		t.Logf("%s is not set to the environment", dbPortEnv)
	} else {
		t.Logf("original %s = %s", dbPortEnv, ogEnvs.dbport.value)
	}
	if !ogEnvs.dbname.ok {
		t.Logf("%s is not set to the environment", dbNameEnv)
	} else {
		t.Logf("original %s = %s", dbNameEnv, ogEnvs.dbname.value)
	}
	if !ogEnvs.dbuser.ok {
		t.Logf("%s is not set to the environment", dbUserEnv)
	} else {
		t.Logf("original %s = %s", dbUserEnv, ogEnvs.dbuser.value)
	}
	if !ogEnvs.dbpassword.ok {
		t.Logf("%s is not set to the environment", dbPasswordEnv)
	} else {
		t.Logf("original %s = %s", dbPasswordEnv, ogEnvs.dbpassword.value)
	}

	emptyFunc := func() {}

	setEnvFunc := func() {
		t.Log("setting environment variables for test")
		os.Setenv(loglevelEnv, "warn")
		os.Setenv(logLevelMinEnv, "debug")
		os.Setenv(logErrorStackEnv, "false")
		os.Setenv(portEnv, "8081")
		os.Setenv(dbHostEnv, "hostwiththemost")
		os.Setenv(dbPortEnv, "5150")
		os.Setenv(dbNameEnv, "whatisinaname")
		os.Setenv(dbUserEnv, "usersarelosers")
		os.Setenv(dbPasswordEnv, "yeet")
		t.Log("Environment setup completed")
	}

	cleanupEnvFunc := func() {
		t.Log("resetting environment variables to original state from test")
		if ogEnvs.logLevel.ok {
			os.Setenv(loglevelEnv, ogEnvs.logLevel.value)
		}
		if ogEnvs.minLogLevel.ok {
			os.Setenv(logLevelMinEnv, ogEnvs.minLogLevel.value)
		}
		if ogEnvs.logErrorStack.ok {
			os.Setenv(logErrorStackEnv, ogEnvs.logErrorStack.value)
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
		t.Log("Environment cleanup completed")
	}

	a1 := args{args: []string{"server", "-log-level=info", "-log-level-min=debug", "-log-error-stack", "-port=8080", "-db-host=localhost", "-db-port=5432", "-db-name=go_api_basic", "-db-user=postgres", "-db-password=sosecret"}}
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
		name        string
		args        args
		envFunc     func()
		cleanupFunc func()
		wantFlgs    flags
		wantErr     bool
	}{
		{"all flags", a1, emptyFunc, emptyFunc, f1, false},
		{"min level flag", a5, emptyFunc, emptyFunc, f5, false},
		{"invalid flag", a4, emptyFunc, emptyFunc, f4, true},
		{"use environment", a2, setEnvFunc, cleanupEnvFunc, f2, false},
		{"mix flags and env", a3, setEnvFunc, cleanupEnvFunc, f3, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.envFunc()
			gotFlgs, err := newFlags(tt.args.args)
			c.Assert(gotFlgs, qt.Equals, tt.wantFlgs)
			if (err != nil) != tt.wantErr {
				t.Errorf("newFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
			tt.cleanupFunc()
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
