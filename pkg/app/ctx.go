package app

import "context"

type key string

const (
	userIDKey key = "userID"
	requestID key = "requestID"
)

func WithUserID[T any](ctx context.Context, uid T) context.Context {
	return context.WithValue(ctx, userIDKey, uid)
}

func UserID[T any](ctx context.Context) (T, bool) {
	uid, ok := ctx.Value(userIDKey).(T)
	return uid, ok
}

func WithRequestID[T any](ctx context.Context, rid T) context.Context {
	return context.WithValue(ctx, requestID, rid)
}

func RequestID[T any](ctx context.Context) (T, bool) {
	rid, ok := ctx.Value(requestID).(T)
	return rid, ok
}
