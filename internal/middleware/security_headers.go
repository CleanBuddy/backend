package middleware

import (
	"net/http"
	"os"
)

// SecurityHeadersMiddleware adds security headers to all responses
// Implements OWASP best practices for HTTP security headers
func SecurityHeadersMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			isProduction := os.Getenv("ENV") == "production"

			// Content Security Policy (CSP)
			// Prevents XSS attacks by controlling which resources can be loaded
			if isProduction {
				// Strict CSP for production
				w.Header().Set("Content-Security-Policy",
					"default-src 'self'; "+
						"script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+ // GraphQL Playground needs unsafe-eval in dev
						"style-src 'self' 'unsafe-inline'; "+
						"img-src 'self' data: https:; "+
						"font-src 'self' data:; "+
						"connect-src 'self' https://api.sidemail.io; "+ // Sidemail API
						"frame-ancestors 'none'; "+
						"base-uri 'self'; "+
						"form-action 'self'")
			} else {
				// Relaxed CSP for development (GraphQL Playground)
				w.Header().Set("Content-Security-Policy",
					"default-src 'self'; "+
						"script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
						"style-src 'self' 'unsafe-inline'; "+
						"img-src 'self' data: https:; "+
						"font-src 'self' data:; "+
						"connect-src 'self' https://api.sidemail.io; "+
						"frame-ancestors 'self'; "+
						"base-uri 'self'; "+
						"form-action 'self'")
			}

			// HTTP Strict Transport Security (HSTS)
			// Forces HTTPS connections for 1 year (only in production)
			if isProduction {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
			}

			// X-Frame-Options: Prevents clickjacking attacks
			// DENY = page cannot be displayed in iframe/frame/embed/object
			w.Header().Set("X-Frame-Options", "DENY")

			// X-Content-Type-Options: Prevents MIME type sniffing
			// nosniff = browser must respect the Content-Type header
			w.Header().Set("X-Content-Type-Options", "nosniff")

			// X-XSS-Protection: Legacy XSS filter (modern browsers use CSP instead)
			// 1; mode=block = enable XSS filter and block page if attack detected
			w.Header().Set("X-XSS-Protection", "1; mode=block")

			// Referrer-Policy: Controls how much referrer information is sent
			// strict-origin-when-cross-origin = send full URL for same-origin, origin only for cross-origin
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions-Policy (formerly Feature-Policy)
			// Restricts browser features to prevent abuse
			w.Header().Set("Permissions-Policy",
				"accelerometer=(), "+
					"camera=(), "+
					"geolocation=(self), "+ // Allow geolocation for check-in/check-out
					"gyroscope=(), "+
					"magnetometer=(), "+
					"microphone=(), "+
					"payment=(), "+
					"usb=()")

			// X-Permitted-Cross-Domain-Policies: Restricts Adobe Flash/PDF cross-domain access
			w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")

			// Cache-Control: Prevent caching of sensitive data
			// Only for non-static assets (static assets should have their own cache headers)
			if r.URL.Path != "/" && r.URL.Path != "/health" {
				w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, private")
				w.Header().Set("Pragma", "no-cache")
				w.Header().Set("Expires", "0")
			}

			next.ServeHTTP(w, r)
		})
	}
}
