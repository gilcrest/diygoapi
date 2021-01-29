package authgateway

import (
	"context"

	"github.com/gilcrest/go-api-basic/domain/auth"

	"github.com/gilcrest/go-api-basic/domain/user"

	"golang.org/x/oauth2"
	googleoauth "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"github.com/gilcrest/go-api-basic/domain/errs"
)

// GoogleToken2User is used to convert an auth.AccessToken to a User
// through Google's API
type GoogleToken2User struct{}

// User calls the Google Userinfo API with the access token and converts
// the Userinfo struct to a User struct
func (c GoogleToken2User) User(ctx context.Context, token auth.AccessToken) (*user.User, error) {
	ui, err := userInfo(ctx, token.NewGoogleOauth2Token())
	if err != nil {
		return nil, err
	}

	return newUser(ui), nil
}

// userInfo makes an outbound https call to Google using their
// Oauth2 v2 api and returns a Userinfo struct which has most
// profile data elements you typically need
func userInfo(ctx context.Context, token *oauth2.Token) (*googleoauth.Userinfo, error) {

	oauthService, err := googleoauth.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(token)))
	if err != nil {
		return nil, errs.E(err)
	}

	userInfo, err := oauthService.Userinfo.Get().Do()
	if err != nil {
		// "In summary, a 401 Unauthorized response should be used for missing or
		// bad authentication, and a 403 Forbidden response should be used afterwards,
		// when the user is authenticated but isn’t authorized to perform the
		// requested operation on the given resource."
		// In this case, we are getting a bad response from Google service, assume
		// they are not able to authenticate properly
		return nil, errs.E(errs.Unauthenticated, err)
	}

	return userInfo, nil
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
