//go:build mage
// +build mage

package main

import (
	"github.com/magefile/mage/sh"

	"github.com/gilcrest/go-api-basic/commands"
)

// DBUp uses the psql command line interface to execute DDL scripts
// in the ./scripts/ddl/db-deploy/up directory and create all required
// DB objects. All files will be executed, regardless of errors within
// an individual file. Check output to determine if any errors occurred.
// Eventually, I will write this to stop on errors, but for now it is
// what it is.
func DBUp() error {
	var (
		err  error
		args []string
	)

	err = commands.OverrideEnv()
	if err != nil {
		return err
	}

	args, err = commands.PSQLArgs(true)
	if err != nil {
		return err
	}

	err = sh.Run("psql", args...)
	if err != nil {
		return err
	}

	return nil
}

// DBDown uses the psql command line interface to execute DDL scripts
// in the ./scripts/ddl/db-deploy/down directory and drops all project-specific
// DB objects. All files will be executed, regardless of errors within
// an individual file. Check output to determine if any errors occurred.
// Eventually, I will write this to stop on errors, but for now it is
// what it is.
func DBDown() error {
	var (
		err  error
		args []string
	)

	err = commands.OverrideEnv()
	if err != nil {
		return err
	}

	args, err = commands.PSQLArgs(false)
	if err != nil {
		return err
	}

	err = sh.Run("psql", args...)
	if err != nil {
		return err
	}

	return nil
}

// TestAll runs all tests for the app
func TestAll() error {
	err := commands.OverrideEnv()
	if err != nil {
		return err
	}

	err = sh.Run("go", "test", "./...")
	if err != nil {
		return err
	}

	return nil
}

// Build creates the binary executable with name srvr
func Build() error {
	err := sh.Run("go", "build", "-o", "srvr")
	if err != nil {
		return err
	}

	return nil
}

// Run runs the binary executable created with Build
func Run() error {
	err := commands.OverrideEnv()
	if err != nil {
		return err
	}

	err = sh.Run("./srvr")
	if err != nil {
		return err
	}

	return nil
}
