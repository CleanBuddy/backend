package models

import (
	"database/sql"
	"fmt"
	"time"
)

// CompanyCleanerStatus constants
const (
	CompanyCleanerStatusActive   = "ACTIVE"
	CompanyCleanerStatusInactive = "INACTIVE"
)

// CompanyCleaner represents a cleaner's association with a company
type CompanyCleaner struct {
	ID        string
	CompanyID string
	CleanerID string
	Status    string
	JoinedAt  time.Time
	LeftAt    sql.NullTime
}

// CompanyCleanerRepository handles database operations for company cleaners
type CompanyCleanerRepository struct {
	db *sql.DB
}

// NewCompanyCleanerRepository creates a new company cleaner repository
func NewCompanyCleanerRepository(db *sql.DB) *CompanyCleanerRepository {
	return &CompanyCleanerRepository{db: db}
}

// Create creates a new company cleaner record
func (r *CompanyCleanerRepository) Create(companyCleaner *CompanyCleaner) error {
	query := `
		INSERT INTO company_cleaners (company_id, cleaner_id, status)
		VALUES ($1, $2, $3)
		RETURNING id, joined_at
	`

	err := r.db.QueryRow(
		query,
		companyCleaner.CompanyID,
		companyCleaner.CleanerID,
		companyCleaner.Status,
	).Scan(&companyCleaner.ID, &companyCleaner.JoinedAt)

	if err != nil {
		return fmt.Errorf("failed to create company cleaner: %w", err)
	}

	return nil
}

// GetByCompanyID retrieves all cleaners for a company
func (r *CompanyCleanerRepository) GetByCompanyID(companyID string) ([]*CompanyCleaner, error) {
	query := `
		SELECT id, company_id, cleaner_id, status, joined_at, left_at
		FROM company_cleaners
		WHERE company_id = $1
		ORDER BY joined_at DESC
	`

	rows, err := r.db.Query(query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company cleaners: %w", err)
	}
	defer rows.Close()

	var cleaners []*CompanyCleaner
	for rows.Next() {
		cleaner := &CompanyCleaner{}
		err := rows.Scan(
			&cleaner.ID,
			&cleaner.CompanyID,
			&cleaner.CleanerID,
			&cleaner.Status,
			&cleaner.JoinedAt,
			&cleaner.LeftAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan company cleaner: %w", err)
		}
		cleaners = append(cleaners, cleaner)
	}

	return cleaners, nil
}

// GetByCleanerID retrieves all companies a cleaner works for
func (r *CompanyCleanerRepository) GetByCleanerID(cleanerID string) ([]*CompanyCleaner, error) {
	query := `
		SELECT id, company_id, cleaner_id, status, joined_at, left_at
		FROM company_cleaners
		WHERE cleaner_id = $1
		ORDER BY joined_at DESC
	`

	rows, err := r.db.Query(query, cleanerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaner's companies: %w", err)
	}
	defer rows.Close()

	var companies []*CompanyCleaner
	for rows.Next() {
		company := &CompanyCleaner{}
		err := rows.Scan(
			&company.ID,
			&company.CompanyID,
			&company.CleanerID,
			&company.Status,
			&company.JoinedAt,
			&company.LeftAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan company cleaner: %w", err)
		}
		companies = append(companies, company)
	}

	return companies, nil
}

// GetByCompanyAndCleaner retrieves a specific company cleaner record
func (r *CompanyCleanerRepository) GetByCompanyAndCleaner(companyID, cleanerID string) (*CompanyCleaner, error) {
	query := `
		SELECT id, company_id, cleaner_id, status, joined_at, left_at
		FROM company_cleaners
		WHERE company_id = $1 AND cleaner_id = $2
	`

	cleaner := &CompanyCleaner{}
	err := r.db.QueryRow(query, companyID, cleanerID).Scan(
		&cleaner.ID,
		&cleaner.CompanyID,
		&cleaner.CleanerID,
		&cleaner.Status,
		&cleaner.JoinedAt,
		&cleaner.LeftAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get company cleaner: %w", err)
	}

	return cleaner, nil
}

// UpdateStatus updates the status of a company cleaner
func (r *CompanyCleanerRepository) UpdateStatus(id, status string) error {
	query := `
		UPDATE company_cleaners
		SET status = $1, left_at = CASE WHEN $1 = 'INACTIVE' THEN CURRENT_TIMESTAMP ELSE NULL END
		WHERE id = $2
	`

	result, err := r.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update company cleaner status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("company cleaner not found")
	}

	return nil
}

// Delete removes a company cleaner
func (r *CompanyCleanerRepository) Delete(id string) error {
	query := `DELETE FROM company_cleaners WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete company cleaner: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("company cleaner not found")
	}

	return nil
}

// HasActiveBookings checks if a cleaner has active bookings through the company
func (r *CompanyCleanerRepository) HasActiveBookings(cleanerID string) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM bookings
		WHERE cleaner_id = $1
		  AND status IN ('PENDING', 'CONFIRMED', 'IN_PROGRESS')
	`

	var hasBookings bool
	err := r.db.QueryRow(query, cleanerID).Scan(&hasBookings)
	if err != nil {
		return false, fmt.Errorf("failed to check active bookings: %w", err)
	}

	return hasBookings, nil
}
