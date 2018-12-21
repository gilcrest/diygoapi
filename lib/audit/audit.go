package audit

import "time"

// Audit represents fields meant to track when
// an action was taken and by whom
type Audit struct {
	CreateClientID  string
	CreateUsername  string
	CreateTimestamp time.Time
	UpdateClientID  string
	UpdateUsername  string
	UpdateTimestamp time.Time
}
