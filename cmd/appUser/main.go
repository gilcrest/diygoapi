// Allows for testing the different flows without having to go through
// the server
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gilcrest/go-API-template/pkg/appUser"
	"github.com/gilcrest/go-API-template/pkg/appUser/appUserDAF"
	"github.com/gilcrest/go-API-template/pkg/config/db"
	"github.com/gilcrest/go-API-template/pkg/config/env"
)

func main() {
	fmt.Println("Start main")

	// Start Timer
	start := time.Now()

	// Get an empty context with a Cancel function included
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Cancel ctx as soon as main returns.

	// returns an open database handle of 0 or more underlying connections
	// func NewDB() (*sql.DB, error)
	sqldb, err := db.NewDB()

	if err != nil {
		log.Fatal(err)
	}

	env := env.Env{Db: sqldb}

	// db.NewContext function creates and begins a new sql.Tx, which pulls from the
	// previously opened database (postgres) connection pool and starts a database
	// transaction.  In addition, the pointer to this "started" sql.Tx is included
	// in the above created context
	ctx = db.AddDBTx2Context(ctx, env, nil)

	inputUsr := appUser.User{Username: "repoMan", MobileID: "(617) 302-7777", Email: "repoman@alwaysintense.com", FirstName: "Otto", LastName: "Maddox"}
	auditUsr := appUser.User{Username: "bud", MobileID: "(617) 302-7777", Email: "repoman@alwaysintense.com", FirstName: "Otto", LastName: "Maddox"}

	appUserDAF.Create(ctx, inputUsr, auditUsr)

	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())

}
