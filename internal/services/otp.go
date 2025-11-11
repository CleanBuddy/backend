package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// OTPService handles OTP generation and validation
type OTPService struct {
	redis        *redis.Client
	emailService *EmailService
	// In-memory fallback when Redis is unavailable
	mu       sync.RWMutex
	fallback map[string]*otpEntry
}

type otpEntry struct {
	code      string
	expiresAt time.Time
}

// NewOTPService creates a new OTP service
func NewOTPService(redisClient *redis.Client, emailService *EmailService) *OTPService {
	svc := &OTPService{
		redis:        redisClient,
		emailService: emailService,
		fallback:     make(map[string]*otpEntry),
	}

	// Start cleanup goroutine for in-memory fallback
	go svc.cleanupExpiredOTPs(1 * time.Minute)

	return svc
}

// GenerateOTP generates a 6-digit OTP code
func (s *OTPService) GenerateOTP() (string, error) {
	// In development, always return 123456 for easier testing
	if os.Getenv("ENV") == "development" {
		return "123456", nil
	}

	// Generate random 6-digit code
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%06d", n.Int64()), nil
}

// StoreOTP stores OTP in Redis (or in-memory fallback) with 5-minute expiration
func (s *OTPService) StoreOTP(ctx context.Context, email, code string) error {
	// Try Redis first if available
	if s.redis != nil {
		key := fmt.Sprintf("otp:%s", email)
		err := s.redis.Set(ctx, key, code, 5*time.Minute).Err()
		if err == nil {
			return nil
		}
		// If Redis fails, fall through to in-memory storage
		fmt.Printf("⚠️  Redis unavailable, using in-memory OTP storage\n")
	}

	// Fallback to in-memory storage
	s.mu.Lock()
	defer s.mu.Unlock()

	s.fallback[email] = &otpEntry{
		code:      code,
		expiresAt: time.Now().Add(5 * time.Minute),
	}

	return nil
}

// VerifyOTP verifies OTP code and deletes it if valid
func (s *OTPService) VerifyOTP(ctx context.Context, email, code string) (bool, error) {
	// Try Redis first if available
	if s.redis != nil {
		key := fmt.Sprintf("otp:%s", email)

		storedCode, err := s.redis.Get(ctx, key).Result()
		if err == nil {
			// Found in Redis
			if storedCode != code {
				return false, nil // Invalid code
			}
			// Delete OTP after successful verification
			s.redis.Del(ctx, key)
			return true, nil
		}

		if err != redis.Nil {
			// Redis error (not just missing key), fall through to in-memory
			fmt.Printf("⚠️  Redis error during OTP verification, checking in-memory storage\n")
		}
	}

	// Check in-memory fallback
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.fallback[email]
	if !exists {
		return false, nil // OTP not found
	}

	// Check if expired
	if time.Now().After(entry.expiresAt) {
		delete(s.fallback, email)
		return false, nil // OTP expired
	}

	if entry.code != code {
		return false, nil // Invalid code
	}

	// Delete OTP after successful verification
	delete(s.fallback, email)

	return true, nil
}

// SendOTP sends OTP via email using the email service
func (s *OTPService) SendOTP(email, code string) error {
	ctx := context.Background()

	// In development mode, just log the OTP instead of sending email
	if os.Getenv("ENV") == "development" {
		fmt.Printf("\n🔐 DEVELOPMENT MODE - OTP for %s: %s\n", email, code)
		fmt.Printf("   Use this code to login (valid for 5 minutes)\n\n")
		return nil
	}

	// Send OTP email
	if err := s.emailService.SendOTPEmail(ctx, email, code); err != nil {
		return fmt.Errorf("failed to send OTP email: %w", err)
	}

	fmt.Printf("✅ OTP email sent successfully to %s\n", email)
	return nil
}

// cleanupExpiredOTPs periodically removes expired OTP entries from in-memory storage
func (s *OTPService) cleanupExpiredOTPs(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for email, entry := range s.fallback {
			if now.After(entry.expiresAt) {
				delete(s.fallback, email)
			}
		}
		s.mu.Unlock()
	}
}
