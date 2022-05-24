package org

import (
	"github.com/gilcrest/diy-go-api/domain/secure"
	"github.com/google/uuid"
)

// Kind is a way of classifying an organization. Examples are Genesis, Test, Standard
type Kind struct {
	// ID: The unique identifier
	ID uuid.UUID
	// External ID: The unique external identifier
	ExternalID string
	// Description: A longer description of the organization kind
	Description string
}

// Org represents an Organization (company, institution or any other
// organized body of people with a particular purpose)
type Org struct {
	// ID: The unique identifier
	ID uuid.UUID
	// External ID: The unique external identifier
	ExternalID secure.Identifier
	// Name: The organization name
	Name string
	// Description: A longer description of the organization
	Description string
	// Kind: a way of classifying organizations
	Kind Kind
}
