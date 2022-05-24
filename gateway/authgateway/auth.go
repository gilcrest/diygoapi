// Package authgateway encapsulates outbound calls to authenticate
// a User
package authgateway

import (
	"context"

	"golang.org/x/oauth2"
	googleoauth "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"github.com/gilcrest/diy-go-api/domain/errs"
)

// ProviderUserInfo contains common fields from the various Oauth2 providers.
// Currently only using Google, so looks exactly like Google's.
// TODO - maybe make this look like an OpenID Connect struct
type ProviderUserInfo struct {
	// Username: For most providers, the username is the email.
	Username string

	// Email: The user's email address.
	Email string `json:"email,omitempty"`

	// FamilyName: The user's last name.
	FamilyName string `json:"family_name,omitempty"`

	// Gender: The user's gender.
	Gender string `json:"gender,omitempty"`

	// GivenName: The user's first name.
	GivenName string `json:"given_name,omitempty"`

	// Hd: The hosted domain e.g. example.com if the user is Google apps
	// user.
	Hd string `json:"hd,omitempty"`

	// Id: The obfuscated ID of the user.
	Id string `json:"id,omitempty"`

	// Link: URL of the profile page.
	Link string `json:"link,omitempty"`

	// Locale: The user's preferred locale.
	Locale string `json:"locale,omitempty"`

	// Name: The user's full name.
	Name string `json:"name,omitempty"`

	// Picture: URL of the user's picture image.
	Picture string `json:"picture,omitempty"`
}

// GoogleOauth2TokenConverter is used to convert an oauth2.Token to a User
// through Google's API
type GoogleOauth2TokenConverter struct{}

// Convert calls the Google Userinfo API with the access token and converts
// the Userinfo struct to a User struct
func (c GoogleOauth2TokenConverter) Convert(ctx context.Context, realm string, token oauth2.Token) (ProviderUserInfo, error) {
	oauthService, err := googleoauth.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(&token)))
	if err != nil {
		return ProviderUserInfo{}, errs.E(err)
	}

	uInfo, err := oauthService.Userinfo.Get().Do()
	if err != nil {
		// "In summary, a 401 Unauthorized response should be used for missing or
		// bad authentication, and a 403 Forbidden response should be used afterwards,
		// when the user is authenticated but isn’t authorized to perform the
		// requested operation on the given resource."
		// In this case, we are getting a bad response from Google service, assume
		// they are not able to authenticate properly
		return ProviderUserInfo{}, errs.E(errs.Unauthenticated, errs.Realm(realm), err)
	}

	pui := ProviderUserInfo{
		Username:   uInfo.Email,
		Email:      uInfo.Email,
		FamilyName: uInfo.FamilyName,
		Gender:     uInfo.Gender,
		GivenName:  uInfo.GivenName,
		Hd:         uInfo.Hd,
		Id:         uInfo.Id,
		Link:       uInfo.Link,
		Locale:     uInfo.Locale,
		Name:       uInfo.Name,
		Picture:    uInfo.Picture,
	}

	return pui, nil
}
