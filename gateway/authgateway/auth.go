package authgateway

import (
	"context"

	"golang.org/x/oauth2"
	googleoauth "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"

	"github.com/gilcrest/errs"
)

// UserInfo makes an outbound https call to Google using their
// Oauth2 v2 api and returns a Userinfo struct which has most
// profile data elements you typically need
func UserInfo(ctx context.Context, token *oauth2.Token) (*googleoauth.Userinfo, error) {
	const op errs.Op = "gateway/authgateway/UserInfo"

	oauthService, err := googleoauth.NewService(ctx, option.WithTokenSource(oauth2.StaticTokenSource(token)))
	if err != nil {
		return nil, errs.E(op, err)
	}

	userInfo, err := oauthService.Userinfo.Get().Do()
	if err != nil {
		// "In summary, a 401 Unauthorized response should be used for missing or
		// bad authentication, and a 403 Forbidden response should be used afterwards,
		// when the user is authenticated but isn’t authorized to perform the
		// requested operation on the given resource."
		// In this case, we are getting a bad response from Google service, assume
		// they are not able to authenticate properly
		return nil, errs.E(op, errs.Unauthenticated, err)
	}

	return userInfo, nil
}
