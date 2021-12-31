//go:build mage
// +build mage

package main

import (
	"github.com/gilcrest/go-api-basic/commands"
	"github.com/magefile/mage/sh"
)

// DBUp uses the psql command line interface to execute DDL scripts
// in the ./scripts/ddl/deploy/up directory and create all required
// DB objects. All files will be executed, regardless of errors within
// an individual file. Check output to determine if any errors occurred.
// Eventually, I will write this to stop on errors, but for now it is
// what it is.
func DBUp() error {
	args, err := commands.PSQLArgs(true)
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
// in the ./scripts/ddl/deploy/down directory and drops all project-specific
// DB objects. All files will be executed, regardless of errors within
// an individual file. Check output to determine if any errors occurred.
// Eventually, I will write this to stop on errors, but for now it is
// what it is.
func DBDown() error {
	args, err := commands.PSQLArgs(false)
	if err != nil {
		return err
	}

	err = sh.Run("psql", args...)
	if err != nil {
		return err
	}
	return nil
}
