// Package gateway and packages within provide abstractions for
// interacting with external systems or resources
package gateway

import (
	"context"
	"time"

	"golang.org/x/oauth2"
	googleoauth "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"github.com/gilcrest/diygoapi"
	"github.com/gilcrest/diygoapi/errs"
)

// Oauth2TokenExchange is used to convert an oauth2.Token to a ProviderInfo
// struct from details returned from a provider API
type Oauth2TokenExchange struct{}

// Exchange calls the Google Userinfo API with the access token and converts
// the Userinfo struct to a User struct
func (e Oauth2TokenExchange) Exchange(ctx context.Context, realm string, provider diygoapi.Provider, token *oauth2.Token) (*diygoapi.ProviderInfo, error) {
	const op errs.Op = "gateway/Oauth2TokenExchange.Exchange"

	switch provider {
	case diygoapi.Google:
		return googleTokenExchange(ctx, realm, token)
	default:
		return nil, errs.E(op, errs.Unauthenticated, errs.Realm(realm), "provider not recognized")
	}
}

// googleTokenExchange makes a request to Google's OAuth2 API and
// populates ProviderInfo based on the response.
func googleTokenExchange(ctx context.Context, realm string, token *oauth2.Token) (*diygoapi.ProviderInfo, error) {
	const op errs.Op = "gateway/googleTokenExchange"

	var (
		oauthService *googleoauth.Service
		err          error
	)
	// initialize the Oauth2 service
	oauthService, err = googleoauth.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(token)))
	if err != nil {
		return nil, errs.E(op, errs.Internal, err)
	}

	// make api call to get metadata about the token
	var tokenInfo *googleoauth.Tokeninfo
	tokenInfo, err = oauthService.Tokeninfo().Do()
	if err != nil {
		return nil, errs.E(op, errs.Unauthenticated, errs.Realm(realm), err)
	}

	pti := diygoapi.ProviderTokenInfo{
		Token:    token,
		ClientID: tokenInfo.IssuedTo,
		Scope:    tokenInfo.Scope,
		Audience: tokenInfo.Audience,
		IssuedTo: tokenInfo.IssuedTo,
	}

	// calculate the token expiration based on ExpiresIn (seconds)
	pti.Token.Expiry = time.Now().Add(time.Duration(tokenInfo.ExpiresIn) * time.Second)

	var userinfo *googleoauth.Userinfo
	userinfo, err = oauthService.Userinfo.Get().Do()
	if err != nil {
		// "In summary, a 401 Unauthorized response should be used for missing or
		// bad authentication, and a 403 Forbidden response should be used afterwards,
		// when the user is authenticated but isn’t authorized to perform the
		// requested operation on the given resource."
		// In this case, we are getting a bad response from Google service, assume
		// they are not able to authenticate properly
		return nil, errs.E(op, errs.Unauthenticated, errs.Realm(realm), err)
	}

	// according to Google's docs, IssuedTo and Audience are mostly the same,
	// adding this check to figure out when they're not. I need to figure out
	// which is the most reliable, hopefully this error never occurs, and they're
	// in actuality always the same.
	if tokenInfo.IssuedTo != tokenInfo.Audience {
		return nil, errs.E(op, errs.Internal, "tokenInfo.IssuedTo != tokenInfo.Audience")
	}

	// Again, I believe from docs that these are the same, but adding this
	// validation as I need a reliable external key.
	if tokenInfo.UserId != userinfo.Id {
		return nil, errs.E(op, errs.Internal, "tokenInfo.UserId != tokenInfo.Id")
	}

	pui := diygoapi.ProviderUserInfo{
		ExternalID:    userinfo.Id,
		Email:         userinfo.Email,
		VerifiedEmail: tokenInfo.VerifiedEmail,
		FirstName:     userinfo.GivenName,
		LastName:      userinfo.FamilyName,
		FullName:      userinfo.Name,
		Gender:        userinfo.Gender,
		HostedDomain:  userinfo.Hd,
		ProfileLink:   userinfo.Link,
		Locale:        userinfo.Locale,
		Picture:       userinfo.Picture,
	}

	pi := diygoapi.ProviderInfo{
		Provider:  diygoapi.Google,
		TokenInfo: &pti,
		UserInfo:  &pui,
	}

	return &pi, nil
}
