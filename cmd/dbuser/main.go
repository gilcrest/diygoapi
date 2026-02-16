package main

import (
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

	args = append(args, "-c", "CREATE USER demo_user WITH CREATEDB PASSWORD 'REPLACE_ME'")

	c := exec.Command("psql", args...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	err = c.Run()
	if err != nil {
		return errs.E(op, err)
	}

	return nil
}
