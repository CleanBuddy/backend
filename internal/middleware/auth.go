package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/cleanbuddy/backend/internal/services"
)

type contextKey string

const (
	UserIDKey   contextKey = "userID"
	UserRoleKey contextKey = "userRole"
)

// AuthMiddleware validates JWT token and adds userID to context
func AuthMiddleware(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string

			// Try to get token from cookie first (more secure)
			cookie, err := r.Cookie("auth_token")
			if err == nil && cookie.Value != "" {
				token = cookie.Value
			} else {
				// Fallback to Authorization header for backwards compatibility
				authHeader := r.Header.Get("Authorization")
				if authHeader == "" {
					next.ServeHTTP(w, r)
					return
				}

				// Extract token (format: "Bearer <token>")
				parts := strings.Split(authHeader, " ")
				if len(parts) != 2 || parts[0] != "Bearer" {
					next.ServeHTTP(w, r)
					return
				}

				token = parts[1]
			}

			// Validate token
			claims, err := authService.ValidateToken(token)
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}

			// Add userID and role to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext extracts userID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}

// GetUserRoleFromContext extracts user role from context
func GetUserRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(UserRoleKey).(string)
	return role, ok
}
