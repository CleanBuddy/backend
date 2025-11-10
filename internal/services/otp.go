package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// OTPService handles OTP generation and validation
type OTPService struct {
	redis        *redis.Client
	emailService *EmailService
}

// NewOTPService creates a new OTP service
func NewOTPService(redisClient *redis.Client, emailService *EmailService) *OTPService {
	return &OTPService{
		redis:        redisClient,
		emailService: emailService,
	}
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

// StoreOTP stores OTP in Redis with 5-minute expiration
func (s *OTPService) StoreOTP(ctx context.Context, email, code string) error {
	key := fmt.Sprintf("otp:%s", email)
	return s.redis.Set(ctx, key, code, 5*time.Minute).Err()
}

// VerifyOTP verifies OTP code and deletes it if valid
func (s *OTPService) VerifyOTP(ctx context.Context, email, code string) (bool, error) {
	key := fmt.Sprintf("otp:%s", email)

	storedCode, err := s.redis.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil // OTP not found or expired
	}
	if err != nil {
		return false, err
	}

	if storedCode != code {
		return false, nil // Invalid code
	}

	// Delete OTP after successful verification
	s.redis.Del(ctx, key)

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
