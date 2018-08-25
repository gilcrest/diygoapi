package response

import (
	"context"

	"github.com/gilcrest/go-API-template/server/todo"
)

// Audit struct should be used for all responses
type Audit struct {
	RequestID  string `json:"id"`
	RequestURL string `json:"url"`
}

// NewAudit is a constructor for the Audit struct
func NewAudit(ctx context.Context) (*Audit, error) {
	info := new(Audit)
	info.RequestID = todo.ID(ctx)
	info.RequestURL = "fakeURL"
	return info, nil
}
