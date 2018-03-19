// Package appUser has the User type and associated methods to
// create, modify and delete application users
package appUser

import (
	"context"
	"database/sql"
	"net/mail"
	"time"

	"github.com/gilcrest/go-API-template/env"
)

// User represents an application user.  A user can access multiple systems.
// The User-Application relationship is kept elsewhere...
type User struct {
	Username     string    `json:"username"`
	MobileID     string    `json:"mobile_id"`
	Email        string    `json:"email"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	CreateUserID string    `json:"create_user_id"`
	CreateDate   time.Time `json:"create_date"`
	UpdateUserID string    `json:"update_user_id"`
	UpdateDate   time.Time `json:"update_date"`
}

// Create performs business validations prior to writing to the db
func (u *User) Create(ctx context.Context, env *env.Env) (*sql.Tx, error) {

	// Get a new logger instance
	log := env.Logger

	log.Debug().Msg("Start User.Create")
	defer log.Debug().Msg("Finish User.Create")

	// You should add more validations than this, but since this is a template, you
	// get the point
	_, err := mail.ParseAddress(u.Email)
	if err != nil {
		return nil, err
	}

	// Ideally this would be set from the user id adding the resource,
	// but since I haven't implemented that yet, using this hack
	u.CreateUserID = "chillcrest"

	// Write to db
	tx, err := u.createDB(ctx, env)

	return tx, err
}

// Creates a record in the appUser table using a stored function
func (u *User) createDB(ctx context.Context, env *env.Env) (*sql.Tx, error) {

	var (
		createDate time.Time
	)

	// Calls the BeginTx method of the MainDb opened database
	tx, err := env.DS.MainDb.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	// Prepare the sql statement using bind variables
	stmt, err := tx.PrepareContext(ctx, `select demo.create_app_user (
		p_username => $1,
		p_mobile_id => $2,
		p_email_address => $3,
		p_first_name => $4,
		p_last_name => $5,
		p_create_user_id => $6)`)

	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	// Execute stored function that returns the create_date timestamp,
	// hence the use of QueryContext instead of Exec
	rows, err := stmt.QueryContext(ctx,
		u.Username,  //$1
		u.MobileID,  //$2
		u.Email,     //$3
		u.FirstName, //$4
		u.LastName,  //$5
		"gilcrest")  //$6

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Iterate through the returned record(s)
	for rows.Next() {
		if err := rows.Scan(&createDate); err != nil {
			return nil, err
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// set the CreateDate field to the create_date set as part of the insert in
	// the stored function call above
	u.CreateDate = createDate

	return tx, nil

}
