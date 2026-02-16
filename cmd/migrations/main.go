package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/gilcrest/diygoapi/cmd"
	"github.com/gilcrest/diygoapi/errs"
)

func main() {
	if err := DBUp(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error from DBUp(): %s\n", err)
		os.Exit(1)
	}
}

// DBUp executes DDL scripts which create all required DB objects,
// example: mage -v dbup local.
//
// All files will be executed, regardless of errors within an individual
// file. Check output to determine if any errors occurred. Eventually,
// I will write this to stop on errors, but for now it is what it is.
func DBUp(args []string) (err error) {
	const op errs.Op = "main/DBUp"

	args, err = cmd.PSQLArgs(true, os.Args)
	if err != nil {
		return errs.E(op, err)
	}

	c := exec.Command("psql", args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err = c.Run()
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}

//// DBDown executes DDL scripts which drops all project-specific DB objects,
//// example: mage -v dbdown local.
////
//// All files will be executed, regardless of errors within an individual
//// file. Check output to determine if any errors occurred. Eventually,
//// I will write this to stop on errors, but for now it is what it is.
//func DBDown(env string) (err error) {
//	const op errs.Op = "main/DBDown"
//
//	var args []string
//
//	err = cmd.LoadEnv(cmd.ParseEnv(env))
//	if err != nil {
//		return errs.E(op, err)
//	}
//
//	args, err = cmd.PSQLArgs(false)
//	if err != nil {
//		return errs.E(op, err)
//	}
//
//	err = sh.Run("psql", args...)
//	if err != nil {
//		return errs.E(op, err)
//	}
//
//	return nil
//}
