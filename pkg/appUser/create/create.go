// Business validations for creating an application user
package create

import (
	"context"

	"github.com/gilcrest/go-API-template/pkg/appUser"
	"github.com/gilcrest/go-API-template/pkg/appUser/appUserDAF"
)

// Perform business validations prior to writing to the db
func Create(ctx context.Context, inputUser *appUser.User, auditUser *appUser.User) (int, error) {

	// Write to db -- function returns rows impacted (should always be 1)
	// or an error
	rows, err := appUserDAF.Create(ctx, inputUser, auditUser)
	return rows, err
}
