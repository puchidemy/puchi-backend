package auth

import "context"

type contextKey string

const (
	userIDKey    contextKey = "user_id"
	userEmailKey contextKey = "user_email"
	userRolesKey contextKey = "user_roles"
)

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

// NewContextWithRoles creates a new context with the user roles set.
func NewContextWithRoles(ctx context.Context, roles []string) context.Context {
	return context.WithValue(ctx, userRolesKey, roles)
}

// NewContextWithEmail creates a new context with the user email set.
func NewContextWithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, userEmailKey, email)
}

// EmailFromContext extracts the authenticated user email from context.
func EmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(userEmailKey).(string)
	return email, ok && email != ""
}

// RolesFromContext extracts the user roles from context.
// Returns the roles slice and true if found.
func RolesFromContext(ctx context.Context) ([]string, bool) {
	v, ok := ctx.Value(userRolesKey).([]string)
	return v, ok
}

// HasRole checks whether the context contains the given role.
func HasRole(ctx context.Context, role string) bool {
	roles, ok := RolesFromContext(ctx)
	if !ok {
		return false
	}
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}
