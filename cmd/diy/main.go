package main

import (
	"fmt"
	"os"

	"github.com/gilcrest/diygoapi/cmd"
)

func main() {
	if err := cmd.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error from commands.Run(): %s\n", err)
		os.Exit(1)
	}
}
