package middleware

import (
	"context"
	"net/http"
)

type responseWriterKey struct{}

// ResponseWriterMiddleware injects the HTTP response writer into context
// This allows GraphQL resolvers to set cookies
func ResponseWriterMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), responseWriterKey{}, w)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetResponseWriter extracts response writer from context
func GetResponseWriter(ctx context.Context) (http.ResponseWriter, bool) {
	w, ok := ctx.Value(responseWriterKey{}).(http.ResponseWriter)
	return w, ok
}
