// Allows for testing the different flows without having to go through
// the http server
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gilcrest/go-API-template/pkg/domain/appUser"
	"github.com/gilcrest/go-API-template/pkg/env"
)

func main() {
	fmt.Println("Start main")

	// Start Timer
	start := time.Now()

	// Get an empty context with a Cancel function included
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Cancel ctx as soon as main returns.

	env, err := env.NewEnv()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(env)
	inputUsr := appUser.User{Username: "repoMan", MobileID: "(617) 302-7777", Email: "repoman@alwaysintense.com", FirstName: "Otto", LastName: "Maddox"}
	// //auditUsr := appUser.User{Username: "bud", MobileID: "(617) 302-7777", Email: "repoman@alwaysintense.com", FirstName: "Otto", LastName: "Maddox"}

	// //
	logsWritten, tx, err := inputUsr.Create(ctx, env)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("logsWritten = %d\n", logsWritten)

	if logsWritten > 0 {
		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}
	} else if logsWritten <= 0 {
		log.Fatal(err)
	}

	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())

}
