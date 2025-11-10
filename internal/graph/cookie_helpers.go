package graph

import (
	"net/http"
	"os"
)

// setAuthCookie sets a secure httpOnly authentication cookie
func setAuthCookie(w http.ResponseWriter, token string) {
	// Check if we're in production
	isProduction := os.Getenv("ENV") == "production"

	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		MaxAge:   7 * 24 * 60 * 60, // 7 days (matches JWT expiration)
		HttpOnly: true,              // Prevents JavaScript access (XSS protection)
		Secure:   isProduction,      // HTTPS only in production
		SameSite: http.SameSiteLaxMode, // CSRF protection with better compatibility (changed from Strict)
	}

	http.SetCookie(w, cookie)
}

// clearAuthCookie clears the authentication cookie
func clearAuthCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Delete cookie
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: http.SameSiteLaxMode, // Match the SameSite mode used when setting the cookie
	}

	http.SetCookie(w, cookie)
}
