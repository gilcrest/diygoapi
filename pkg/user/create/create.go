// Business validations for creating a user
package create

import (
	"context"

	"github.com/gilcrest/go-API-template/pkg/user"
	"github.com/gilcrest/go-API-template/pkg/user/userDAF"
)

// Perform business validations prior to writing to the db
func createUser(ctx context.Context, inputUser user.User, auditUser user.User) (int, error) {

	// Write to db -- function returns rows impacted (should always be 1)
	// or an error
	rows, err := userDAF.Create(ctx, inputUser, auditUser)
	return rows, err
}
