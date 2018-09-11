// appuser_test validates the appuser package methods and objects
package appuser_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/gilcrest/go-API-template/appuser"
	"github.com/gilcrest/go-API-template/db"
	"github.com/gilcrest/go-API-template/env"
	"github.com/rs/zerolog"
)

func TestCreate(t *testing.T) {

	// Get an empty context with a Cancel function included
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Cancel ctx as soon as main returns.

	// Initializes "environment" struct type
	env, err := env.NewEnv(zerolog.DebugLevel)
	if err != nil {
		t.Errorf("Error Initializing env, err = %s", err)
	}

	cur := new(appuser.CreateUserRequest)
	cur.Username = "gilcrest"
	cur.Password = "fakepassword"
	cur.MobileID = "976"
	cur.LastName = "Gilcrest"
	cur.FirstName = "Dan"
	cur.Email = "testcrest@gmail.com"
	cur.UpdateUserID = "gilcrest"

	// Creates a new instance of the appuser.User struct type
	inputUsr, err := appuser.NewUser(ctx, cur)
	if err != nil {
		t.Errorf("Error committing tx, err = %s", err)
	}

	log := env.Logger

	tx, err := env.DS.BeginTx(ctx, nil, db.AppDB)
	if err != nil {
		t.Errorf("Error from BeginTx method, err = %s", err)
	}

	// Create method does validation and then inserts user into db
	err = inputUsr.Create(ctx, log)
	if err != nil {
		t.Errorf("Error from Create method, err = %s", err)
	}

	// Call the create method of the User object to write
	// to the database
	err = inputUsr.CreateDB(ctx, log, tx)
	if err != nil {
		t.Errorf("Error from CreateDB method, err = %s", err)
	}

	// Check to ensure that the CreateDate struct field is populated by
	// making sure it's not at it's zero value to ensure that the db
	// transaction was successful before commiting
	if !inputUsr.UpdateTimestamp().IsZero() {
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

	usr := new(appuser.User)
	usr.SetUsername("repoMan")
	usr.SetMobileID("(617) 302-7777")
	usr.SetEmail("repoman@alwaysintense.com")
	usr.SetFirstName("Otto")
	usr.SetLastName("Maddox")

	fmt.Println(usr.Username())
	fmt.Println(usr.MobileID())
	fmt.Println(usr.Email())
	fmt.Println(usr.FirstName())
	fmt.Println(usr.LastName())

	// Output:
	// repoMan
	// (617) 302-7777
	// repoman@alwaysintense.com
	// Otto
	// Maddox
}

// func TestUserFromUsername(t *testing.T) {
// 	// Get an empty context with a Cancel function included
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel() // Cancel ctx as soon as test returns.

// 	// Initializes "environment" struct type
// 	ev, err := env.NewEnv(zerolog.DebugLevel)
// 	if err != nil {
// 		t.Errorf("Error Initializing env, err = %s", err)
// 	}

// 	// set the *sql.Tx for the
// 	// datastore within the environment
// 	_, err = ev.DS.BeginTx(ctx, nil, db.AppDB)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	usr := new(appuser.User)
// 	usr.SetUsername("fuckface")

// 	type args struct {
// 		ctx      context.Context
// 		log      zerolog.Logger
// 		tx       *sql.Tx
// 		username string
// 	}

// 	tests := []struct {
// 		name    string
// 		args    args
// 		want    *appuser.User
// 		wantErr bool
// 	}{
// 		{"Test Dan", args{ctx: ctx, env: ev, username: "asdf"}, usr, false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			got, err := appuser.UserFromUsername(tt.args.ctx, tt.args.env, tt.args.username)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("UserFromUsername() error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			if !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("UserFromUsername() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
