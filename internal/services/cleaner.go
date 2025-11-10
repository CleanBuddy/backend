package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/cleanbuddy/backend/internal/models"
	"github.com/cleanbuddy/backend/internal/utils"
)

// CleanerService handles cleaner profile business logic
type CleanerService struct {
	cleanerRepo  *models.CleanerRepository
	userRepo     *models.UserRepository
	emailService *EmailService
}

// NewCleanerService creates a new cleaner service
func NewCleanerService(db *sql.DB, emailService *EmailService) *CleanerService {
	return &CleanerService{
		cleanerRepo:  models.NewCleanerRepository(db),
		userRepo:     models.NewUserRepository(db),
		emailService: emailService,
	}
}

// GetOrCreateCleanerProfile gets or creates a cleaner profile for a user
func (s *CleanerService) GetOrCreateCleanerProfile(userID string) (*models.Cleaner, error) {
	// Check if cleaner profile exists
	cleaner, err := s.cleanerRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaner profile: %w", err)
	}

	// If exists, return it
	if cleaner != nil {
		return cleaner, nil
	}

	return nil, fmt.Errorf("cleaner profile not found")
}

// CreateCleanerProfile creates a new cleaner profile
func (s *CleanerService) CreateCleanerProfile(
	userID string,
	phoneNumber string,
	dateOfBirth *string,
	streetAddress, city, county, postalCode *string,
	yearsOfExperience int,
	bio *string,
	specializations, languages []string,
) (*models.Cleaner, error) {
	// Check if profile already exists
	existing, err := s.cleanerRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing profile: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("cleaner profile already exists")
	}

	// Update user role to CLEANER
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	user.Role = models.RoleCleaner
	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user role: %w", err)
	}

	// Prepare JSONB fields
	var specializationsJSON, languagesJSON []byte
	if specializations == nil {
		specializations = []string{}
	}
	if languages == nil {
		languages = []string{"ro"}
	}

	specializationsJSON, err = json.Marshal(specializations)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal specializations: %w", err)
	}

	languagesJSON, err = json.Marshal(languages)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal languages: %w", err)
	}

	// Create cleaner profile
	cleaner := &models.Cleaner{
		UserID:            userID,
		PhoneNumber:       phoneNumber,
		YearsOfExperience: yearsOfExperience,
		Specializations:   specializationsJSON,
		Languages:         languagesJSON,
		ApprovalStatus:    models.ApprovalStatusPending,
		IsActive:          true,
		IsAvailable:       false,
		TotalJobs:         0,
		TotalEarnings:     0.0,
	}

	// Set optional fields
	if dateOfBirth != nil && *dateOfBirth != "" {
		parsedDate, err := time.Parse("2006-01-02", *dateOfBirth)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
		cleaner.DateOfBirth = sql.NullTime{Time: parsedDate, Valid: true}
	}

	if streetAddress != nil && *streetAddress != "" {
		cleaner.StreetAddress = sql.NullString{String: *streetAddress, Valid: true}
	}
	if city != nil && *city != "" {
		cleaner.City = sql.NullString{String: *city, Valid: true}
	}
	if county != nil && *county != "" {
		cleaner.County = sql.NullString{String: *county, Valid: true}
	}
	if postalCode != nil && *postalCode != "" {
		cleaner.PostalCode = sql.NullString{String: *postalCode, Valid: true}
	}
	if bio != nil && *bio != "" {
		cleaner.Bio = sql.NullString{String: *bio, Valid: true}
	}

	if err := s.cleanerRepo.Create(cleaner); err != nil {
		return nil, fmt.Errorf("failed to create cleaner profile: %w", err)
	}

	return cleaner, nil
}

// UpdateCleanerProfile updates a cleaner profile
func (s *CleanerService) UpdateCleanerProfile(
	userID string,
	phoneNumber *string,
	dateOfBirth *string,
	streetAddress, city, county, postalCode *string,
	yearsOfExperience *int,
	bio *string,
	specializations, languages *[]string,
	iban *string,
	isAvailable *bool,
) (*models.Cleaner, error) {
	// Get existing profile
	cleaner, err := s.cleanerRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaner profile: %w", err)
	}
	if cleaner == nil {
		return nil, fmt.Errorf("cleaner profile not found")
	}

	// Update fields if provided
	if phoneNumber != nil && *phoneNumber != "" {
		cleaner.PhoneNumber = *phoneNumber
	}

	if dateOfBirth != nil && *dateOfBirth != "" {
		parsedDate, err := time.Parse("2006-01-02", *dateOfBirth)
		if err != nil {
			return nil, fmt.Errorf("invalid date format: %w", err)
		}
		cleaner.DateOfBirth = sql.NullTime{Time: parsedDate, Valid: true}
	}

	if streetAddress != nil {
		if *streetAddress != "" {
			cleaner.StreetAddress = sql.NullString{String: *streetAddress, Valid: true}
		} else {
			cleaner.StreetAddress = sql.NullString{Valid: false}
		}
	}

	if city != nil {
		if *city != "" {
			cleaner.City = sql.NullString{String: *city, Valid: true}
		} else {
			cleaner.City = sql.NullString{Valid: false}
		}
	}

	if county != nil {
		if *county != "" {
			cleaner.County = sql.NullString{String: *county, Valid: true}
		} else {
			cleaner.County = sql.NullString{Valid: false}
		}
	}

	if postalCode != nil {
		if *postalCode != "" {
			cleaner.PostalCode = sql.NullString{String: *postalCode, Valid: true}
		} else {
			cleaner.PostalCode = sql.NullString{Valid: false}
		}
	}

	if yearsOfExperience != nil {
		cleaner.YearsOfExperience = *yearsOfExperience
	}

	if bio != nil {
		if *bio != "" {
			cleaner.Bio = sql.NullString{String: *bio, Valid: true}
		} else {
			cleaner.Bio = sql.NullString{Valid: false}
		}
	}

	if specializations != nil {
		specializationsJSON, err := json.Marshal(*specializations)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal specializations: %w", err)
		}
		cleaner.Specializations = specializationsJSON
	}

	if languages != nil {
		languagesJSON, err := json.Marshal(*languages)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal languages: %w", err)
		}
		cleaner.Languages = languagesJSON
	}

	if iban != nil {
		log.Printf("UpdateCleanerProfile: IBAN provided: %s", *iban)
		if *iban != "" {
			// Encrypt IBAN before storing
			encryptedIBAN, err := utils.EncryptIBAN(*iban)
			if err != nil {
				log.Printf("ERROR: Failed to encrypt IBAN: %v", err)
				return nil, fmt.Errorf("failed to encrypt IBAN: %w", err)
			}
			log.Printf("UpdateCleanerProfile: IBAN encrypted successfully, length: %d", len(encryptedIBAN))
			cleaner.IBAN = sql.NullString{String: encryptedIBAN, Valid: true}
		} else {
			log.Printf("UpdateCleanerProfile: IBAN is empty string")
			cleaner.IBAN = sql.NullString{Valid: false}
		}
	} else {
		log.Printf("UpdateCleanerProfile: IBAN is nil, not updating")
	}

	if isAvailable != nil {
		cleaner.IsAvailable = *isAvailable
	}

	if err := s.cleanerRepo.Update(cleaner); err != nil {
		return nil, fmt.Errorf("failed to update cleaner profile: %w", err)
	}

	return cleaner, nil
}

// UploadDocument updates document URLs for a cleaner
func (s *CleanerService) UploadDocument(userID, documentType, fileURL string) (*models.Cleaner, error) {
	// Get existing profile
	cleaner, err := s.cleanerRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaner profile: %w", err)
	}
	if cleaner == nil {
		return nil, fmt.Errorf("cleaner profile not found")
	}

	// Update document URL based on type
	switch documentType {
	case "id_document":
		cleaner.IDDocumentURL = sql.NullString{String: fileURL, Valid: true}
	case "background_check":
		cleaner.BackgroundCheckURL = sql.NullString{String: fileURL, Valid: true}
	case "profile_photo":
		cleaner.ProfilePhotoURL = sql.NullString{String: fileURL, Valid: true}
	default:
		return nil, fmt.Errorf("invalid document type: %s", documentType)
	}

	if err := s.cleanerRepo.Update(cleaner); err != nil {
		return nil, fmt.Errorf("failed to update cleaner: %w", err)
	}

	return cleaner, nil
}

// GetCleanerByID gets a cleaner by ID
func (s *CleanerService) GetCleanerByID(id string) (*models.Cleaner, error) {
	return s.cleanerRepo.GetByID(id)
}

// ApproveCleanerProfile approves a cleaner profile (admin only)
func (s *CleanerService) ApproveCleanerProfile(cleanerID, adminID string) (*models.Cleaner, error) {
	cleaner, err := s.cleanerRepo.GetByID(cleanerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaner: %w", err)
	}
	if cleaner == nil {
		return nil, fmt.Errorf("cleaner not found")
	}

	cleaner.ApprovalStatus = models.ApprovalStatusApproved
	cleaner.ApprovedBy = sql.NullString{String: adminID, Valid: true}
	now := time.Now()
	cleaner.ApprovedAt = sql.NullTime{Time: now, Valid: true}

	if err := s.cleanerRepo.Update(cleaner); err != nil {
		return nil, fmt.Errorf("failed to approve cleaner: %w", err)
	}

	// Send approval email to cleaner (async)
	go func() {
		ctx := context.Background()
		user, err := s.userRepo.GetByID(cleaner.UserID)
		if err == nil && user != nil && user.Email.Valid {
			cleanerName := "Cleaner"
			if user.FirstName.Valid {
				cleanerName = user.FirstName.String
			}

			_ = s.emailService.SendCleanerApprovedEmail(
				ctx,
				user.Email.String,
				cleanerName,
			)
		}
	}()

	return cleaner, nil
}

// RejectCleanerProfile rejects a cleaner profile (admin only)
func (s *CleanerService) RejectCleanerProfile(cleanerID, adminID, reason string) (*models.Cleaner, error) {
	cleaner, err := s.cleanerRepo.GetByID(cleanerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaner: %w", err)
	}
	if cleaner == nil {
		return nil, fmt.Errorf("cleaner not found")
	}

	cleaner.ApprovalStatus = models.ApprovalStatusRejected
	cleaner.ApprovedBy = sql.NullString{String: adminID, Valid: true}
	now := time.Now()
	cleaner.ApprovedAt = sql.NullTime{Time: now, Valid: true}
	cleaner.RejectedReason = sql.NullString{String: reason, Valid: true}

	if err := s.cleanerRepo.Update(cleaner); err != nil {
		return nil, fmt.Errorf("failed to reject cleaner: %w", err)
	}

	// Send rejection email to cleaner (async)
	go func() {
		ctx := context.Background()
		user, err := s.userRepo.GetByID(cleaner.UserID)
		if err == nil && user != nil && user.Email.Valid {
			cleanerName := "Cleaner"
			if user.FirstName.Valid {
				cleanerName = user.FirstName.String
			}

			_ = s.emailService.SendCleanerRejectedEmail(
				ctx,
				user.Email.String,
				cleanerName,
				reason,
			)
		}
	}()

	return cleaner, nil
}

// GetPendingCleaners gets all cleaners with pending approval status (admin only)
func (s *CleanerService) GetPendingCleaners() ([]*models.Cleaner, error) {
	return s.cleanerRepo.GetByApprovalStatus(models.ApprovalStatusPending)
}

// GetApprovedCleaners returns a list of approved cleaners with pagination and search
func (s *CleanerService) GetApprovedCleaners(limit, offset int, search string) ([]*models.Cleaner, error) {
	return s.cleanerRepo.GetApprovedCleaners(limit, offset, search)
}

// GetCleaners returns a list of cleaners with optional filters
func (s *CleanerService) GetCleaners(limit, offset int, status *string, search *string) ([]*models.Cleaner, error) {
	return s.cleanerRepo.GetCleaners(limit, offset, status, search)
}
