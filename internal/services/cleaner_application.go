package services

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/cleanbuddy/backend/internal/models"
)

// CleanerApplicationService handles cleaner application business logic
type CleanerApplicationService struct {
	appRepo *models.CleanerApplicationRepository
	db      *sql.DB
}

// NewCleanerApplicationService creates a new cleaner application service
func NewCleanerApplicationService(db *sql.DB) *CleanerApplicationService {
	return &CleanerApplicationService{
		appRepo: models.NewCleanerApplicationRepository(db),
		db:      db,
	}
}

// GetBySessionID retrieves an application by session ID (for anonymous users)
func (s *CleanerApplicationService) GetBySessionID(sessionID string) (*models.CleanerApplication, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	app, err := s.appRepo.GetBySessionID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	return app, nil
}

// GetByUserID retrieves application for a user
func (s *CleanerApplicationService) GetByUserID(userID string) (*models.CleanerApplication, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	app, err := s.appRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}

	return app, nil
}

// GetPendingApplications retrieves all pending applications for admin review
func (s *CleanerApplicationService) GetPendingApplications(limit, offset int) ([]*models.CleanerApplication, error) {
	apps, err := s.appRepo.GetPendingApplications(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending applications: %w", err)
	}

	return apps, nil
}

// SaveApplication creates or updates a cleaner application
func (s *CleanerApplicationService) SaveApplication(input *models.CleanerApplicationInput) (*models.CleanerApplication, error) {
	// Need either session_id or user_id
	if (input.SessionID == nil || *input.SessionID == "") && input.UserID == nil {
		return nil, fmt.Errorf("session ID or user ID is required")
	}

	// Check if application already exists
	var existingApp *models.CleanerApplication
	var err error

	if input.SessionID != nil && *input.SessionID != "" {
		existingApp, err = s.appRepo.GetBySessionID(*input.SessionID)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing application: %w", err)
		}
	} else if input.UserID != nil {
		userIDStr := input.UserID.String()
		existingApp, err = s.appRepo.GetByUserID(userIDStr)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing application: %w", err)
		}
	}

	// Convert input to application data
	appData := input.ToApplicationData()

	if existingApp != nil {
		// Update existing application
		existingApp.ApplicationData = appData
		existingApp.CurrentStep = input.CurrentStep

		// Update user_id if provided (auth happened)
		if input.UserID != nil {
			existingApp.UserID = input.UserID
		}

		if err := s.appRepo.Update(existingApp); err != nil {
			return nil, fmt.Errorf("failed to update application: %w", err)
		}

		return existingApp, nil
	}

	// Create new application
	app := &models.CleanerApplication{
		SessionID:       input.SessionID,
		UserID:          input.UserID,
		ApplicationData: appData,
		CurrentStep:     input.CurrentStep,
		Status:          models.ApplicationStatusDraft,
	}

	if err := s.appRepo.Create(app); err != nil {
		return nil, fmt.Errorf("failed to create application: %w", err)
	}

	return app, nil
}

// SubmitApplication submits an application for review
func (s *CleanerApplicationService) SubmitApplication(applicationID, userID string) (*models.CleanerApplication, error) {
	// Get the application
	app, err := s.appRepo.GetByID(applicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		return nil, fmt.Errorf("application not found")
	}

	// Verify ownership
	if app.UserID == nil || app.UserID.String() != userID {
		return nil, fmt.Errorf("unauthorized")
	}

	// Validate application
	if err := app.CanSubmit(); err != nil {
		return nil, fmt.Errorf("cannot submit application: %w", err)
	}

	// Mark as submitted
	if err := s.appRepo.MarkAsSubmitted(applicationID); err != nil {
		return nil, fmt.Errorf("failed to submit application: %w", err)
	}

	// Refresh application
	app, err = s.appRepo.GetByID(applicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated application: %w", err)
	}

	// TODO: Send notification to admin
	// TODO: Trigger background check process

	return app, nil
}

// ReviewApplication handles admin approval/rejection
func (s *CleanerApplicationService) ReviewApplication(
	applicationID, adminID string,
	approve bool,
	rejectionReason, adminNotes string,
) (*models.CleanerApplication, error) {
	// Get the application
	app, err := s.appRepo.GetByID(applicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get application: %w", err)
	}
	if app == nil {
		return nil, fmt.Errorf("application not found")
	}

	// Validate status
	if app.Status != models.ApplicationStatusSubmitted && app.Status != models.ApplicationStatusUnderReview {
		return nil, fmt.Errorf("application is not in reviewable state")
	}

	// Validate rejection reason if rejecting
	if !approve && rejectionReason == "" {
		return nil, fmt.Errorf("rejection reason is required")
	}

	// Mark as reviewed
	if err := s.appRepo.MarkAsReviewed(applicationID, adminID, approve, rejectionReason, adminNotes); err != nil {
		return nil, fmt.Errorf("failed to review application: %w", err)
	}

	// If approved, create cleaner profile
	if approve {
		cleanerID, err := s.createCleanerFromApplication(app)
		if err != nil {
			return nil, fmt.Errorf("failed to create cleaner profile: %w", err)
		}

		// Mark as converted
		if err := s.appRepo.MarkAsConverted(applicationID, cleanerID); err != nil {
			return nil, fmt.Errorf("failed to mark as converted: %w", err)
		}
	}

	// Refresh application
	app, err = s.appRepo.GetByID(applicationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated application: %w", err)
	}

	// TODO: Send notification to applicant

	return app, nil
}

// createCleanerFromApplication converts approved application to cleaner profile
func (s *CleanerApplicationService) createCleanerFromApplication(app *models.CleanerApplication) (string, error) {
	if app.UserID == nil {
		return "", fmt.Errorf("application must have user_id")
	}

	appData := app.ApplicationData

	// Parse years of experience from string to int
	yearsOfExperience := 0
	if appData.Eligibility != nil {
		switch appData.Eligibility.Experience {
		case "none":
			yearsOfExperience = 0
		case "1-2":
			yearsOfExperience = 1
		case "3-5":
			yearsOfExperience = 3
		case "5+":
			yearsOfExperience = 5
		}
	}

	// Prepare languages JSONB
	var languagesJSON []byte
	var err error
	if appData.Profile != nil && len(appData.Profile.Languages) > 0 {
		languagesJSON, err = json.Marshal(appData.Profile.Languages)
		if err != nil {
			return "", fmt.Errorf("failed to marshal languages: %w", err)
		}
	} else {
		languagesJSON = []byte("[]")
	}

	// Prepare equipment as specializations (map equipment to cleaning specializations)
	var specializationsJSON []byte
	if appData.Profile != nil && len(appData.Profile.Equipment) > 0 {
		specializationsJSON, err = json.Marshal(appData.Profile.Equipment)
		if err != nil {
			return "", fmt.Errorf("failed to marshal equipment: %w", err)
		}
	} else {
		specializationsJSON = []byte("[]")
	}

	// Create cleaner record
	cleaner := &models.Cleaner{
		UserID:            app.UserID.String(),
		PhoneNumber:       "", // Will be extracted from user profile
		YearsOfExperience: yearsOfExperience,
		Languages:         languagesJSON,
		Specializations:   specializationsJSON,
		ApprovalStatus:    models.ApprovalStatusApproved,
		IsActive:          true,
		IsAvailable:       true,
		TotalJobs:         0,
		TotalEarnings:     0,
	}

	// Set bio if available
	if appData.Profile != nil && appData.Profile.Bio != "" {
		cleaner.Bio = sql.NullString{String: appData.Profile.Bio, Valid: true}
	}

	// Set profile photo if available
	if appData.Profile != nil && appData.Profile.PhotoURL != nil && *appData.Profile.PhotoURL != "" {
		cleaner.ProfilePhotoURL = sql.NullString{String: *appData.Profile.PhotoURL, Valid: true}
	}

	// Set IBAN if available (encrypted)
	if appData.Legal != nil && appData.Legal.IBAN != "" {
		// TODO: Implement encryption for IBAN
		// For now, store as plaintext (will be encrypted in future)
		cleaner.IBAN = sql.NullString{String: appData.Legal.IBAN, Valid: true}
	}

	// Set document URLs if available
	if appData.Documents != nil {
		if appData.Documents.IDFrontURL != nil && *appData.Documents.IDFrontURL != "" {
			cleaner.IDDocumentURL = sql.NullString{String: *appData.Documents.IDFrontURL, Valid: true}
			cleaner.IDDocumentVerified = true // Auto-verify since admin approved
		}
		if appData.Documents.CazierURL != nil && *appData.Documents.CazierURL != "" {
			cleaner.BackgroundCheckURL = sql.NullString{String: *appData.Documents.CazierURL, Valid: true}
			cleaner.BackgroundCheckVerified = true // Auto-verify since admin approved
		}
	}

	// Create cleaner in database
	cleanerRepo := models.NewCleanerRepository(s.db)
	if err := cleanerRepo.Create(cleaner); err != nil {
		return "", fmt.Errorf("failed to create cleaner record: %w", err)
	}

	// TODO: Set up initial availability based on application data
	// TODO: Send welcome email to cleaner

	return cleaner.ID, nil
}

// CalculateEarnings estimates earning potential for cleaner
func (s *CleanerApplicationService) CalculateEarnings(hoursPerWeek string, areas []string) (*models.EarningPotential, error) {
	// Parse hours range
	var minHours, maxHours float64
	switch hoursPerWeek {
	case "5-10":
		minHours, maxHours = 5, 10
	case "10-20":
		minHours, maxHours = 10, 20
	case "20-30":
		minHours, maxHours = 20, 30
	case "30+":
		minHours, maxHours = 30, 40
	default:
		return nil, fmt.Errorf("invalid hours per week range")
	}

	baseRate := 50.0 // RON/hour

	// Area multiplier (premium areas pay more)
	areaMultiplier := 1.0
	for _, area := range areas {
		if area == "sector_1" || area == "sector_2" {
			areaMultiplier = 1.1
			break
		}
	}

	adjustedRate := baseRate * areaMultiplier

	potential := &models.EarningPotential{
		WeeklyMin:         minHours * adjustedRate,
		WeeklyMax:         maxHours * adjustedRate,
		MonthlyMin:        minHours * adjustedRate * 4.33, // Average weeks per month
		MonthlyMax:        maxHours * adjustedRate * 4.33,
		BaseRate:          baseRate,
		TopCleanerMonthly: 8000.0, // Aspirational goal for top performers
	}

	return potential, nil
}
