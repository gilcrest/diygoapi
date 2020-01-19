package controller

import (
	"net/http"

	"github.com/gilcrest/errs"
	"github.com/rs/xid"
)

// RequestID is the unique Request ID for each request
type RequestID struct {
	xid.ID
}

// StandardResponseFields is meant to be included in all response bodies
// and includes "standard" response fields
type StandardResponseFields struct {
	Path string    `json:"path,omitempty"`
	ID   RequestID `json:"request_id"`
}

// NewRequestID is an initializer for RequestID
func NewRequestID(id xid.ID) RequestID {
	return RequestID{ID: id}
}

// NewStandardResponseFields is an initializer for the StandardResponseFields struct
func NewStandardResponseFields(id RequestID, r *http.Request) StandardResponseFields {
	const op errs.Op = "controller/moviectl/NewStandardResponse"

	var sr StandardResponseFields
	sr.ID = id
	sr.Path = r.URL.EscapedPath()

	return sr
}
