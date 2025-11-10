package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/cleanbuddy/backend/internal/models"
)

// AvailabilityService handles business logic for availability
type AvailabilityService struct {
	cleanerRepo      *models.CleanerRepository
	availabilityRepo *models.AvailabilityRepository
}

// NewAvailabilityService creates a new availability service
func NewAvailabilityService(db *sql.DB) *AvailabilityService {
	return &AvailabilityService{
		cleanerRepo:      models.NewCleanerRepository(db),
		availabilityRepo: models.NewAvailabilityRepository(db),
	}
}

// CreateAvailability creates a new availability slot
func (s *AvailabilityService) CreateAvailability(
	cleanerID string,
	availabilityType string,
	dayOfWeek *int,
	specificDate *time.Time,
	startTime string,
	endTime string,
	isActive *bool,
	notes *string,
) (*models.Availability, error) {
	// Validate cleaner exists
	cleaner, err := s.cleanerRepo.GetByID(cleanerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaner: %w", err)
	}
	if cleaner == nil {
		return nil, fmt.Errorf("cleaner not found")
	}

	// Default isActive to true if not provided
	active := true
	if isActive != nil {
		active = *isActive
	}

	// Prepare availability
	availability := &models.Availability{
		CleanerID: cleanerID,
		Type:      availabilityType,
		StartTime: startTime,
		EndTime:   endTime,
		IsActive:  active,
	}

	if dayOfWeek != nil {
		availability.DayOfWeek = sql.NullInt32{Int32: int32(*dayOfWeek), Valid: true}
	}

	if specificDate != nil {
		availability.SpecificDate = sql.NullTime{Time: *specificDate, Valid: true}
	}

	if notes != nil {
		availability.Notes = sql.NullString{String: *notes, Valid: true}
	}

	// Check for conflicts
	hasConflict, err := s.availabilityRepo.CheckConflict(
		cleanerID,
		availabilityType,
		availability.DayOfWeek,
		availability.SpecificDate,
		startTime,
		endTime,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to check conflict: %w", err)
	}
	if hasConflict {
		return nil, fmt.Errorf("availability slot conflicts with existing schedule")
	}

	// Create availability
	err = s.availabilityRepo.Create(availability)
	if err != nil {
		return nil, fmt.Errorf("failed to create availability: %w", err)
	}

	return availability, nil
}

// UpdateAvailability updates an existing availability slot
func (s *AvailabilityService) UpdateAvailability(
	availabilityID string,
	cleanerID string,
	startTime *string,
	endTime *string,
	isActive *bool,
	notes *string,
) (*models.Availability, error) {
	// Get existing availability
	availability, err := s.availabilityRepo.GetByID(availabilityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get availability: %w", err)
	}
	if availability == nil {
		return nil, fmt.Errorf("availability not found")
	}

	// Ownership check
	if availability.CleanerID != cleanerID {
		return nil, fmt.Errorf("unauthorized: you can only update your own availability")
	}

	// Update fields
	if startTime != nil {
		availability.StartTime = *startTime
	}
	if endTime != nil {
		availability.EndTime = *endTime
	}
	if isActive != nil {
		availability.IsActive = *isActive
	}
	if notes != nil {
		availability.Notes = sql.NullString{String: *notes, Valid: true}
	}

	// Update in database
	err = s.availabilityRepo.Update(availability)
	if err != nil {
		return nil, fmt.Errorf("failed to update availability: %w", err)
	}

	return availability, nil
}

// DeleteAvailability deletes an availability slot
func (s *AvailabilityService) DeleteAvailability(availabilityID, cleanerID string) error {
	// Get existing availability
	availability, err := s.availabilityRepo.GetByID(availabilityID)
	if err != nil {
		return fmt.Errorf("failed to get availability: %w", err)
	}
	if availability == nil {
		return fmt.Errorf("availability not found")
	}

	// Ownership check
	if availability.CleanerID != cleanerID {
		return fmt.Errorf("unauthorized: you can only delete your own availability")
	}

	// Delete from database
	err = s.availabilityRepo.Delete(availabilityID)
	if err != nil {
		return fmt.Errorf("failed to delete availability: %w", err)
	}

	return nil
}

// GetCleanerAvailability gets all availability for a cleaner
func (s *AvailabilityService) GetCleanerAvailability(cleanerID, userID string) ([]*models.Availability, error) {
	// Verify cleaner exists
	cleaner, err := s.cleanerRepo.GetByID(cleanerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaner: %w", err)
	}
	if cleaner == nil {
		return nil, fmt.Errorf("cleaner not found")
	}

	// Get availability
	availabilities, err := s.availabilityRepo.GetByCleanerID(cleanerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaner availability: %w", err)
	}

	return availabilities, nil
}

// SetWeeklySchedule bulk sets recurring weekly schedule
func (s *AvailabilityService) SetWeeklySchedule(cleanerID, userID string, weeklySlots []struct {
	DayOfWeek int
	StartTime string
	EndTime   string
	Notes     *string
}) ([]*models.Availability, error) {
	// Verify cleaner exists and user is the cleaner
	cleaner, err := s.cleanerRepo.GetByID(cleanerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaner: %w", err)
	}
	if cleaner == nil {
		return nil, fmt.Errorf("cleaner not found")
	}
	if cleaner.UserID != userID {
		return nil, fmt.Errorf("unauthorized: you can only set your own schedule")
	}

	// Create availability slots
	var availabilities []*models.Availability
	for _, slot := range weeklySlots {
		active := true
		availability, err := s.CreateAvailability(
			cleanerID,
			models.AvailabilityTypeRecurring,
			&slot.DayOfWeek,
			nil,
			slot.StartTime,
			slot.EndTime,
			&active,
			slot.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create weekly slot: %w", err)
		}
		availabilities = append(availabilities, availability)
	}

	return availabilities, nil
}
