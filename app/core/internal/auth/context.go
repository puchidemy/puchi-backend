package auth

import "context"

type contextKey string

const userIDKey contextKey = "user_id"

// UserIDFromContext extracts the authenticated user ID from context.
// Returns the user ID string and true if found, empty string and false otherwise.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDKey).(string)
	return id, ok && id != ""
}

// NewContextWithUserID creates a new context with the user ID set.
func NewContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
