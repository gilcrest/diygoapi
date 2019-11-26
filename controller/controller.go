package controller

import (
	"net/http"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/rs/xid"
)

// StandardResponseFields is meant to be included in all response bodies
// and includes "standard" response fields
type StandardResponseFields struct {
	Path      string `json:"path,omitempty"`
	RequestID xid.ID `json:"request_id"`
}

// NewStandardResponseFields is an initializer for the StandardResponseFields struct
func NewStandardResponseFields(id xid.ID, r *http.Request) *StandardResponseFields {
	const op errs.Op = "controller/moviectl/NewStandardResponse"

	sr := new(StandardResponseFields)
	sr.RequestID = id
	sr.Path = r.URL.EscapedPath()

	return sr
}
