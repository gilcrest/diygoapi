package main

import (
	"fmt"
	"github.com/gilcrest/diy-go-api/cmd"
	"os"
)

func main() {
	if err := cmd.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error from commands.Run(): %s\n", err)
		os.Exit(1)
	}
}
