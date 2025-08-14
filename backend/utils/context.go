package utils

import "context"

// Context key types for storing values in context
type contextKey string

// Context keys for storing user information in request context
const (
	// UserIDKey is the key used to store user ID in context
	UserIDKey contextKey = "userID"
	// UsernameKey is the key used to store username in context
	UsernameKey contextKey = "username"
	// UserRoleKey is the key used to store user role in context
	UserRoleKey contextKey = "role"
)

// GetUserIDFromContext 获取用户ID
func GetUserIDFromContext(ctx context.Context) (uint, bool) {
	userID, ok := ctx.Value(UserIDKey).(uint)
	return userID, ok
}

// GetUsernameFromContext 获取用户名
func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(UsernameKey).(string)
	return username, ok
}

// GetUserRoleFromContext 获取用户角色
func GetUserRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleKey).(string)
	return role, ok
}
