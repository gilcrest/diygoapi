package response

import (
	"context"
)

// Audit struct should be used for all responses
type Audit struct {
	RequestID  string `json:"id"`
	RequestURL string `json:"url"`
}

// NewAudit is a constructor for the Audit struct
func NewAudit(ctx context.Context) (*Audit, error) {
	info := new(Audit)
	info.RequestID = "fakeRequestID" // TODO
	info.RequestURL = "fakeURL"
	return info, nil
}
