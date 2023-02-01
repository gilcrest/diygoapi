package diygoapi

import (
	"context"

	"github.com/google/uuid"

	"github.com/gilcrest/diygoapi/errs"
	"github.com/gilcrest/diygoapi/secure"
)

// OrgServicer manages the retrieval and manipulation of an Org
type OrgServicer interface {
	// Create manages the creation of an Org (and optional app)
	Create(ctx context.Context, r *CreateOrgRequest, adt Audit) (*OrgResponse, error)
	Update(ctx context.Context, r *UpdateOrgRequest, adt Audit) (*OrgResponse, error)
	Delete(ctx context.Context, extlID string) (DeleteResponse, error)
	FindAll(ctx context.Context) ([]*OrgResponse, error)
	FindByExternalID(ctx context.Context, extlID string) (*OrgResponse, error)
}

// OrgKind is a way of classifying an organization. Examples are Genesis, Test, Standard
type OrgKind struct {
	// ID: The unique identifier
	ID uuid.UUID
	// External ID: The unique external identifier
	ExternalID string
	// Description: A longer description of the organization kind
	Description string
}

// Validate determines whether the Person has proper data to be considered valid
func (o OrgKind) Validate() error {
	const op errs.Op = "diygoapi/OrgKind.Validate"

	switch {
	case o.ID == uuid.Nil:
		return errs.E(op, errs.Validation, "OrgKind ID cannot be nil")
	case o.ExternalID == "":
		return errs.E(op, errs.Validation, "OrgKind ExternalID cannot be empty")
	case o.Description == "":
		return errs.E(op, errs.Validation, "OrgKind Description cannot be empty")
	}

	return nil
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
	Kind *OrgKind
}

// Validate determines whether the Org has proper data to be considered valid
func (o Org) Validate() (err error) {
	const op errs.Op = "diygoapi/Org.Validate"

	switch {
	case o.ID == uuid.Nil:
		return errs.E(op, errs.Validation, "Org ID cannot be nil")
	case o.ExternalID.String() == "":
		return errs.E(op, errs.Validation, "Org ExternalID cannot be empty")
	case o.Name == "":
		return errs.E(op, errs.Validation, "Org Name cannot be empty")
	case o.Description == "":
		return errs.E(op, errs.Validation, "Org Description cannot be empty")
	}

	if err = o.Kind.Validate(); err != nil {
		return errs.E(op, err)
	}

	return nil
}

// CreateOrgRequest is the request struct for Creating an Org
type CreateOrgRequest struct {
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	Kind             string            `json:"kind"`
	CreateAppRequest *CreateAppRequest `json:"app"`
}

// Validate determines whether the CreateOrgRequest has proper data to be considered valid
func (r CreateOrgRequest) Validate() error {
	const op errs.Op = "diygoapi/CreateOrgRequest.Validate"

	switch {
	case r.Name == "":
		return errs.E(op, errs.Validation, "org name is required")
	case r.Description == "":
		return errs.E(op, errs.Validation, "org description is required")
	case r.Kind == "":
		return errs.E(op, errs.Validation, "org kind is required")
	}
	return nil
}

// UpdateOrgRequest is the request struct for Updating an Org
type UpdateOrgRequest struct {
	ExternalID  string
	Name        string `json:"name"`
	Description string `json:"description"`
}

// OrgResponse is the response struct for an Org.
// It contains only one app (even though an org can have many apps).
// This app is only present in the response when creating an org and
// accompanying app. I may change this later to be different response
// structs for different purposes, but for now, this works.
type OrgResponse struct {
	ExternalID          string       `json:"external_id"`
	Name                string       `json:"name"`
	KindExternalID      string       `json:"kind_description"`
	Description         string       `json:"description"`
	CreateAppExtlID     string       `json:"create_app_extl_id"`
	CreateUserFirstName string       `json:"create_user_first_name"`
	CreateUserLastName  string       `json:"create_user_last_name"`
	CreateDateTime      string       `json:"create_date_time"`
	UpdateAppExtlID     string       `json:"update_app_extl_id"`
	UpdateUserFirstName string       `json:"update_user_first_name"`
	UpdateUserLastName  string       `json:"update_user_last_name"`
	UpdateDateTime      string       `json:"update_date_time"`
	App                 *AppResponse `json:"app,omitempty"`
}
