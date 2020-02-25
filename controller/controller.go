package controller

import (
	"fmt"
	"net/http"

	"github.com/rs/xid"
)

// RequestID is the unique Request ID for each request
type TraceID struct {
	xid.ID
}

// StandardResponseFields is meant to be included in all response bodies
// and includes "standard" response fields
type StandardResponseFields struct {
	Path    string  `json:"path,omitempty"`
	TraceID TraceID `json:"trace_id,omitempty"`
}

// NewTraceID is an initializer for TraceID
func NewTraceID(id xid.ID) TraceID {
	return TraceID{ID: id}
}

// NewMockTraceID is an initializer for TraceID which returns a
// static "mocked"
func NewMockTraceID() TraceID {
	x, err := xid.FromString("bpa182jipt3b2b78879g")
	if err != nil {
		fmt.Println(err)
	}
	return TraceID{ID: x}
}

// NewStandardResponseFields is an initializer for the StandardResponseFields struct
func NewStandardResponseFields(id TraceID, r *http.Request) StandardResponseFields {
	var sr StandardResponseFields
	sr.TraceID = id
	sr.Path = r.URL.EscapedPath()

	return sr
}
