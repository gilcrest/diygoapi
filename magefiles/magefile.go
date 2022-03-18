package main

import (
	"github.com/magefile/mage/sh"

	"github.com/gilcrest/go-api-basic/commands"
)

// DBUp uses the psql command line interface to execute DDL scripts
// in the up directory and create all required DB objects. All files
// will be executed, regardless of errors within an individual file.
// Check output to determine if any errors occurred.
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
// in the down directory and drops all project-specific DB objects.
// All files will be executed, regardless of errors within
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

	err = sh.Run("go", "test", "-v", "./...")
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
func Run() (err error) {
	err = commands.OverrideEnv()
	if err != nil {
		return err
	}

	err = sh.Run("go", "build", "-o", "srvr")
	if err != nil {
		return err
	}

	err = sh.Run("./srvr")
	if err != nil {
		return err
	}

	return nil
}

// Genesis runs all tests including executing the Genesis service
func Genesis() (err error) {
	err = commands.OverrideEnv()
	if err != nil {
		return err
	}

	err = commands.Genesis()
	if err != nil {
		return err
	}

	return nil
}

// NewKey generates a new encryption key
func NewKey() {
	commands.NewEncryptionKey()
}
