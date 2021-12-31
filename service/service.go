// Package service orchestrates components between handlers and other
// packages (datastore, gateway, domain, etc.)
package service

import (
	"time"

	"github.com/gilcrest/go-api-basic/domain/audit"
)

// CryptoRandomGenerator is the interface that generates random data
type CryptoRandomGenerator interface {
	RandomBytes(n int) ([]byte, error)
	RandomString(n int) (string, error)
}

// auditResponse is to be embedded into other structs for the purpose
// of displaying audit information.
type auditResponse struct {
	AppExternalID string `json:"app_external_id"`
	AppName       string `json:"app_name"`
	Username      string `json:"username"`
	AuditTime     string `json:"audit_time"`
}

func newAuditResponse(adt audit.Audit) auditResponse {
	return auditResponse{
		AppExternalID: adt.App.ExternalID.String(),
		AppName:       adt.App.Name,
		Username:      adt.User.Username,
		AuditTime:     adt.Moment.Format(time.RFC3339),
	}
}
