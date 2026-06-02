package session

import "context"

type ctxKeyType struct{}

var ctxKey = ctxKeyType{}

func WithSessionID(ctx context.Context, sid string) context.Context {
	return context.WithValue(ctx, ctxKey, sid)
}

func SessionIDFromContext(ctx context.Context) string {
	sid, _ := ctx.Value(ctxKey).(string)
	return sid
}
