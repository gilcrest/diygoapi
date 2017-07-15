package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gilcrest/go-API-template/pkg/appUser"
	"github.com/gilcrest/go-API-template/pkg/config/db"
	"github.com/gilcrest/go-API-template/pkg/config/env"
)

type testUser struct {
	Username  string
	MobileID  string
	Email     string
	FirstName string
	LastName  string
}

type UserHandler struct {
	Env *env.Env
}

/*
Creates a user in the database, but also:
	- writes a log of the request and response
	- "pretty prints" the request
*/
func (uh *UserHandler) CreateUserHandler(w http.ResponseWriter, req *http.Request) {
	// retrieve the context from the http.Request
	ctx := req.Context()
	logger := uh.Env.Logger
	logger.Debug("handleMbrLog started")

	defer uh.Env.Logger.Sync()
	defer logger.Debug("handleMbrLog ended")

	//logRequest(req)
	//prettyPrintRequest(req)

	// db.NewContext function creates and begins a new sql.Tx, which pulls from the
	// previously opened database (postgres) connection pool and starts a database
	// transaction.  In addition, the pointer to this "started" sql.Tx is included
	// in the above created context
	ctx = db.AddDBTx2Context(ctx, uh.Env, nil)

	decoder := json.NewDecoder(req.Body)
	var t testUser
	err := decoder.Decode(&t)
	if err != nil {
		panic(err)
	}
	defer req.Body.Close()
	log.Println(t)

	inputUsr := appUser.User{Username: "repoMan", MobileID: "(617) 302-7777", Email: "repoman@alwaysintense.com", FirstName: "Otto", LastName: "Maddox"}
	//auditUsr := appUser.User{Username: "bud", MobileID: "(617) 302-7777", Email: "repoman@alwaysintense.com", FirstName: "Otto", LastName: "Maddox"}

	// Call the create method of the appUser object to validate data and write to db
	logsWritten, err := inputUsr.Create(ctx)

	fmt.Fprintf(w, "logsWritten = %d\n", logsWritten)

	tx, ok := db.DBTxFromContext(ctx)

	if ok && logsWritten > 0 {
		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}
	} else if logsWritten <= 0 {
		log.Fatal(err)
	}
}
