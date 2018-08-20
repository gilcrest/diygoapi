package auth

import (
	"context"
	"database/sql"

	"github.com/rs/xid"
	"github.com/rs/zerolog"

	"github.com/gilcrest/go-API-template/appuser"
	"github.com/gilcrest/go-API-template/errors"
	"golang.org/x/crypto/bcrypt"
)

type contextKey string

func (c contextKey) String() string {
	return "context key " + string(c)
}

// RequestID is a unique identifier for each inbound request
var requestID = contextKey("RequestID")

// ErrPassNotMatch is an error when a given password hash
// does not match the password hash in the database
var ErrPassNotMatch = errors.Str("Password does not match")

// Credentials stores username/password
type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// JwtToken has the JSON Web Token
type JwtToken struct {
	Token string `json:"token"`
}

// Authorize validates a user/password
//  If valid, the user struct will be populated and error will be nil
//  If invalid, the user struct will be nil and an error will be populated
func Authorize(ctx context.Context, log zerolog.Logger, tx *sql.Tx, c *Credentials) (*appuser.User, error) {
	const op errors.Op = "auth.validatePassword"

	usr, err := appuser.UserFromUsername(ctx, log, tx, c.Username)
	if err != nil {
		return nil, err
	}

	ok := checkPasswordHash(c.Password, usr.Password())

	if !ok {
		return nil, errors.E(op, ErrPassNotMatch)
	}

	return usr, nil
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// SetRequestID adds a unique ID as RequestID to the context
func SetRequestID(ctx context.Context) context.Context {
	// get byte Array representation of guid from xid package (12 bytes)
	guid := xid.New()

	// use the String method of the guid object to convert byte array to string (20 bytes)
	rID := guid.String()

	ctx = context.WithValue(ctx, requestID, rID)

	return ctx

}

// RequestID gets the Request ID from the context.
func RequestID(ctx context.Context) string {
	requestIDstr := ctx.Value(requestID).(string)
	return requestIDstr
}
