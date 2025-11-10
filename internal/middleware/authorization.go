package middleware

import (
	"context"
	"fmt"
)

// UserRole constants matching the database enum
const (
	RoleClient         = "CLIENT"
	RoleCleaner        = "CLEANER"
	RoleCompanyAdmin   = "COMPANY_ADMIN"
	RolePlatformAdmin  = "PLATFORM_ADMIN"
)

// RequireAuth checks if user is authenticated
// Returns userID and error if not authenticated
func RequireAuth(ctx context.Context) (string, error) {
	userID, ok := GetUserIDFromContext(ctx)
	if !ok || userID == "" {
		return "", fmt.Errorf("AUTHENTICATION_REQUIRED: You must be logged in to perform this action")
	}
	return userID, nil
}

// RequireRole checks if user is authenticated and has the specified role
// Returns userID and error if not authenticated or wrong role
func RequireRole(ctx context.Context, allowedRoles ...string) (string, error) {
	userID, err := RequireAuth(ctx)
	if err != nil {
		return "", err
	}

	userRole, ok := GetUserRoleFromContext(ctx)
	if !ok || userRole == "" {
		return "", fmt.Errorf("PERMISSION_DENIED: Unable to verify user role")
	}

	// Check if user's role is in allowed roles
	for _, role := range allowedRoles {
		if userRole == role {
			return userID, nil
		}
	}

	return "", fmt.Errorf("PERMISSION_DENIED: This action requires %v role. Your role: %s", allowedRoles, userRole)
}

// RequireAdmin checks if user is authenticated and has PLATFORM_ADMIN role
// This is a convenience wrapper for admin-only operations
func RequireAdmin(ctx context.Context) (string, error) {
	return RequireRole(ctx, RolePlatformAdmin)
}

// RequireAdminOrCompanyAdmin allows both platform admins and company admins
func RequireAdminOrCompanyAdmin(ctx context.Context) (string, error) {
	return RequireRole(ctx, RolePlatformAdmin, RoleCompanyAdmin)
}

// RequireCleanerRole checks if user is authenticated and has CLEANER role
func RequireCleanerRole(ctx context.Context) (string, error) {
	return RequireRole(ctx, RoleCleaner)
}

// RequireClientRole checks if user is authenticated and has CLIENT role
func RequireClientRole(ctx context.Context) (string, error) {
	return RequireRole(ctx, RoleClient)
}

// IsAdmin checks if the current user has admin role (without returning error)
func IsAdmin(ctx context.Context) bool {
	userRole, ok := GetUserRoleFromContext(ctx)
	return ok && userRole == RolePlatformAdmin
}

// HasRole checks if the current user has any of the specified roles (without returning error)
func HasRole(ctx context.Context, roles ...string) bool {
	userRole, ok := GetUserRoleFromContext(ctx)
	if !ok {
		return false
	}

	for _, role := range roles {
		if userRole == role {
			return true
		}
	}
	return false
}
