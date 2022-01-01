package main

import (
	"fmt"
	"os"

	"github.com/gilcrest/go-api-basic/commands"
)

func main() {
	if err := commands.Run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error from commands.Run(): %s\n", err)
		os.Exit(1)
	}
}
