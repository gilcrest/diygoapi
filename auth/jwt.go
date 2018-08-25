package auth

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gilcrest/go-API-template/appuser"
	"github.com/gilcrest/go-API-template/errors"
)

const dayinsecs int64 = 86400

// LoginClaims struct has the jwt claims that will be added to the
// login response token
type LoginClaims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// ServerClaims struct has the jwt claims that will be added to the
// token used for non-Bearer APIs
type ServerClaims struct {
	jwt.StandardClaims
}

// JwtToken has the JSON Web Token
type JwtToken struct {
	Token string `json:"token"`
}

// LoginToken takes a user and provides a token for the user
func LoginToken(usr *appuser.User) (JwtToken, error) {
	const op errors.Op = "auth.LoginToken"

	var (
		authToken JwtToken
	)

	// Create the Claims
	claims := LoginClaims{
		usr.Username(),
		jwt.StandardClaims{
			ExpiresAt: loginTokenExpire(dayinsecs),
			Issuer:    "https://github.com/gilcrest",
			IssuedAt:  loginTokenIssuedAt(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(jwtSigningKey())
	if err != nil {
		return authToken, errors.E(op, err)
	}

	fmt.Printf("%v %v", ss, err)

	authToken.Token = ss

	return authToken, nil

}

// ServerToken generates a server token
func ServerToken() (string, error) {
	const op errors.Op = "auth.ServerToken"

	// Create the Claims
	claims := ServerClaims{
		jwt.StandardClaims{
			Issuer:   "https://github.com/gilcrest",
			IssuedAt: loginTokenIssuedAt(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString(jwtSigningKey())
	if err != nil {
		return "", errors.E(op, err)
	}

	return ss, nil

}

func jwtSigningKey() []byte {
	return []byte("ssitariismixedupandblind")
}

func loginTokenExpire(seconds2expire int64) int64 {
	return time.Now().Add(time.Second * time.Duration(seconds2expire)).Unix()
}

func loginTokenIssuedAt() int64 {
	return time.Now().Unix()
}

func validateJWT(jwTolkien string) bool {

	token, err := jwt.Parse(jwTolkien, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return jwtSigningKey(), nil
	})

	fmt.Printf("token.Valid = %v\n", token.Valid)

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		fmt.Printf("%+v", claims)
		//fmt.Printf("%v %v\n", claims.Username, claims.StandardClaims.Issuer)
		return true
	}

	fmt.Println(err)
	return false
}
