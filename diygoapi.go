// Package diygoapi comprises application or business domain data types and functions.
package diygoapi

import (
	"context"
	"time"
)

// LoggerServicer reads and updates the logger state
type LoggerServicer interface {
	Read() *LoggerResponse
	Update(r *LoggerRequest) (*LoggerResponse, error)
}

// GenesisServicer initializes the database with dependent data
type GenesisServicer interface {
	// Arche creates the initial seed data in the database.
	Arche(ctx context.Context, r *GenesisRequest) (GenesisResponse, error)
	// ReadConfig reads the local config file generated as part of Seed (when run locally).
	// Is only a utility to help with local testing.
	ReadConfig() (GenesisResponse, error)
}

// Audit represents the moment an App/User interacted with the system.
type Audit struct {
	App    *App
	User   *User
	Moment time.Time
}

// SimpleAudit captures the first time a record was written as well
// as the last time the record was updated. The first time a record
// is written Create and Update will be identical.
type SimpleAudit struct {
	Create Audit `json:"create"`
	Update Audit `json:"update"`
}

// DeleteResponse is the general response struct for things
// which have been deleted
type DeleteResponse struct {
	ExternalID string `json:"extl_id"`
	Deleted    bool   `json:"deleted"`
}

// LoggerRequest is the request struct for the app logger
type LoggerRequest struct {
	GlobalLogLevel string `json:"global_log_level"`
	LogErrorStack  string `json:"log_error_stack"`
}

// LoggerResponse is the response struct for the current
// state of the app logger
type LoggerResponse struct {
	LoggerMinimumLevel string `json:"logger_minimum_level"`
	GlobalLogLevel     string `json:"global_log_level"`
	LogErrorStack      bool   `json:"log_error_stack"`
}

// GenesisRequest is the request struct for the genesis service
type GenesisRequest struct {
	User struct {
		// Provider: The Oauth2 provider.
		Provider string `json:"provider"`

		// Token: The Oauth2 token to be used to create the user.
		Token string `json:"token"`
	} `json:"user"`

	UserInitiatedOrg CreateOrgRequest `json:"org"`

	// PermissionRequests: The list of permissions to be created as part of Genesis
	CreatePermissionRequests []CreatePermissionRequest `json:"permissions"`

	// CreateRoleRequests: The list of Roles to be created as part of Genesis
	CreateRoleRequests []CreateRoleRequest `json:"roles"`
}

// GenesisResponse contains both the Genesis response and the Test response
type GenesisResponse struct {
	Principal     *OrgResponse `json:"principal"`
	Test          *OrgResponse `json:"test"`
	UserInitiated *OrgResponse `json:"userInitiated,omitempty"`
}
