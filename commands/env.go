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
// this is purely demonstrative and the password value is invalid.
func OverrideEnv() error {
	var err error

	err = os.Setenv("DB_NAME", "gab_local")
	if err != nil {
		return err
	}
	err = os.Setenv("DB_USER", "demo_user")
	if err != nil {
		return err
	}
	err = os.Setenv("DB_PASSWORD", "REPLACE_ME")
	if err != nil {
		return err
	}
	err = os.Setenv("DB_HOST", "localhost")
	if err != nil {
		return err
	}
	err = os.Setenv("DB_PORT", "5432")
	if err != nil {
		return err
	}
	err = os.Setenv("DB_SEARCH_PATH", "demo")
	if err != nil {
		return err
	}

	return nil
}
