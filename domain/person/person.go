package person

import (
	"time"

	"github.com/google/uuid"

	"github.com/gilcrest/diy-go-api/domain/org"
)

// Person is a single person that exists within the system for an Org
// A Person can have multiple Profiles
type Person struct {
	// id: The unique identifier of the Person.
	ID uuid.UUID

	// org: The Org the person belongs to.
	Org org.Org
}

// Profile is profile information about a Person
type Profile struct {
	// id: The unique identifier for the Person's profile
	ID uuid.UUID

	// person: The person that is the owner of the profile
	Person Person

	// namePrefix: The name prefix for the Profile (e.g. Mx., Ms., Mr., etc.)
	NamePrefix string

	// FirstName: The person's first name.
	FirstName string

	// MiddleName: The person's middle name.
	MiddleName string

	// LastName: The person's last name.
	LastName string

	// FullName: The person's full name.
	FullName string

	// nameSuffix: The name suffix for the person's name (e.g. "PhD", "CCNA", "OBE").
	// Other examples include generational designations like "Sr." and "Jr." and "I", "II", "III", etc.
	NameSuffix string

	// Nickname: The person's nickname
	Nickname string

	// CompanyName: The Company Name that the person works at
	CompanyName string

	// CompanyDepartment: is the department at the company that the person works at
	CompanyDepartment string

	// JobTitle: The person's Job Title
	JobTitle string

	// BirthDate: The full birthdate of a person (e.g. Dec 18, 1953)
	BirthDate time.Time

	// LanguageID: TODO - setup a Language struct, lookup table, etc. using TBD ISO standard
	LanguageID uuid.UUID

	// HostedDomain: The hosted domain e.g. example.com.
	HostedDomain string

	// PictureURL: URL of the person's picture image for the profile.
	PictureURL string

	// ProfileLink: URL of the profile page.
	ProfileLink string

	// ProfileSource: The source of the profile (e.g. Google Oauth2, Apple Oauth2, etc.)
	ProfileSource string
}
