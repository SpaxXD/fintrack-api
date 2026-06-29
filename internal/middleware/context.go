package middleware

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	// RequestIDKey is the context key for the request ID.
	RequestIDKey contextKey = "request_id"
	// UserIDKey is the context key for the authenticated user ID.
	UserIDKey contextKey = "user_id"

	// unexported aliases used internally by tests and middleware
	requestIDKey = RequestIDKey
	userIDKey    = UserIDKey
)

// GetRequestID retrieves the request ID from the context.
// Returns an empty string if not set.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserID retrieves the authenticated user ID from the context.
// Returns uuid.Nil if not set.
func GetUserID(ctx context.Context) uuid.UUID {
	if id, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return id
	}
	return uuid.Nil
}
