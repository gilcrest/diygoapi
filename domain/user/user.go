// Package user holds details about a person who is using the application
package user

// User holds details of a User from Google
type User struct {
	// Email: The user's email address.
	Email string `json:"email,omitempty"`

	// LastName: The user's last name.
	LastName string `json:"last_name,omitempty"`

	// FirstName: The user's first name.
	FirstName string `json:"first_name,omitempty"`

	// FullName: The user's full name.
	FullName string `json:"full_name,omitempty"`

	// Gender: The user's gender.
	//Gender string `json:"gender,omitempty"`

	// HostedDomain: The hosted domain e.g. example.com if the user
	// is Google apps user.
	HostedDomain string `json:"hosted_domain,omitempty"`

	// PictureURL: URL of the user's picture image.
	PictureURL string `json:"picture_url,omitempty"`

	// ProfileLink: URL of the profile page.
	ProfileLink string `json:"profile_link,omitempty"`
}

// IsValid determines whether or not the User has proper
// data to be considered valid
func (u User) IsValid() bool {
	switch {
	case u.Email == "":
		return false
	case u.FirstName == "":
		return false
	case u.LastName == "":
		return false
	}
	return true
}
