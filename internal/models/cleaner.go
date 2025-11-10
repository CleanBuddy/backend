package models

import (
	"database/sql"
	"fmt"
	"time"
)

// ApprovalStatus represents cleaner approval status
type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "PENDING"
	ApprovalStatusApproved ApprovalStatus = "APPROVED"
	ApprovalStatusRejected ApprovalStatus = "REJECTED"
)

// Cleaner represents a cleaner profile
type Cleaner struct {
	ID     string
	UserID string

	// Personal Information
	PhoneNumber  string
	DateOfBirth  sql.NullTime
	StreetAddress sql.NullString
	City          sql.NullString
	County        sql.NullString
	PostalCode    sql.NullString
	Latitude      sql.NullFloat64 // Decimal degrees (-90 to 90)
	Longitude     sql.NullFloat64 // Decimal degrees (-180 to 180)

	// Experience & Skills
	YearsOfExperience int
	Bio               sql.NullString
	Specializations   []byte // JSONB
	Languages         []byte // JSONB

	// Financial Information
	IBAN          sql.NullString // DEPRECATED: Plaintext IBAN (use EncryptedIBAN instead)
	EncryptedIBAN sql.NullString // AES-256-GCM encrypted IBAN (base64-encoded)

	// KYC Documents
	IDDocumentURL           sql.NullString
	IDDocumentVerified      bool
	BackgroundCheckURL      sql.NullString
	BackgroundCheckVerified bool
	ProfilePhotoURL         sql.NullString

	// Ratings & Stats
	AverageRating  sql.NullFloat64
	TotalJobs      int
	TotalEarnings  float64

	// Status
	ApprovalStatus ApprovalStatus
	IsActive       bool
	IsAvailable    bool

	// Admin fields
	AdminNotes     sql.NullString
	ApprovedBy     sql.NullString
	ApprovedAt     sql.NullTime
	RejectedReason sql.NullString

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

// CleanerRepository handles cleaner database operations
type CleanerRepository struct {
	db *sql.DB
}

// NewCleanerRepository creates a new cleaner repository
func NewCleanerRepository(db *sql.DB) *CleanerRepository {
	return &CleanerRepository{db: db}
}

// GetByUserID finds a cleaner by user ID
func (r *CleanerRepository) GetByUserID(userID string) (*Cleaner, error) {
	cleaner := &Cleaner{}
	err := r.db.QueryRow(`
		SELECT id, user_id, phone_number, date_of_birth, street_address, city, county, postal_code,
		       latitude, longitude,
		       years_of_experience, bio, specializations, languages, iban,
		       id_document_url, id_document_verified, background_check_url, background_check_verified, profile_photo_url,
		       average_rating, total_jobs, total_earnings,
		       approval_status, is_active, is_available,
		       admin_notes, approved_by, approved_at, rejected_reason,
		       created_at, updated_at
		FROM cleaners
		WHERE user_id = $1
	`, userID).Scan(
		&cleaner.ID, &cleaner.UserID, &cleaner.PhoneNumber, &cleaner.DateOfBirth,
		&cleaner.StreetAddress, &cleaner.City, &cleaner.County, &cleaner.PostalCode,
		&cleaner.Latitude, &cleaner.Longitude,
		&cleaner.YearsOfExperience, &cleaner.Bio, &cleaner.Specializations, &cleaner.Languages, &cleaner.IBAN,
		&cleaner.IDDocumentURL, &cleaner.IDDocumentVerified,
		&cleaner.BackgroundCheckURL, &cleaner.BackgroundCheckVerified, &cleaner.ProfilePhotoURL,
		&cleaner.AverageRating, &cleaner.TotalJobs, &cleaner.TotalEarnings,
		&cleaner.ApprovalStatus, &cleaner.IsActive, &cleaner.IsAvailable,
		&cleaner.AdminNotes, &cleaner.ApprovedBy, &cleaner.ApprovedAt, &cleaner.RejectedReason,
		&cleaner.CreatedAt, &cleaner.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return cleaner, nil
}

// GetByID finds a cleaner by ID
func (r *CleanerRepository) GetByID(id string) (*Cleaner, error) {
	cleaner := &Cleaner{}
	err := r.db.QueryRow(`
		SELECT id, user_id, phone_number, date_of_birth, street_address, city, county, postal_code,
		       latitude, longitude,
		       years_of_experience, bio, specializations, languages, iban,
		       id_document_url, id_document_verified, background_check_url, background_check_verified, profile_photo_url,
		       average_rating, total_jobs, total_earnings,
		       approval_status, is_active, is_available,
		       admin_notes, approved_by, approved_at, rejected_reason,
		       created_at, updated_at
		FROM cleaners
		WHERE id = $1
	`, id).Scan(
		&cleaner.ID, &cleaner.UserID, &cleaner.PhoneNumber, &cleaner.DateOfBirth,
		&cleaner.StreetAddress, &cleaner.City, &cleaner.County, &cleaner.PostalCode,
		&cleaner.Latitude, &cleaner.Longitude,
		&cleaner.YearsOfExperience, &cleaner.Bio, &cleaner.Specializations, &cleaner.Languages, &cleaner.IBAN,
		&cleaner.IDDocumentURL, &cleaner.IDDocumentVerified,
		&cleaner.BackgroundCheckURL, &cleaner.BackgroundCheckVerified, &cleaner.ProfilePhotoURL,
		&cleaner.AverageRating, &cleaner.TotalJobs, &cleaner.TotalEarnings,
		&cleaner.ApprovalStatus, &cleaner.IsActive, &cleaner.IsAvailable,
		&cleaner.AdminNotes, &cleaner.ApprovedBy, &cleaner.ApprovedAt, &cleaner.RejectedReason,
		&cleaner.CreatedAt, &cleaner.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return cleaner, nil
}

// Create creates a new cleaner profile
func (r *CleanerRepository) Create(cleaner *Cleaner) error {
	return r.db.QueryRow(`
		INSERT INTO cleaners (
			user_id, phone_number, date_of_birth, street_address, city, county, postal_code,
			latitude, longitude,
			years_of_experience, bio, specializations, languages,
			approval_status, is_active, is_available
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING id, created_at, updated_at
	`, cleaner.UserID, cleaner.PhoneNumber, cleaner.DateOfBirth,
		cleaner.StreetAddress, cleaner.City, cleaner.County, cleaner.PostalCode,
		cleaner.Latitude, cleaner.Longitude,
		cleaner.YearsOfExperience, cleaner.Bio, cleaner.Specializations, cleaner.Languages,
		cleaner.ApprovalStatus, cleaner.IsActive, cleaner.IsAvailable).
		Scan(&cleaner.ID, &cleaner.CreatedAt, &cleaner.UpdatedAt)
}

// Update updates a cleaner profile
func (r *CleanerRepository) Update(cleaner *Cleaner) error {
	_, err := r.db.Exec(`
		UPDATE cleaners
		SET phone_number = $2, date_of_birth = $3, street_address = $4, city = $5, county = $6, postal_code = $7,
		    latitude = $8, longitude = $9,
		    years_of_experience = $10, bio = $11, specializations = $12, languages = $13, iban = $14,
		    id_document_url = $15, id_document_verified = $16,
		    background_check_url = $17, background_check_verified = $18,
		    profile_photo_url = $19,
		    average_rating = $20, total_jobs = $21, total_earnings = $22,
		    approval_status = $23, is_active = $24, is_available = $25,
		    admin_notes = $26, approved_by = $27, approved_at = $28
		WHERE id = $1
	`, cleaner.ID, cleaner.PhoneNumber, cleaner.DateOfBirth,
		cleaner.StreetAddress, cleaner.City, cleaner.County, cleaner.PostalCode,
		cleaner.Latitude, cleaner.Longitude,
		cleaner.YearsOfExperience, cleaner.Bio, cleaner.Specializations, cleaner.Languages, cleaner.IBAN,
		cleaner.IDDocumentURL, cleaner.IDDocumentVerified,
		cleaner.BackgroundCheckURL, cleaner.BackgroundCheckVerified,
		cleaner.ProfilePhotoURL,
		cleaner.AverageRating, cleaner.TotalJobs, cleaner.TotalEarnings,
		cleaner.ApprovalStatus, cleaner.IsActive, cleaner.IsAvailable,
		cleaner.AdminNotes, cleaner.ApprovedBy, cleaner.ApprovedAt)
	return err
}

// ListPending returns all cleaners pending approval
func (r *CleanerRepository) ListPending() ([]*Cleaner, error) {
	rows, err := r.db.Query(`
		SELECT id, user_id, phone_number, date_of_birth, street_address, city, county, postal_code,
		       latitude, longitude,
		       years_of_experience, bio, specializations, languages, iban,
		       id_document_url, id_document_verified, background_check_url, background_check_verified, profile_photo_url,
		       average_rating, total_jobs, total_earnings,
		       approval_status, is_active, is_available,
		       admin_notes, approved_by, approved_at,
		       created_at, updated_at
		FROM cleaners
		WHERE approval_status = 'PENDING'
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cleaners := []*Cleaner{}
	for rows.Next() {
		cleaner := &Cleaner{}
		err := rows.Scan(
			&cleaner.ID, &cleaner.UserID, &cleaner.PhoneNumber, &cleaner.DateOfBirth,
			&cleaner.StreetAddress, &cleaner.City, &cleaner.County, &cleaner.PostalCode,
			&cleaner.Latitude, &cleaner.Longitude,
			&cleaner.YearsOfExperience, &cleaner.Bio, &cleaner.Specializations, &cleaner.Languages, &cleaner.IBAN,
			&cleaner.IDDocumentURL, &cleaner.IDDocumentVerified,
			&cleaner.BackgroundCheckURL, &cleaner.BackgroundCheckVerified, &cleaner.ProfilePhotoURL,
			&cleaner.AverageRating, &cleaner.TotalJobs, &cleaner.TotalEarnings,
			&cleaner.ApprovalStatus, &cleaner.IsActive, &cleaner.IsAvailable,
			&cleaner.AdminNotes, &cleaner.ApprovedBy, &cleaner.ApprovedAt,
			&cleaner.CreatedAt, &cleaner.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		cleaners = append(cleaners, cleaner)
	}

	return cleaners, rows.Err()
}

// GetByApprovalStatus retrieves all cleaners with a specific approval status
func (r *CleanerRepository) GetByApprovalStatus(status ApprovalStatus) ([]*Cleaner, error) {
	query := `
		SELECT id, user_id, phone_number, date_of_birth, street_address,
			   city, county, postal_code, latitude, longitude,
			   years_of_experience, bio,
			   specializations, languages, iban, id_document_url, id_document_verified,
			   background_check_url, background_check_verified, profile_photo_url,
			   average_rating, total_jobs, total_earnings,
			   approval_status, is_active, is_available,
			   admin_notes, approved_by, approved_at, rejected_reason,
			   created_at, updated_at
		FROM cleaners
		WHERE approval_status = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaners by approval status: %w", err)
	}
	defer rows.Close()

	var cleaners []*Cleaner
	for rows.Next() {
		cleaner := &Cleaner{}
		err := rows.Scan(
			&cleaner.ID,
			&cleaner.UserID,
			&cleaner.PhoneNumber,
			&cleaner.DateOfBirth,
			&cleaner.StreetAddress,
			&cleaner.City,
			&cleaner.County,
			&cleaner.PostalCode,
			&cleaner.Latitude,
			&cleaner.Longitude,
			&cleaner.YearsOfExperience,
			&cleaner.Bio,
			&cleaner.Specializations,
			&cleaner.Languages,
			&cleaner.IBAN,
			&cleaner.IDDocumentURL,
			&cleaner.IDDocumentVerified,
			&cleaner.BackgroundCheckURL,
			&cleaner.BackgroundCheckVerified,
			&cleaner.ProfilePhotoURL,
			&cleaner.AverageRating,
			&cleaner.TotalJobs,
			&cleaner.TotalEarnings,
			&cleaner.ApprovalStatus,
			&cleaner.IsActive,
			&cleaner.IsAvailable,
			&cleaner.AdminNotes,
			&cleaner.ApprovedBy,
			&cleaner.ApprovedAt,
			&cleaner.RejectedReason,
			&cleaner.CreatedAt,
			&cleaner.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cleaner: %w", err)
		}
		cleaners = append(cleaners, cleaner)
	}

	return cleaners, rows.Err()
}

// GetApprovedCleaners gets approved cleaners with optional search and pagination
func (r *CleanerRepository) GetApprovedCleaners(limit, offset int, search string) ([]*Cleaner, error) {
	query := `
		SELECT c.id, c.user_id, c.phone_number, c.date_of_birth, c.street_address,
			   c.city, c.county, c.postal_code, c.years_of_experience, c.bio,
			   c.specializations, c.languages, c.iban, c.id_document_url, c.id_document_verified,
			   c.background_check_url, c.background_check_verified, c.profile_photo_url,
			   c.average_rating, c.total_jobs, c.total_earnings,
			   c.approval_status, c.is_active, c.is_available,
			   c.admin_notes, c.approved_by, c.approved_at, c.rejected_reason,
			   c.created_at, c.updated_at
		FROM cleaners c
		LEFT JOIN users u ON c.user_id = u.id
		WHERE c.approval_status = 'APPROVED'
	`

	args := []interface{}{}
	argIndex := 1

	if search != "" {
		query += ` AND (
			u.first_name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR
			u.last_name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR
			c.city ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR
			c.id ILIKE $` + fmt.Sprintf("%d", argIndex) + `
		)`
		args = append(args, "%"+search+"%")
		argIndex++
	}

	query += ` ORDER BY c.average_rating DESC NULLS LAST, c.total_jobs DESC`

	if limit > 0 {
		query += ` LIMIT $` + fmt.Sprintf("%d", argIndex)
		args = append(args, limit)
		argIndex++
	}

	if offset > 0 {
		query += ` OFFSET $` + fmt.Sprintf("%d", argIndex)
		args = append(args, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get approved cleaners: %w", err)
	}
	defer rows.Close()

	var cleaners []*Cleaner
	for rows.Next() {
		cleaner := &Cleaner{}
		err := rows.Scan(
			&cleaner.ID,
			&cleaner.UserID,
			&cleaner.PhoneNumber,
			&cleaner.DateOfBirth,
			&cleaner.StreetAddress,
			&cleaner.City,
			&cleaner.County,
			&cleaner.PostalCode,
			&cleaner.YearsOfExperience,
			&cleaner.Bio,
			&cleaner.Specializations,
			&cleaner.Languages,
			&cleaner.IBAN,
			&cleaner.IDDocumentURL,
			&cleaner.IDDocumentVerified,
			&cleaner.BackgroundCheckURL,
			&cleaner.BackgroundCheckVerified,
			&cleaner.ProfilePhotoURL,
			&cleaner.AverageRating,
			&cleaner.TotalJobs,
			&cleaner.TotalEarnings,
			&cleaner.ApprovalStatus,
			&cleaner.IsActive,
			&cleaner.IsAvailable,
			&cleaner.AdminNotes,
			&cleaner.ApprovedBy,
			&cleaner.ApprovedAt,
			&cleaner.RejectedReason,
			&cleaner.CreatedAt,
			&cleaner.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cleaner: %w", err)
		}
		cleaners = append(cleaners, cleaner)
	}

	return cleaners, rows.Err()
}

// GetCleaners gets cleaners with optional status filter, search, and pagination
func (r *CleanerRepository) GetCleaners(limit, offset int, status *string, search *string) ([]*Cleaner, error) {
	query := `
		SELECT c.id, c.user_id, c.phone_number, c.date_of_birth, c.street_address,
			   c.city, c.county, c.postal_code, c.years_of_experience, c.bio,
			   c.specializations, c.languages, c.iban, c.id_document_url, c.id_document_verified,
			   c.background_check_url, c.background_check_verified, c.profile_photo_url,
			   c.average_rating, c.total_jobs, c.total_earnings,
			   c.approval_status, c.is_active, c.is_available,
			   c.admin_notes, c.approved_by, c.approved_at, c.rejected_reason,
			   c.created_at, c.updated_at
		FROM cleaners c
		LEFT JOIN users u ON c.user_id = u.id
		WHERE 1=1
	`

	args := []interface{}{}
	argIndex := 1

	// Filter by status if provided
	if status != nil && *status != "" {
		query += ` AND c.approval_status = $` + fmt.Sprintf("%d", argIndex)
		args = append(args, *status)
		argIndex++
	}

	// Filter by search if provided
	if search != nil && *search != "" {
		query += ` AND (
			u.first_name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR
			u.last_name ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR
			c.city ILIKE $` + fmt.Sprintf("%d", argIndex) + ` OR
			c.id ILIKE $` + fmt.Sprintf("%d", argIndex) + `
		)`
		args = append(args, "%"+*search+"%")
		argIndex++
	}

	query += ` ORDER BY c.average_rating DESC NULLS LAST, c.total_jobs DESC`

	if limit > 0 {
		query += ` LIMIT $` + fmt.Sprintf("%d", argIndex)
		args = append(args, limit)
		argIndex++
	}

	if offset > 0 {
		query += ` OFFSET $` + fmt.Sprintf("%d", argIndex)
		args = append(args, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaners: %w", err)
	}
	defer rows.Close()

	var cleaners []*Cleaner
	for rows.Next() {
		cleaner := &Cleaner{}
		err := rows.Scan(
			&cleaner.ID,
			&cleaner.UserID,
			&cleaner.PhoneNumber,
			&cleaner.DateOfBirth,
			&cleaner.StreetAddress,
			&cleaner.City,
			&cleaner.County,
			&cleaner.PostalCode,
			&cleaner.YearsOfExperience,
			&cleaner.Bio,
			&cleaner.Specializations,
			&cleaner.Languages,
			&cleaner.IBAN,
			&cleaner.IDDocumentURL,
			&cleaner.IDDocumentVerified,
			&cleaner.BackgroundCheckURL,
			&cleaner.BackgroundCheckVerified,
			&cleaner.ProfilePhotoURL,
			&cleaner.AverageRating,
			&cleaner.TotalJobs,
			&cleaner.TotalEarnings,
			&cleaner.ApprovalStatus,
			&cleaner.IsActive,
			&cleaner.IsAvailable,
			&cleaner.AdminNotes,
			&cleaner.ApprovedBy,
			&cleaner.ApprovedAt,
			&cleaner.RejectedReason,
			&cleaner.CreatedAt,
			&cleaner.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan cleaner: %w", err)
		}
		cleaners = append(cleaners, cleaner)
	}

	return cleaners, rows.Err()
}

// CleanerStats holds statistics for a cleaner
type CleanerStats struct {
	TotalBookings      int
	CompletedBookings  int
	CancelledBookings  int
	NoShowCount        int
	AverageRating      sql.NullFloat64
	TotalEarnings      float64
	CompletionRate     float64
	ResponseTime       sql.NullFloat64 // Average response time in hours
	LastActiveDate     sql.NullTime
}

// GetCleanerStats retrieves statistics for a specific cleaner
func (r *CleanerRepository) GetCleanerStats(cleanerID string) (*CleanerStats, error) {
	stats := &CleanerStats{}

	// Query to get booking statistics and cleaner info in one go
	query := `
		SELECT
			COALESCE(COUNT(b.id), 0) as total_bookings,
			COALESCE(COUNT(CASE WHEN b.status = 'COMPLETED' THEN 1 END), 0) as completed_bookings,
			COALESCE(COUNT(CASE WHEN b.status IN ('CLIENT_CANCELED', 'CLEANER_CANCELED') THEN 1 END), 0) as cancelled_bookings,
			COALESCE(COUNT(CASE WHEN b.status = 'NO_SHOW' THEN 1 END), 0) as no_show_count,
			c.average_rating,
			c.total_earnings,
			CASE
				WHEN COUNT(b.id) > 0 THEN
					(COUNT(CASE WHEN b.status = 'COMPLETED' THEN 1 END)::float / COUNT(b.id)::float) * 100
				ELSE 0
			END as completion_rate,
			MAX(b.updated_at) as last_active_date
		FROM cleaners c
		LEFT JOIN bookings b ON c.id = b.cleaner_id
		WHERE c.id = $1
		GROUP BY c.id, c.average_rating, c.total_earnings
	`

	err := r.db.QueryRow(query, cleanerID).Scan(
		&stats.TotalBookings,
		&stats.CompletedBookings,
		&stats.CancelledBookings,
		&stats.NoShowCount,
		&stats.AverageRating,
		&stats.TotalEarnings,
		&stats.CompletionRate,
		&stats.LastActiveDate,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("cleaner not found")
		}
		return nil, fmt.Errorf("failed to get cleaner stats: %w", err)
	}

	return stats, nil
}
