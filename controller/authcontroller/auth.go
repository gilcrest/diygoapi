package authcontroller

import (
	"context"

	"golang.org/x/oauth2"
	googleoauth "google.golang.org/api/oauth2/v2"

	"github.com/gilcrest/errs"
	"github.com/gilcrest/go-api-basic/domain/auth"
	"github.com/gilcrest/go-api-basic/domain/user"
	"github.com/gilcrest/go-api-basic/gateway/authgateway"
)

// AuthorizeAccessToken takes an access token string, validates
// that the user exists by calling out to Google's Userinfo API and
// then authorizes the user
func AuthorizeAccessToken(ctx context.Context, token string) (*user.User, error) {
	const op errs.Op = "controller/authcontroller/AuthorizeTokenController"

	// Setup oauth token
	oauthToken := oauth2.Token{AccessToken: token, TokenType: "Bearer"}
	// use Google Oauth2 API to get user info
	userInfo, err := authgateway.UserInfo(ctx, &oauthToken)
	if err != nil {
		// "In summary, a 401 Unauthorized response should be used for missing or
		// bad authentication, and a 403 Forbidden response should be used afterwards,
		// when the user is authenticated but isn’t authorized to perform the
		// requested operation on the given resource."
		// In this case, we are getting a bad response from Google service, assume
		// they are not able to authenticate properly
		return nil, errs.E(op, errs.Unauthenticated, err)
	}

	// Set userInfo from google into domain user
	u := newUser(userInfo)

	// validate that user is authorized
	err = auth.AuthorizeUser(ctx, u)
	if err != nil {
		return nil, errs.E(op, err)
	}

	return u, nil
}

// newUser initializes the user.User struct given a Userinfo struct
// from Google
func newUser(userinfo *googleoauth.Userinfo) *user.User {
	return &user.User{
		Email:     userinfo.Email,
		LastName:  userinfo.FamilyName,
		FirstName: userinfo.GivenName,
		FullName:  userinfo.Name,
		//Gender:       userinfo.Gender,
		HostedDomain: userinfo.Hd,
		PictureURL:   userinfo.Picture,
		ProfileLink:  userinfo.Link,
	}
}
