package services

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/cleanbuddy/backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// AuthService handles authentication logic
type AuthService struct {
	userRepo   *models.UserRepository
	otpService *OTPService
	jwtSecret  []byte
}

// NewAuthService creates a new auth service
func NewAuthService(db *sql.DB, redisClient *redis.Client, emailService *EmailService) *AuthService {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "dev-secret-change-in-production"
	}

	return &AuthService{
		userRepo:   models.NewUserRepository(db),
		otpService: NewOTPService(redisClient, emailService),
		jwtSecret:  []byte(jwtSecret),
	}
}

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// RequestOTP generates and sends OTP to user's email
func (s *AuthService) RequestOTP(ctx context.Context, email string) error {
	// Basic email validation
	if len(email) < 5 || !contains(email, "@") {
		return fmt.Errorf("invalid email address")
	}

	// Generate OTP
	code, err := s.otpService.GenerateOTP()
	if err != nil {
		return fmt.Errorf("failed to generate OTP: %w", err)
	}

	// Store OTP in Redis
	if err := s.otpService.StoreOTP(ctx, email, code); err != nil {
		return fmt.Errorf("failed to store OTP: %w", err)
	}

	// Send OTP via email
	if err := s.otpService.SendOTP(email, code); err != nil {
		return fmt.Errorf("failed to send OTP: %w", err)
	}

	return nil
}

// contains checks if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// LoginWithOTP verifies OTP and returns JWT token
// LoginWithOTPWithRole logs in with OTP and creates user with specified role if new
func (s *AuthService) LoginWithOTPWithRole(ctx context.Context, email, code string, role models.UserRole) (string, *models.User, error) {
	// Verify OTP
	valid, err := s.otpService.VerifyOTP(ctx, email, code)
	if err != nil {
		return "", nil, fmt.Errorf("failed to verify OTP: %w", err)
	}
	if !valid {
		return "", nil, fmt.Errorf("invalid or expired OTP")
	}

	// Get or create user
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user == nil {
		// Create new user with specified role
		user = &models.User{
			Email:         sql.NullString{String: email, Valid: true},
			Role:          role,
			IsActive:      true,
			EmailVerified: true,
		}
		if err := s.userRepo.Create(user); err != nil {
			return "", nil, fmt.Errorf("failed to create user: %w", err)
		}
	} else {
		// Mark email as verified if not already
		if !user.EmailVerified {
			user.EmailVerified = true
			if err := s.userRepo.Update(user); err != nil {
				return "", nil, fmt.Errorf("failed to update user: %w", err)
			}
		}
	}

	return s.generateTokenForUser(ctx, user)
}

// LoginWithOTP logs in with OTP (defaults to CLIENT role for new users)
func (s *AuthService) LoginWithOTP(ctx context.Context, email, code string) (string, *models.User, error) {
	return s.LoginWithOTPWithRole(ctx, email, code, models.RoleClient)
}

// generateTokenForUser is a helper to generate token for authenticated user
func (s *AuthService) generateTokenForUser(ctx context.Context, user *models.User) (string, *models.User, error) {

	// Check if user is active
	if !user.IsActive {
		return "", nil, fmt.Errorf("user account is inactive")
	}

	// Generate JWT token
	token, err := s.GenerateToken(user)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return token, user, nil
}

// GenerateToken generates a JWT token for user
func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID: user.ID,
		Role:   string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour * 7)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

// ValidateToken validates JWT token and returns claims
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// GetUserByID gets user by ID
func (s *AuthService) GetUserByID(userID string) (*models.User, error) {
	return s.userRepo.GetByID(userID)
}

// UpdateUserProfile updates user profile information
func (s *AuthService) UpdateUserProfile(userID string, firstName, lastName, email *string) (*models.User, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update fields if provided
	if firstName != nil {
		user.FirstName = sql.NullString{String: *firstName, Valid: *firstName != ""}
	}
	if lastName != nil {
		user.LastName = sql.NullString{String: *lastName, Valid: *lastName != ""}
	}
	if email != nil {
		// Check if email is already taken
		existingUser, err := s.userRepo.GetByEmail(*email)
		if err != nil {
			return nil, fmt.Errorf("failed to check email: %w", err)
		}
		if existingUser != nil && existingUser.ID != userID {
			return nil, fmt.Errorf("email already in use")
		}
		user.Email = sql.NullString{String: *email, Valid: *email != ""}
	}

	// Update user
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}
