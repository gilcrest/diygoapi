package controller

import (
	"io"
	"net/http"

	"github.com/pkg/errors"

	"github.com/gilcrest/go-api-basic/domain/errs"
	"github.com/rs/zerolog/hlog"
)

// StandardResponse is meant to be included in all non-error
// response bodies and includes "standard" response fields
type StandardResponse struct {
	Path      string      `json:"path,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
	Data      interface{} `json:"data"`
}

// NewStandardResponse is an initializer for the StandardResponse struct
func NewStandardResponse(r *http.Request, d interface{}) (*StandardResponse, error) {
	var sr StandardResponse
	sr.Path = r.URL.EscapedPath()
	// gets Trace ID from request
	id, ok := hlog.IDFromRequest(r)
	if !ok {
		return nil, errs.E(errors.New("trace ID not properly set to request context"))
	}
	sr.RequestID = id.String()

	sr.Data = d

	return &sr, nil
}

// DecoderErr handles an error returned by json.NewDecoder(r.Body).Decode(&data)
// this function will determine the appropriate error response
func DecoderErr(err error) error {
	switch {
	// If the request body is empty (io.EOF)
	// return an error
	case err == io.EOF:
		return errs.E(errs.InvalidRequest, errors.New("Request Body cannot be empty"))
	// If the request body has malformed JSON (io.ErrUnexpectedEOF)
	// return an error
	case err == io.ErrUnexpectedEOF:
		return errs.E(errs.InvalidRequest, errors.New("Malformed JSON"))
	// return all other errors
	case err != nil:
		return errs.E(err)
	}
	return nil
}
