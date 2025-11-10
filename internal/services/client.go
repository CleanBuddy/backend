package services

import (
	"database/sql"
	"fmt"

	"github.com/cleanbuddy/backend/internal/models"
)

// ClientService handles client profile business logic
type ClientService struct {
	clientRepo *models.ClientRepository
	userRepo   *models.UserRepository
}

// NewClientService creates a new client service
func NewClientService(db *sql.DB) *ClientService {
	return &ClientService{
		clientRepo: models.NewClientRepository(db),
		userRepo:   models.NewUserRepository(db),
	}
}

// GetOrCreateClientProfile gets or creates a client profile for a user
func (s *ClientService) GetOrCreateClientProfile(userID string) (*models.Client, error) {
	// Check if client profile exists
	client, err := s.clientRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get client profile: %w", err)
	}

	// If exists, return it
	if client != nil {
		return client, nil
	}

	// Otherwise, create new client profile
	client = &models.Client{
		UserID:            userID,
		PreferredLanguage: "ro",
		TotalBookings:     0,
		TotalSpent:        0.0,
	}

	if err := s.clientRepo.Create(client); err != nil {
		return nil, fmt.Errorf("failed to create client profile: %w", err)
	}

	return client, nil
}

// UpdateClientProfile updates client profile and user information
func (s *ClientService) UpdateClientProfile(userID string, firstName, lastName, phoneNumber, preferredLanguage *string) (*models.Client, error) {
	// Get or create client profile
	client, err := s.GetOrCreateClientProfile(userID)
	if err != nil {
		return nil, err
	}

	// Update client fields if provided
	if phoneNumber != nil && *phoneNumber != "" {
		client.PhoneNumber = sql.NullString{String: *phoneNumber, Valid: true}
	}
	if preferredLanguage != nil && *preferredLanguage != "" {
		client.PreferredLanguage = *preferredLanguage
	}

	// Update client profile
	if err := s.clientRepo.Update(client); err != nil {
		return nil, fmt.Errorf("failed to update client profile: %w", err)
	}

	// Update user fields if provided
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if firstName != nil && *firstName != "" {
		user.FirstName = sql.NullString{String: *firstName, Valid: true}
	}
	if lastName != nil && *lastName != "" {
		user.LastName = sql.NullString{String: *lastName, Valid: true}
	}
	if phoneNumber != nil && *phoneNumber != "" {
		user.Phone = sql.NullString{String: *phoneNumber, Valid: true}
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return client, nil
}
