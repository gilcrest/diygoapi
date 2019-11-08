package audit

import (
	"time"

	"github.com/google/uuid"
)

// Audit represents fields meant to track when
// an action was taken and by whom
type Audit struct {
	CreateClientID  uuid.UUID
	CreatePersonID  uuid.UUID
	CreateTimestamp time.Time
	UpdateClientID  uuid.UUID
	UpdatePersonID  uuid.UUID
	UpdateTimestamp time.Time
}
