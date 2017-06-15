// Data Access Functions (database reads, writes, deletes) for User data
package userDAF

import (
	"context"
	"errors"

	"github.com/gilcrest/go-API-template/pkg/config/db"
	"github.com/gilcrest/go-API-template/pkg/user"
)

// Creates a record in the user table using a stored function which
// returns the number of rows inserted
func Create(ctx context.Context, inputUser user.User, auditUser user.User) (int, error) {

	var (
		rowsInserted int
	)

	// pull pointer to sql.Tx as tx from context passed in parameter
	tx, ok := db.DBTxFromContext(ctx)

	// ensure there is a sql.Tx in the context by checking boolean passed back above
	if !ok {
		return -1, errors.New("Unable to retrieve sql.Tx from DBTxFromContext function")
	}

	// Prepare the sql statement using bind variables
	stmt, err := tx.PrepareContext(ctx, "select lp.create_app_user(p_username => $1, p_mobile_id => $2, p_email_address => $3, p_first_name => $4, p_last_name => $5, p_create_user_id => $6)")
	if err != nil {
		return -1, err
	}
	defer stmt.Close()

	// Execute stored function that returns rows impacted, hence the use of QueryContext instead of Exec
	rows, err := stmt.QueryContext(ctx, inputUser.Username, inputUser.MobileID, inputUser.Email, inputUser.FirstName, inputUser.LastName, auditUser.Username)
	if err != nil {
		return -1, err
	}
	defer rows.Close()

	// Iterate through
	for rows.Next() {
		if err := rows.Scan(&rowsInserted); err != nil {
			return -1, err
		}
	}

	if err := rows.Err(); err != nil {
		return -1, err
	}

	return rowsInserted, nil

}
