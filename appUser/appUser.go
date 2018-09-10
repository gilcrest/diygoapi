// Package appuser has the User type and associated methods to
// create, modify and delete application users
package appuser

import (
	"context"
	"database/sql"
	"net/mail"
	"time"

	"github.com/gilcrest/go-API-template/errors"
	"github.com/gilcrest/httplog"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

// ErrNoUser is an error when a user is passed
// that does not exist in the db
var ErrNoUser = errors.Str("User does not exist")

// User represents an application user.  A user can access multiple systems.
// The User-Application relationship is kept elsewhere...
type User struct {
	username        string
	password        string
	mobileID        string
	email           string
	firstName       string
	lastName        string
	updateClientID  string
	updateUserID    string
	updateTimestamp time.Time
}

// CreateUserRequest is the expected service request fields
type CreateUserRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	MobileID     string `json:"mobile_id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	UpdateUserID string `json:"udpate_user_id"`
}

// CreateUserResponse is the expected service response fields
type CreateUserResponse struct {
	Username       string         `json:"username"`
	MobileID       string         `json:"mobile_id"`
	Email          string         `json:"email"`
	FirstName      string         `json:"first_name"`
	LastName       string         `json:"last_name"`
	UpdateUserID   string         `json:"update_user_id"`
	UpdateUnixTime int64          `json:"created"`
	Audit          *httplog.Audit `json:"audit"`
}

// NewUser constructor performs basic service validations
//  and wires request data into User business object
func NewUser(ctx context.Context, cur *CreateUserRequest) (*User, error) {

	// declare a new instance of appuser.User
	usr := new(User)

	err := usr.SetUsername(cur.Username)
	if err != nil {
		return nil, err
	}
	err = usr.setPassword(ctx, cur.Password)
	if err != nil {
		return nil, err
	}
	err = usr.SetMobileID(cur.MobileID)
	if err != nil {
		return nil, err
	}
	err = usr.SetEmail(cur.Email)
	if err != nil {
		return nil, err
	}
	err = usr.SetFirstName(cur.FirstName)
	if err != nil {
		return nil, err
	}
	err = usr.SetLastName(cur.LastName)
	if err != nil {
		return nil, err
	}
	err = usr.SetUpdateClientID("client a")
	if err != nil {
		return nil, err
	}
	err = usr.SetUpdateUserID(cur.UpdateUserID)
	if err != nil {
		return nil, err
	}

	return usr, nil

}

// Username is a getter for User.username
func (u *User) Username() string {
	return u.username
}

// SetUsername is a setter for User.username
func (u *User) SetUsername(username string) error {
	// for each field you can go through whatever validations you wish
	// and use the SetErr method of the HTTPErr struct to add the proper
	// error text
	switch {
	// Username is required
	case username == "":
		return errors.Str("Username is a required field")
	// Username cannot be blah...
	case username == "blah":
		return errors.Str("Username cannot be blah")
	default:
		u.username = username
	}
	u.username = username
	return nil
}

// Password is a getter for User.username
func (u *User) Password() string {
	return u.password
}

func (u *User) setPassword(ctx context.Context, password string) error {

	if password == "" {
		return errors.Str("Password is mandatory")
	}

	// Salt and hash the password using the bcrypt algorithm
	passHash, err := bcrypt.GenerateFromPassword([]byte(password), 8)
	if err != nil {
		return err
	}

	u.password = string(passHash)

	return nil
}

// MobileID is a getter for User.mobileID
func (u *User) MobileID() string {
	return u.mobileID
}

// SetMobileID is a setter for User.username
func (u *User) SetMobileID(mobileID string) error {
	u.mobileID = mobileID
	return nil
}

// Email is a getter for User.mail
func (u *User) Email() string {
	return u.email
}

// SetEmail is a setter for User.email
func (u *User) SetEmail(email string) error {
	_, err := mail.ParseAddress(email)
	if err != nil {
		return err
	}
	u.email = email
	return nil
}

// FirstName is a getter for User.firstName
func (u *User) FirstName() string {
	return u.firstName
}

// SetFirstName is a setter for User.firstName
func (u *User) SetFirstName(firstName string) error {
	u.firstName = firstName
	return nil
}

// LastName is a getter for User.lastName
func (u *User) LastName() string {
	return u.lastName
}

// SetLastName is a setter for User.lastName
func (u *User) SetLastName(lastName string) error {
	u.lastName = lastName
	return nil
}

// UpdateClientID is a getter for User.updateClientID
func (u *User) UpdateClientID() string {
	return u.updateClientID
}

// SetUpdateClientID is a setter for User.updateClientID
func (u *User) SetUpdateClientID(clientID string) error {
	u.updateClientID = clientID
	return nil
}

// UpdateUserID is a getter for User.updateUserID
func (u *User) UpdateUserID() string {
	return u.updateUserID
}

// SetUpdateUserID is a setter for User.UpdateUserID
func (u *User) SetUpdateUserID(userID string) error {
	u.updateUserID = userID
	return nil
}

// UpdateTimestamp is a getter for User.updateDate
func (u *User) UpdateTimestamp() time.Time {
	return u.updateTimestamp
}

// SetUpdateTimestamp is a setter for User.updateTimestamp
func (u *User) SetUpdateTimestamp(updateTimestamp time.Time) error {
	u.updateTimestamp = updateTimestamp
	return nil
}

// Create performs business validations prior to writing to the db
func (u *User) Create(ctx context.Context, log zerolog.Logger) error {

	log.Debug().Msg("Start User.Create")
	defer log.Debug().Msg("Finish User.Create")

	// Ideally this would be set from the user id adding the resource,
	// but since I haven't implemented that yet, using this hack
	u.updateUserID = "chillcrest"

	return nil
}

// CreateDB creates a record in the appuser table using a stored function
func (u *User) CreateDB(ctx context.Context, log zerolog.Logger, tx *sql.Tx) error {

	var (
		updateTimestamp time.Time
	)

	// Prepare the sql statement using bind variables
	stmt, err := tx.PrepareContext(ctx, `select demo.create_app_user (
		p_username => $1,
		p_password => $2,
		p_mobile_id => $3,
		p_email_address => $4,
		p_first_name => $5,
		p_last_name => $6,
		p_user_id => $7)`)

	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute stored function that returns the create_date timestamp,
	// hence the use of QueryContext instead of Exec
	rows, err := stmt.QueryContext(ctx,
		u.username,  //$1
		u.password,  //$2
		u.mobileID,  //$3
		u.email,     //$4
		u.firstName, //$5
		u.lastName,  //$6
		u.username)  //$7

	if err != nil {
		return err
	}
	defer rows.Close()

	// Iterate through the returned record(s)
	for rows.Next() {
		if err := rows.Scan(&updateTimestamp); err != nil {
			return err
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	// set the CreateDate field to the create_date set as part of the insert in
	// the stored function call above
	u.updateTimestamp = updateTimestamp

	return nil

}

// UserFromUsername constructs a User given a username
func UserFromUsername(ctx context.Context, log zerolog.Logger, tx *sql.Tx, username string) (*User, error) {
	const op errors.Op = "appuser.UserFromUsername"

	// Prepare the sql statement using bind variables
	row := tx.QueryRowContext(ctx,
		`select username,
				password,
				mobile_id,
				email_address,
				first_name,
				last_name,
				update_client_id,
				update_user_id,
				update_timestamp
  		   from demo.user
          where username = $1`, username)

	usr := new(User)
	err := row.Scan(&usr.username,
		&usr.password,
		&usr.mobileID,
		&usr.email,
		&usr.firstName,
		&usr.lastName,
		&usr.updateClientID,
		&usr.updateUserID,
		&usr.updateTimestamp,
	)
	if err == sql.ErrNoRows {
		return nil, errors.E(op, ErrNoUser)
	} else if err != nil {
		return nil, errors.E(op, err)
	}

	return usr, nil
}
