package middleware

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	TraceIDKey   contextKey = "trace_id"
)

func Tracing() func(c context.Context) context.Context {
	return func(c context.Context) context.Context {
		traceID := uuid.New().String()
		requestID := uuid.New().String()

		ctx := context.WithValue(c, TraceIDKey, traceID)
		ctx = context.WithValue(ctx, RequestIDKey, requestID)

		return ctx
	}
}

func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return ""
}

func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}
