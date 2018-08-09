package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gilcrest/go-API-template/auth"
	"github.com/gilcrest/go-API-template/env"
)

// LoginHandler is for member login
// TODO better description
func LoginHandler(env *env.Env, w http.ResponseWriter, req *http.Request) error {

	// retrieve the context from the http.Request
	ctx := req.Context()

	// Fire up the db txns (MainDb and Logger DB)
	err := env.DS.SetTx(ctx, nil)
	if err != nil {
		return err
	}

	creds := new(auth.Credentials)

	_ = json.NewDecoder(req.Body).Decode(&creds)

	usr, err := auth.Authorise(ctx, env, creds)
	if err != nil {
		return err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": usr.Username(),
	})

	tokenString, error := token.SignedString([]byte("secret"))
	if error != nil {
		fmt.Println(error)
	}

	json.NewEncoder(w).Encode(auth.JwtToken{Token: tokenString})

	return nil
}
