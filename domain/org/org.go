package org

import (
	"time"

	"github.com/gilcrest/go-api-basic/domain/secure"
	"github.com/google/uuid"
)

// Org represents an Organization (company, institution or any other
// organized body of people with a particular purpose)
type Org struct {
	ID           uuid.UUID
	ExternalID   secure.Identifier
	Name         string
	Description  string
	CreateAppID  uuid.UUID
	CreateUserID uuid.UUID
	CreateTime   time.Time
	UpdateAppID  uuid.UUID
	UpdateUserID uuid.UUID
	UpdateTime   time.Time
}
