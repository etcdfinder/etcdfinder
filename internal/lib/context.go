package lib

import (
	"context"
)

// ContextKey is a type for the keys of values stored in the context
type ContextKey string

const (
	CtxRequestID ContextKey = "request_id"
)

func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(CtxRequestID).(string); ok {
		return requestID
	}
	return ""
}
