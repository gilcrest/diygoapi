// appUser_test validates the appUser package methods and objects
package appUser_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gilcrest/go-API-template/appUser"
	"github.com/gilcrest/go-API-template/env"
)

func TestCreate(t *testing.T) {

	// Get an empty context with a Cancel function included
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Cancel ctx as soon as main returns.

	// Initializes "environment" struct type
	env, err := env.NewEnv()
	if err != nil {
		t.Errorf("Error Initializing env, err = %s", err)
	}

	// Creates a new instance of the appUser.User struct type
	inputUsr := appUser.User{Username: "repoMan",
		MobileID:  "(617) 302-7777",
		Email:     "repoman@alwaysintense.com",
		FirstName: "Otto",
		LastName:  "Maddox"}

	// Create method does validation and then inserts user into db
	tx, err := inputUsr.Create(ctx, env)
	if err != nil {
		t.Errorf("Error from Create method, err = %s", err)
	}

	// Check to ensure that the CreateDate struct field is populated by
	// making sure it's not at it's zero value to ensure that the db
	// transaction was successful before commiting
	if !inputUsr.CreateDate.IsZero() {
		err = tx.Commit()
		if err != nil {
			t.Errorf("Error committing tx, err = %s", err)
		}
	} else {
		err = tx.Rollback()
		if err != nil {
			t.Errorf("Error in tx Rollback, err = %s", err)
		}
	}
}

func ExampleUser() {

	usr := appUser.User{Username: "repoMan", MobileID: "(617) 302-7777", Email: "repoman@alwaysintense.com", FirstName: "Otto", LastName: "Maddox"}

	fmt.Println(usr.Username)
	fmt.Println(usr.MobileID)
	fmt.Println(usr.Email)
	fmt.Println(usr.FirstName)
	fmt.Println(usr.LastName)
	// Output:
	// repoMan
	// (617) 302-7777
	// repoman@alwaysintense.com
	// Otto
	// Maddox
}
