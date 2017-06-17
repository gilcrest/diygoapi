// Validates the User object
package appUser_test

import (
	"fmt"

	"github.com/gilcrest/go-API-template/pkg/appUser"
)

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
