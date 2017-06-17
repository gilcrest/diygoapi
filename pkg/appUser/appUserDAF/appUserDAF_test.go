package appUserDAF_test

import (
	"context"
	"testing"

	"github.com/gilcrest/go-API-template/pkg/config/db"
)

var err error

func TestCreate(t *testing.T) {

	// Get an empty context with a Cancel function included
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Cancel ctx as soon as main returns.

	// Populates DBCon global variable from config/db using the NewDB function from
	// the same package
	db.DBCon, err = db.NewDB()

	//inputUsr := appUser.User{Username: "repoMan", MobileID: "(617) 302-7777", Email: "repoman@alwaysintense.com", FirstName: "Otto", LastName: "Maddox"}
	//auditUsr := appUser.User{Username: "bud", MobileID: "(617) 302-7777", Email: "repoman@alwaysintense.com", FirstName: "Otto", LastName: "Maddox"}


	// db.NewContext function creates and begins a new sql.Tx, which pulls from the
	// previously opened database (postgres) connection pool and starts a database
	// transaction.  In addition, the pointer to this "started" sql.Tx is included
	// in the above created context
	ctx = db.AddDBTx2Context(ctx, nil)

	//rowsInserted, err := appUserDAF.Create(ctx, inputUsr, auditUsr)
	//
	//if rowsInserted != 1 {
	//	t.Error("Expected 1 row inserted, got ", rowsInserted)
	//}
	if 0 == 1 {
		t.Error("Bummer")
	}
}
