package models

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

// ErrInvalidJSONB is returned when JSONB data cannot be parsed
var ErrInvalidJSONB = errors.New("invalid JSONB data")

// CleanerApplication represents a cleaner application from anonymous visitor through approval
type CleanerApplication struct {
	ID                      uuid.UUID                  `json:"id"`
	SessionID               *string                    `json:"session_id,omitempty"`
	UserID                  *uuid.UUID                 `json:"user_id,omitempty"`
	ApplicationData         CleanerApplicationData     `json:"application_data"`
	CurrentStep             int                        `json:"current_step"`
	Status                  ApplicationStatus          `json:"status"`
	ReviewedBy              *uuid.UUID                 `json:"reviewed_by,omitempty"`
	ReviewedAt              *time.Time                 `json:"reviewed_at,omitempty"`
	RejectionReason         *string                    `json:"rejection_reason,omitempty"`
	AdminNotes              *string                    `json:"admin_notes,omitempty"`
	ConvertedToCleanerID    *uuid.UUID                 `json:"converted_to_cleaner_id,omitempty"`
	CreatedAt               time.Time                  `json:"created_at"`
	UpdatedAt               time.Time                  `json:"updated_at"`
	SubmittedAt             *time.Time                 `json:"submitted_at,omitempty"`
}

// ApplicationStatus represents the lifecycle of an application
type ApplicationStatus string

const (
	ApplicationStatusDraft        ApplicationStatus = "draft"
	ApplicationStatusSubmitted    ApplicationStatus = "submitted"
	ApplicationStatusUnderReview  ApplicationStatus = "under_review"
	ApplicationStatusApproved     ApplicationStatus = "approved"
	ApplicationStatusRejected     ApplicationStatus = "rejected"
	ApplicationStatusIncomplete   ApplicationStatus = "incomplete"
)

// CleanerApplicationData contains all application details
type CleanerApplicationData struct {
	Eligibility *EligibilityData  `json:"eligibility,omitempty"`
	Availability *AvailabilityData `json:"availability,omitempty"`
	Profile     *ProfileData       `json:"profile,omitempty"`
	Legal       *LegalData         `json:"legal,omitempty"`
	Documents   *DocumentData      `json:"documents,omitempty"`
}

// EligibilityData from Step 1
type EligibilityData struct {
	Age18Plus  bool   `json:"age_18_plus"`
	WorkRight  string `json:"work_right"` // romanian_citizen, eu_citizen, work_permit
	Experience string `json:"experience"` // professional, some, willing_to_learn
}

// AvailabilityData from Step 2
type AvailabilityData struct {
	HoursPerWeek             string   `json:"hours_per_week"` // 5-10, 10-20, 20-30, 30+
	Areas                    []string `json:"areas"`
	Days                     []string `json:"days"`
	TimeSlots                []string `json:"time_slots"`
	EstimatedMonthlyEarnings *float64 `json:"estimated_monthly_earnings,omitempty"`
}

// ProfileData from Step 4
type ProfileData struct {
	PhotoURL  *string  `json:"photo_url,omitempty"`
	Bio       string   `json:"bio"`
	Languages []string `json:"languages"`
	Equipment []string `json:"equipment"`
}

// LegalData from Step 5
type LegalData struct {
	Status       string  `json:"status"` // pfa, srl, individual
	CIF          *string `json:"cif,omitempty"`
	CNPEncrypted string  `json:"cnp_encrypted"` // Always encrypted
	IBAN         string  `json:"iban"`
	BankName     *string `json:"bank_name,omitempty"`
}

// DocumentData from Step 6
type DocumentData struct {
	CazierURL     *string `json:"cazier_url,omitempty"`
	IDFrontURL    *string `json:"id_front_url,omitempty"`
	IDBackURL     *string `json:"id_back_url,omitempty"`
	InsuranceURL  *string `json:"insurance_url,omitempty"`
}

// Value implements the driver.Valuer interface
func (d CleanerApplicationData) Value() (driver.Value, error) {
	return json.Marshal(d)
}

// Scan implements the sql.Scanner interface
func (d *CleanerApplicationData) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return ErrInvalidJSONB
	}

	return json.Unmarshal(b, d)
}

// CleanerApplicationInput is used for creating/updating applications
type CleanerApplicationInput struct {
	SessionID    *string              `json:"session_id,omitempty"`
	UserID       *uuid.UUID           `json:"user_id,omitempty"`
	CurrentStep  int                  `json:"current_step"`
	Eligibility  *EligibilityInput    `json:"eligibility,omitempty"`
	Availability *AvailabilityInput   `json:"availability,omitempty"`
	Profile      *ProfileInput        `json:"profile,omitempty"`
	Legal        *LegalInput          `json:"legal,omitempty"`
	Documents    *DocumentInput       `json:"documents,omitempty"`
}

// EligibilityInput for Step 1
type EligibilityInput struct {
	Age18Plus  bool   `json:"age_18_plus"`
	WorkRight  string `json:"work_right"`
	Experience string `json:"experience"`
}

// AvailabilityInput for Step 2
type AvailabilityInput struct {
	HoursPerWeek string   `json:"hours_per_week"`
	Areas        []string `json:"areas"`
	Days         []string `json:"days"`
	TimeSlots    []string `json:"time_slots"`
}

// ProfileInput for Step 4
type ProfileInput struct {
	PhotoURL  *string  `json:"photo_url,omitempty"`
	Bio       string   `json:"bio"`
	Languages []string `json:"languages"`
	Equipment []string `json:"equipment"`
}

// LegalInput for Step 5
type LegalInput struct {
	Status       string  `json:"status"`
	CIF          *string `json:"cif,omitempty"`
	CNPEncrypted string  `json:"cnp_encrypted"`
	IBAN         string  `json:"iban"`
	BankName     *string `json:"bank_name,omitempty"`
}

// DocumentInput for Step 6
type DocumentInput struct {
	CazierURL    *string `json:"cazier_url,omitempty"`
	IDFrontURL   *string `json:"id_front_url,omitempty"`
	IDBackURL    *string `json:"id_back_url,omitempty"`
	InsuranceURL *string `json:"insurance_url,omitempty"`
}

// ToApplicationData converts input to data model
func (input *CleanerApplicationInput) ToApplicationData() CleanerApplicationData {
	data := CleanerApplicationData{}

	if input.Eligibility != nil {
		data.Eligibility = &EligibilityData{
			Age18Plus:  input.Eligibility.Age18Plus,
			WorkRight:  input.Eligibility.WorkRight,
			Experience: input.Eligibility.Experience,
		}
	}

	if input.Availability != nil {
		data.Availability = &AvailabilityData{
			HoursPerWeek: input.Availability.HoursPerWeek,
			Areas:        input.Availability.Areas,
			Days:         input.Availability.Days,
			TimeSlots:    input.Availability.TimeSlots,
		}
	}

	if input.Profile != nil {
		data.Profile = &ProfileData{
			PhotoURL:  input.Profile.PhotoURL,
			Bio:       input.Profile.Bio,
			Languages: input.Profile.Languages,
			Equipment: input.Profile.Equipment,
		}
	}

	if input.Legal != nil {
		data.Legal = &LegalData{
			Status:       input.Legal.Status,
			CIF:          input.Legal.CIF,
			CNPEncrypted: input.Legal.CNPEncrypted,
			IBAN:         input.Legal.IBAN,
			BankName:     input.Legal.BankName,
		}
	}

	if input.Documents != nil {
		data.Documents = &DocumentData{
			CazierURL:    input.Documents.CazierURL,
			IDFrontURL:   input.Documents.IDFrontURL,
			IDBackURL:    input.Documents.IDBackURL,
			InsuranceURL: input.Documents.InsuranceURL,
		}
	}

	return data
}

// IsComplete checks if application has all required data
func (a *CleanerApplication) IsComplete() bool {
	data := a.ApplicationData

	if data.Eligibility == nil || !data.Eligibility.Age18Plus {
		return false
	}

	if data.Availability == nil || len(data.Availability.Areas) == 0 {
		return false
	}

	if data.Profile == nil || data.Profile.Bio == "" {
		return false
	}

	if data.Legal == nil || data.Legal.IBAN == "" {
		return false
	}

	// Documents are optional for now (can be uploaded later by admin request)
	// TODO: Make documents mandatory once document upload is implemented
	// if data.Documents == nil || data.Documents.CazierURL == nil || data.Documents.IDFrontURL == nil {
	// 	return false
	// }

	return true
}

// CanSubmit checks if application is ready for submission
func (a *CleanerApplication) CanSubmit() error {
	if a.Status != ApplicationStatusDraft {
		return errors.New("application already submitted")
	}

	if !a.IsComplete() {
		return errors.New("application is incomplete")
	}

	if a.UserID == nil {
		return errors.New("user authentication required")
	}

	return nil
}

// EarningPotential represents estimated earning potential
type EarningPotential struct {
	WeeklyMin         float64 `json:"weekly_min"`
	WeeklyMax         float64 `json:"weekly_max"`
	MonthlyMin        float64 `json:"monthly_min"`
	MonthlyMax        float64 `json:"monthly_max"`
	BaseRate          float64 `json:"base_rate"`
	TopCleanerMonthly float64 `json:"top_cleaner_monthly"`
}

// CleanerApplicationRepository handles cleaner application database operations
type CleanerApplicationRepository struct {
	db *sql.DB
}

// NewCleanerApplicationRepository creates a new cleaner application repository
func NewCleanerApplicationRepository(db *sql.DB) *CleanerApplicationRepository {
	return &CleanerApplicationRepository{db: db}
}

// GetBySessionID retrieves an application by session ID (for anonymous users)
func (r *CleanerApplicationRepository) GetBySessionID(sessionID string) (*CleanerApplication, error) {
	app := &CleanerApplication{}
	var appDataJSON []byte
	var sessionIDStr, userIDStr, reviewedByStr, convertedIDStr sql.NullString

	err := r.db.QueryRow(`
		SELECT id, session_id, user_id, application_data, current_step, status,
		       reviewed_by, reviewed_at, rejection_reason, admin_notes,
		       converted_to_cleaner_id, created_at, updated_at, submitted_at
		FROM cleaner_applications
		WHERE session_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`, sessionID).Scan(
		&app.ID, &sessionIDStr, &userIDStr, &appDataJSON, &app.CurrentStep, &app.Status,
		&reviewedByStr, &app.ReviewedAt, &app.RejectionReason, &app.AdminNotes,
		&convertedIDStr, &app.CreatedAt, &app.UpdatedAt, &app.SubmittedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Convert nullable fields
	if sessionIDStr.Valid {
		app.SessionID = &sessionIDStr.String
	}
	if userIDStr.Valid {
		uid, _ := uuid.Parse(userIDStr.String)
		app.UserID = &uid
	}
	if reviewedByStr.Valid {
		rid, _ := uuid.Parse(reviewedByStr.String)
		app.ReviewedBy = &rid
	}
	if convertedIDStr.Valid {
		cid, _ := uuid.Parse(convertedIDStr.String)
		app.ConvertedToCleanerID = &cid
	}

	// Unmarshal JSONB
	if err := json.Unmarshal(appDataJSON, &app.ApplicationData); err != nil {
		return nil, err
	}

	return app, nil
}

// GetByID retrieves an application by ID
func (r *CleanerApplicationRepository) GetByID(id string) (*CleanerApplication, error) {
	app := &CleanerApplication{}
	var appDataJSON []byte
	var sessionIDStr, userIDStr, reviewedByStr, convertedIDStr sql.NullString

	err := r.db.QueryRow(`
		SELECT id, session_id, user_id, application_data, current_step, status,
		       reviewed_by, reviewed_at, rejection_reason, admin_notes,
		       converted_to_cleaner_id, created_at, updated_at, submitted_at
		FROM cleaner_applications
		WHERE id = $1
	`, id).Scan(
		&app.ID, &sessionIDStr, &userIDStr, &appDataJSON, &app.CurrentStep, &app.Status,
		&reviewedByStr, &app.ReviewedAt, &app.RejectionReason, &app.AdminNotes,
		&convertedIDStr, &app.CreatedAt, &app.UpdatedAt, &app.SubmittedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Convert nullable fields
	if sessionIDStr.Valid {
		app.SessionID = &sessionIDStr.String
	}
	if userIDStr.Valid {
		uid, _ := uuid.Parse(userIDStr.String)
		app.UserID = &uid
	}
	if reviewedByStr.Valid {
		rid, _ := uuid.Parse(reviewedByStr.String)
		app.ReviewedBy = &rid
	}
	if convertedIDStr.Valid {
		cid, _ := uuid.Parse(convertedIDStr.String)
		app.ConvertedToCleanerID = &cid
	}

	// Unmarshal JSONB
	if err := json.Unmarshal(appDataJSON, &app.ApplicationData); err != nil {
		return nil, err
	}

	return app, nil
}

// GetByUserID retrieves application for a user
func (r *CleanerApplicationRepository) GetByUserID(userID string) (*CleanerApplication, error) {
	app := &CleanerApplication{}
	var appDataJSON []byte
	var sessionIDStr, userIDStr, reviewedByStr, convertedIDStr sql.NullString

	err := r.db.QueryRow(`
		SELECT id, session_id, user_id, application_data, current_step, status,
		       reviewed_by, reviewed_at, rejection_reason, admin_notes,
		       converted_to_cleaner_id, created_at, updated_at, submitted_at
		FROM cleaner_applications
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT 1
	`, userID).Scan(
		&app.ID, &sessionIDStr, &userIDStr, &appDataJSON, &app.CurrentStep, &app.Status,
		&reviewedByStr, &app.ReviewedAt, &app.RejectionReason, &app.AdminNotes,
		&convertedIDStr, &app.CreatedAt, &app.UpdatedAt, &app.SubmittedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// Convert nullable fields
	if sessionIDStr.Valid {
		app.SessionID = &sessionIDStr.String
	}
	if userIDStr.Valid {
		uid, _ := uuid.Parse(userIDStr.String)
		app.UserID = &uid
	}
	if reviewedByStr.Valid {
		rid, _ := uuid.Parse(reviewedByStr.String)
		app.ReviewedBy = &rid
	}
	if convertedIDStr.Valid {
		cid, _ := uuid.Parse(convertedIDStr.String)
		app.ConvertedToCleanerID = &cid
	}

	// Unmarshal JSONB
	if err := json.Unmarshal(appDataJSON, &app.ApplicationData); err != nil {
		return nil, err
	}

	return app, nil
}

// GetPendingApplications retrieves all pending applications for admin review
func (r *CleanerApplicationRepository) GetPendingApplications(limit, offset int) ([]*CleanerApplication, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := r.db.Query(`
		SELECT id, session_id, user_id, application_data, current_step, status,
		       reviewed_by, reviewed_at, rejection_reason, admin_notes,
		       converted_to_cleaner_id, created_at, updated_at, submitted_at
		FROM cleaner_applications
		WHERE status IN ('submitted', 'under_review')
		ORDER BY submitted_at ASC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applications []*CleanerApplication
	for rows.Next() {
		app := &CleanerApplication{}
		var appDataJSON []byte
		var sessionIDStr, userIDStr, reviewedByStr, convertedIDStr sql.NullString

		err := rows.Scan(
			&app.ID, &sessionIDStr, &userIDStr, &appDataJSON, &app.CurrentStep, &app.Status,
			&reviewedByStr, &app.ReviewedAt, &app.RejectionReason, &app.AdminNotes,
			&convertedIDStr, &app.CreatedAt, &app.UpdatedAt, &app.SubmittedAt,
		)
		if err != nil {
			return nil, err
		}

		// Convert nullable fields
		if sessionIDStr.Valid {
			app.SessionID = &sessionIDStr.String
		}
		if userIDStr.Valid {
			uid, _ := uuid.Parse(userIDStr.String)
			app.UserID = &uid
		}
		if reviewedByStr.Valid {
			rid, _ := uuid.Parse(reviewedByStr.String)
			app.ReviewedBy = &rid
		}
		if convertedIDStr.Valid {
			cid, _ := uuid.Parse(convertedIDStr.String)
			app.ConvertedToCleanerID = &cid
		}

		// Unmarshal JSONB
		if err := json.Unmarshal(appDataJSON, &app.ApplicationData); err != nil {
			return nil, err
		}

		applications = append(applications, app)
	}

	return applications, nil
}

// Create creates a new cleaner application
func (r *CleanerApplicationRepository) Create(app *CleanerApplication) error {
	appDataJSON, err := json.Marshal(app.ApplicationData)
	if err != nil {
		return err
	}

	var sessionIDStr, userIDStr *string
	if app.SessionID != nil {
		sessionIDStr = app.SessionID
	}
	if app.UserID != nil {
		str := app.UserID.String()
		userIDStr = &str
	}

	return r.db.QueryRow(`
		INSERT INTO cleaner_applications (
			session_id, user_id, application_data, current_step, status
		) VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`, sessionIDStr, userIDStr, appDataJSON, app.CurrentStep, app.Status).Scan(
		&app.ID, &app.CreatedAt, &app.UpdatedAt,
	)
}

// Update updates an existing cleaner application
func (r *CleanerApplicationRepository) Update(app *CleanerApplication) error {
	appDataJSON, err := json.Marshal(app.ApplicationData)
	if err != nil {
		return err
	}

	var sessionIDStr, userIDStr *string
	if app.SessionID != nil {
		sessionIDStr = app.SessionID
	}
	if app.UserID != nil {
		str := app.UserID.String()
		userIDStr = &str
	}

	_, err = r.db.Exec(`
		UPDATE cleaner_applications
		SET application_data = $1,
		    current_step = $2,
		    status = $3,
		    user_id = $4,
		    session_id = $5,
		    updated_at = NOW()
		WHERE id = $6
	`, appDataJSON, app.CurrentStep, app.Status, userIDStr, sessionIDStr, app.ID.String())

	return err
}

// MarkAsSubmitted marks application as submitted
func (r *CleanerApplicationRepository) MarkAsSubmitted(applicationID string) error {
	_, err := r.db.Exec(`
		UPDATE cleaner_applications
		SET status = $1,
		    submitted_at = NOW(),
		    updated_at = NOW()
		WHERE id = $2
	`, ApplicationStatusSubmitted, applicationID)
	return err
}

// MarkAsReviewed marks application as approved/rejected
func (r *CleanerApplicationRepository) MarkAsReviewed(applicationID, reviewerID string, approved bool, rejectionReason, adminNotes string) error {
	status := ApplicationStatusApproved
	if !approved {
		status = ApplicationStatusRejected
	}

	_, err := r.db.Exec(`
		UPDATE cleaner_applications
		SET status = $1,
		    reviewed_by = $2,
		    reviewed_at = NOW(),
		    rejection_reason = $3,
		    admin_notes = $4,
		    updated_at = NOW()
		WHERE id = $5
	`, status, reviewerID, rejectionReason, adminNotes, applicationID)
	return err
}

// MarkAsConverted marks application as converted to cleaner profile
func (r *CleanerApplicationRepository) MarkAsConverted(applicationID, cleanerID string) error {
	_, err := r.db.Exec(`
		UPDATE cleaner_applications
		SET converted_to_cleaner_id = $1,
		    updated_at = NOW()
		WHERE id = $2
	`, cleanerID, applicationID)
	return err
}
