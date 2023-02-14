package diygoapi

import (
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/text/language"

	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/secure"
)

// RegisterUserServicer registers a new user
type RegisterUserServicer interface {
	SelfRegister(ctx context.Context, adt Audit) error
}

// Person - from Wikipedia: "A person (plural people or persons) is a being that
// has certain capacities or attributes such as reason, morality, consciousness or
// self-consciousness, and being a part of a culturally established form of social
// relations such as kinship, ownership of property, or legal responsibility.
//
// The defining features of personhood and, consequently, what makes a person count
// as a person, differ widely among cultures and contexts."
//
// A Person can have multiple Users.
type Person struct {
	// ID: The unique identifier of the Person.
	ID uuid.UUID

	// ExternalID: unique external identifier of the Person
	ExternalID secure.Identifier

	// Users: All the users that are linked to the Person
	// (e.g. a GitHub user, a Google user, etc.).
	Users []*User
}

// NullUUID returns ID as uuid.NullUUID
func (p Person) NullUUID() uuid.NullUUID {
	if p.ID == uuid.Nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{
		UUID:  p.ID,
		Valid: true,
	}
}

// Validate determines whether the Person has proper data to be considered valid
func (p Person) Validate() (err error) {
	const op errs.Op = "diygoapi/Person.Validate"

	switch {
	case p.ID == uuid.Nil:
		return errs.E(op, errs.Validation, "Person ID cannot be nil")
	case p.ExternalID.String() == "":
		return errs.E(op, errs.Validation, "Person ExternalID cannot be empty")
	}

	return nil
}

// UserResponse - from Wikipedia: "A user is a person who utilizes a computer or network service." In the context of this
// project, given that we allow Persons to authenticate with multiple providers, a User is akin to a persona
// (Wikipedia - "The word persona derives from Latin, where it originally referred to a theatrical mask. On the
// social web, users develop virtual personas as online identities.") and as such, a Person can have one or many
// Users (for instance, I can have a GitHub user and a Google user, but I am just one Person).
//
// As a general, practical matter, most operations are considered at the User level. For instance, roles are
// assigned at the user level instead of the Person level, which allows for more fine-grained access control.
type UserResponse struct {
	// ID: The unique identifier for the Person's profile
	ID uuid.UUID

	// ExternalID: unique external identifier of the User
	ExternalID secure.Identifier `json:"external_id"`

	// NamePrefix: The name prefix for the Profile (e.g. Mx., Ms., Mr., etc.)
	NamePrefix string `json:"name_prefix"`

	// FirstName: The person's first name.
	FirstName string `json:"first_name"`

	// MiddleName: The person's middle name.
	MiddleName string `json:"middle_name"`

	// LastName: The person's last name.
	LastName string `json:"last_name"`

	// FullName: The person's full name.
	FullName string `json:"full_name"`

	// NameSuffix: The name suffix for the person's name (e.g. "PhD", "CCNA", "OBE").
	// Other examples include generational designations like "Sr." and "Jr." and "I", "II", "III", etc.
	NameSuffix string `json:"name_suffix"`

	// Nickname: The person's nickname
	Nickname string `json:"nickname"`

	// Email: The primary email for the User
	Email string `json:"email"`

	// CompanyName: The Company Name that the person works at
	CompanyName string `json:"company_name"`

	// CompanyDepartment: is the department at the company that the person works at
	CompanyDepartment string `json:"company_department"`

	// JobTitle: The person's Job Title
	JobTitle string `json:"job_title"`

	// BirthDate: The full birthdate of a person (e.g. Dec 18, 1953)
	BirthDate time.Time `json:"birth_date"`

	// LanguagePreferences is the user's language tag preferences.
	LanguagePreferences []language.Tag `json:"language_preferences"`

	// HostedDomain: The hosted domain e.g. example.com.
	HostedDomain string `json:"hosted_domain"`

	// PictureURL: URL of the person's picture image for the profile.
	PictureURL string `json:"picture_url"`

	// ProfileLink: URL of the profile page.
	ProfileLink string `json:"profile_link"`

	// Source: The origin of the User (e.g. Google Oauth2, Apple Oauth2, etc.)
	Source string `json:"source"`
}

// User - from Wikipedia: "A user is a person who utilizes a computer or network service." In the context of this
// project, given that we allow Persons to authenticate with multiple providers, a User is akin to a persona
// (Wikipedia - "The word persona derives from Latin, where it originally referred to a theatrical mask. On the
// social web, users develop virtual personas as online identities.") and as such, a Person can have one or many
// Users (for instance, I can have a GitHub user and a Google user, but I am just one Person).
//
// As a general, practical matter, most operations are considered at the User level. For instance, roles are
// assigned at the user level instead of the Person level, which allows for more fine-grained access control.
type User struct {
	// ID: The unique identifier for the Person's profile
	ID uuid.UUID

	// ExternalID: unique external identifier of the User
	ExternalID secure.Identifier

	// NamePrefix: The name prefix for the Profile (e.g. Mx., Ms., Mr., etc.)
	NamePrefix string

	// FirstName: The person's first name.
	FirstName string

	// MiddleName: The person's middle name.
	MiddleName string

	// LastName: The person's last name.
	LastName string

	// FullName: The person's full name.
	FullName string

	// NameSuffix: The name suffix for the person's name (e.g. "PhD", "CCNA", "OBE").
	// Other examples include generational designations like "Sr." and "Jr." and "I", "II", "III", etc.
	NameSuffix string

	// Nickname: The person's nickname
	Nickname string

	// Gender: The user's gender. TODO - setup Gender properly. not binary.
	Gender string

	// Email: The primary email for the User
	Email string

	// CompanyName: The Company Name that the person works at
	CompanyName string

	// CompanyDepartment: is the department at the company that the person works at
	CompanyDepartment string

	// JobTitle: The person's Job Title
	JobTitle string

	// BirthDate: The full birthdate of a person (e.g. Dec 18, 1953)
	BirthDate time.Time

	// LanguagePreferences is the user's language tag preferences.
	LanguagePreferences []language.Tag

	// HostedDomain: The hosted domain e.g. example.com.
	HostedDomain string

	// PictureURL: URL of the person's picture image for the profile.
	PictureURL string

	// ProfileLink: URL of the profile page.
	ProfileLink string

	// Source: The origin of the User (e.g. Google Oauth2, Apple Oauth2, etc.)
	Source string
}

// Validate determines whether the Person has proper data to be considered valid
func (u User) Validate() error {
	const op errs.Op = "diygoapi/User.Validate"

	switch {
	case u.ID == uuid.Nil:
		return errs.E(op, errs.Validation, "User ID cannot be nil")
	case u.ExternalID.String() == "":
		return errs.E(op, errs.Validation, "User ExternalID cannot be empty")
	case u.LastName == "":
		return errs.E(op, errs.Validation, "User LastName cannot be empty")
	case u.FirstName == "":
		return errs.E(op, errs.Validation, "User FirstName cannot be empty")
	}

	return nil
}

// NullUUID returns ID as uuid.NullUUID
func (u User) NullUUID() uuid.NullUUID {
	if u.ID == uuid.Nil {
		return uuid.NullUUID{}
	}
	return uuid.NullUUID{
		UUID:  u.ID,
		Valid: true,
	}
}

// NewUserFromProviderInfo creates a new User struct to be used in db user creation
func NewUserFromProviderInfo(pi *ProviderInfo, lm language.Matcher) *User {
	var langPrefs []language.Tag
	langPref, _, _ := lm.Match(language.Make(pi.UserInfo.Locale))
	langPrefs = append(langPrefs, langPref)

	// create User from ProviderInfo
	u := &User{
		ID:                  uuid.New(),
		ExternalID:          secure.NewID(),
		NamePrefix:          pi.UserInfo.NamePrefix,
		FirstName:           pi.UserInfo.FirstName,
		MiddleName:          pi.UserInfo.MiddleName,
		LastName:            pi.UserInfo.LastName,
		FullName:            pi.UserInfo.FullName,
		NameSuffix:          pi.UserInfo.NameSuffix,
		Nickname:            pi.UserInfo.Nickname,
		Gender:              pi.UserInfo.Gender,
		Email:               pi.UserInfo.Email,
		BirthDate:           pi.UserInfo.BirthDate,
		LanguagePreferences: langPrefs,
		HostedDomain:        pi.UserInfo.HostedDomain,
		PictureURL:          pi.UserInfo.Picture,
		ProfileLink:         pi.UserInfo.ProfileLink,
		Source:              pi.Provider.String(),
	}

	return u
}
