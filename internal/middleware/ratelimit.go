package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// RateLimiter implements token bucket rate limiting using Redis
type RateLimiter struct {
	redisClient *redis.Client
	mu          sync.RWMutex
	// Fallback in-memory limiter when Redis is unavailable
	fallback map[string]*bucket
}

type bucket struct {
	tokens    int
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter with Redis backing
func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
	return &RateLimiter{
		redisClient: redisClient,
		fallback:    make(map[string]*bucket),
	}
}

// RateLimitMiddleware creates HTTP middleware for rate limiting
// - Anonymous requests: 20 requests per minute per IP
// - Authenticated requests: 100 requests per minute per user
func (rl *RateLimiter) RateLimitMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get identifier (userID if authenticated, IP otherwise)
			var identifier string
			userID, ok := GetUserIDFromContext(r.Context())
			if ok && userID != "" {
				identifier = "user:" + userID
			} else {
				identifier = "ip:" + getClientIP(r)
			}

			// Check rate limit
			allowed, err := rl.checkLimit(r.Context(), identifier, ok && userID != "")
			if err != nil {
				// Log error but allow request (fail open for availability)
				fmt.Printf("Rate limit check error: %v\n", err)
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"RATE_LIMIT_EXCEEDED","message":"Too many requests. Please try again later."}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// checkLimit checks if the request is within rate limits
// Returns true if allowed, false if rate limit exceeded
func (rl *RateLimiter) checkLimit(ctx context.Context, identifier string, authenticated bool) (bool, error) {
	// Set limits based on authentication status
	maxTokens := 20      // Anonymous: 20 req/min
	refillRate := 20     // Refill 20 tokens per minute
	windowSeconds := 60

	if authenticated {
		maxTokens = 100      // Authenticated: 100 req/min
		refillRate = 100
	}

	// If Redis is not available, use in-memory fallback
	if rl.redisClient == nil {
		return rl.checkLimitFallback(identifier, maxTokens, refillRate), nil
	}

	// Try Redis-based rate limiting first
	key := fmt.Sprintf("ratelimit:%s", identifier)

	// Use Redis INCR with expiration for simple sliding window
	count, err := rl.redisClient.Incr(ctx, key).Result()
	if err != nil {
		// Fallback to in-memory rate limiting if Redis fails
		return rl.checkLimitFallback(identifier, maxTokens, refillRate), nil
	}

	// Set expiration on first request
	if count == 1 {
		rl.redisClient.Expire(ctx, key, time.Duration(windowSeconds)*time.Second)
	}

	return count <= int64(maxTokens), nil
}

// checkLimitFallback provides in-memory rate limiting when Redis is unavailable
func (rl *RateLimiter) checkLimitFallback(identifier string, maxTokens int, refillRate int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.fallback[identifier]

	if !exists {
		// Create new bucket
		rl.fallback[identifier] = &bucket{
			tokens:    maxTokens - 1,
			lastRefill: now,
		}
		return true
	}

	// Calculate token refill
	elapsed := now.Sub(b.lastRefill)
	tokensToAdd := int(elapsed.Minutes() * float64(refillRate))

	if tokensToAdd > 0 {
		b.tokens = min(maxTokens, b.tokens+tokensToAdd)
		b.lastRefill = now
	}

	// Check if tokens available
	if b.tokens > 0 {
		b.tokens--
		return true
	}

	return false
}

// getClientIP extracts the client IP address from the request
// Checks X-Forwarded-For header first (for proxies), then falls back to RemoteAddr
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (set by proxies/load balancers)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		return forwarded
	}

	// Check X-Real-IP header (alternative proxy header)
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fallback to RemoteAddr
	return r.RemoteAddr
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CleanupFallbackCache periodically cleans up old entries from in-memory cache
// Should be called as a background goroutine
func (rl *RateLimiter) CleanupFallbackCache(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, b := range rl.fallback {
			// Remove entries older than 5 minutes
			if now.Sub(b.lastRefill) > 5*time.Minute {
				delete(rl.fallback, key)
			}
		}
		rl.mu.Unlock()
	}
}
