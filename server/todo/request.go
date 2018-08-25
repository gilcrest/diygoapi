package todo

import (
	"context"

	"github.com/rs/xid"
)

type contextKey string

func (c contextKey) String() string {
	return "context key " + string(c)
}

// RequestID is a unique identifier for each inbound request
var requestID = contextKey("RequestID")

// SetRequestID adds a unique ID as RequestID to the context
func SetRequestID(ctx context.Context) context.Context {
	// get byte Array representation of guid from xid package (12 bytes)
	guid := xid.New()

	// use the String method of the guid object to convert byte array to string (20 bytes)
	rID := guid.String()

	ctx = context.WithValue(ctx, requestID, rID)

	return ctx

}

// ID gets the Request ID from the context.
func ID(ctx context.Context) string {
	requestIDstr := ctx.Value(requestID).(string)
	return requestIDstr
}
