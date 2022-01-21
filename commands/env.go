package commands

import "os"

// OverrideEnv sets the environment for specific fields.
//
// This is a farce and only to be used when developing locally.
//
// You should use Cloud Secrets or something like that and set environment vars
// dynamically through those at deployment/run time.
//
// This file should not be included in your git repository and should
// be added to .gitignore. I have included it in this repository since
// this is purely demonstrative and the password and encryption keys are
// invalid.
func OverrideEnv() error {
	var err error

	// minimum accepted log level
	err = os.Setenv(logLevelMinEnv, "trace")
	if err != nil {
		return err
	}

	// log level
	err = os.Setenv(loglevelEnv, "debug")
	if err != nil {
		return err
	}

	// log error stack
	err = os.Setenv(logErrorStackEnv, "true")
	if err != nil {
		return err
	}

	// server port
	err = os.Setenv(portEnv, "8080")
	if err != nil {
		return err
	}

	// database host
	err = os.Setenv(dbHostEnv, "localhost")
	if err != nil {
		return err
	}

	// database port
	err = os.Setenv(dbPortEnv, "5432")
	if err != nil {
		return err
	}

	// database name
	err = os.Setenv(dbNameEnv, "gab_local")
	if err != nil {
		return err
	}

	// database user
	err = os.Setenv(dbUserEnv, "demo_user")
	if err != nil {
		return err
	}

	// database user password
	err = os.Setenv(dbPasswordEnv, "REPLACE_ME")
	if err != nil {
		return err
	}

	// database search path
	err = os.Setenv(dbSearchPath, "demo")
	if err != nil {
		return err
	}

	// encryption key
	err = os.Setenv(encryptKey, "REPLACE_ME")
	if err != nil {
		return err
	}

	return nil
}
