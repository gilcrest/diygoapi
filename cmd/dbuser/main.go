package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/gilcrest/diygoapi/cmd"
	"github.com/gilcrest/diygoapi/errs"
)

func main() {
	if err := createDBUser(); err != nil {
		fmt.Fprintf(os.Stderr, "error from createDBUser(): %s\n", err)
		os.Exit(1)
	}
}

// createDBUser connects to PostgreSQL and creates the dga_local user.
func createDBUser() error {
	const op errs.Op = "main/createDBUser"

	args, err := cmd.PSQLConnectionArgs(os.Args)
	if err != nil {
		return errs.E(op, err)
	}

	// Check if the role already exists.
	checkArgs := append(append([]string{}, args...), "-tAc", "SELECT 1 FROM pg_roles WHERE rolname='demo_user'")
	var out bytes.Buffer
	check := exec.Command("psql", checkArgs...)
	check.Stdout = &out
	check.Stderr = os.Stderr
	err = check.Run()
	if err != nil {
		return errs.E(op, err)
	}

	if out.Len() > 0 {
		fmt.Println("database user \"demo_user\" already exists, skipping")
		return nil
	}

	createArgs := append(args, "-c", "CREATE USER demo_user WITH CREATEDB PASSWORD 'REPLACE_ME'")
	c := exec.Command("psql", createArgs...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err = c.Run()
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}
