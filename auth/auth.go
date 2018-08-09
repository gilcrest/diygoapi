package auth

import (
	"context"

	"github.com/gilcrest/go-API-template/appuser"
	"github.com/gilcrest/go-API-template/errors"
	"golang.org/x/crypto/bcrypt"

	"github.com/gilcrest/go-API-template/env"
)

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

// Authorise validates a user/password
//  If valid, the user struct will be populated and error will be nil
//  If invalid, the user struct will be nil and an error will be populated
func Authorise(ctx context.Context, env *env.Env, c *Credentials) (*appuser.User, error) {
	const op errors.Op = "auth.validatePassword"

	usr, err := appuser.UserFromUsername(ctx, env, c.Username)
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
