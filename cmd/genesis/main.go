package main

import (
	"fmt"
	"os"

	"github.com/gilcrest/diygoapi/cmd"
)

func main() {
	if err := cmd.Genesis(); err != nil {
		fmt.Fprintf(os.Stderr, "error from cmd.Genesis(): %s\n", err)
		os.Exit(1)
	}
}
