package models

import (
	"database/sql"
	"fmt"
	"time"
)

// CompanyApprovalStatus constants
const (
	CompanyApprovalStatusPending  = "PENDING"
	CompanyApprovalStatusApproved = "APPROVED"
	CompanyApprovalStatusRejected = "REJECTED"
)

// Company represents a cleaning company
type Company struct {
	ID                          string
	Name                        string
	CUI                         string
	RegistrationNumber          sql.NullString
	IBAN                        sql.NullString
	BankName                    sql.NullString
	LegalAddress                sql.NullString
	ContactEmail                sql.NullString
	ContactPhone                sql.NullString
	IDDocumentURL               sql.NullString
	RegistrationDocumentURL     sql.NullString
	IDDocumentVerified          bool
	RegistrationDocumentVerified bool
	ApprovalStatus              string
	RejectedReason              sql.NullString
	ApprovedBy                  sql.NullString
	ApprovedAt                  sql.NullTime
	IsActive                    bool
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
}

// CompanyStats represents company statistics
type CompanyStats struct {
	TotalTeamMembers  int
	ActiveTeamMembers int
	TotalBookings     int
	ActiveBookings    int
	CompletedBookings int
	TotalRevenue      float64
	MonthlyRevenue    float64
	AverageRating     *float64
}

// CompanyRepository handles database operations for companies
type CompanyRepository struct {
	db *sql.DB
}

// NewCompanyRepository creates a new company repository
func NewCompanyRepository(db *sql.DB) *CompanyRepository {
	return &CompanyRepository{db: db}
}

// Create creates a new company
func (r *CompanyRepository) Create(company *Company) error {
	query := `
		INSERT INTO companies (
			name, cui, registration_number, iban, bank_name,
			legal_address, contact_email, contact_phone
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, approval_status, id_document_verified, registration_document_verified,
				  is_active, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		company.Name,
		company.CUI,
		company.RegistrationNumber,
		company.IBAN,
		company.BankName,
		company.LegalAddress,
		company.ContactEmail,
		company.ContactPhone,
	).Scan(
		&company.ID,
		&company.ApprovalStatus,
		&company.IDDocumentVerified,
		&company.RegistrationDocumentVerified,
		&company.IsActive,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create company: %w", err)
	}

	return nil
}

// GetByID retrieves a company by ID
func (r *CompanyRepository) GetByID(id string) (*Company, error) {
	query := `
		SELECT id, name, cui, registration_number, iban, bank_name,
			   legal_address, contact_email, contact_phone,
			   id_document_url, registration_document_url,
			   id_document_verified, registration_document_verified,
			   approval_status, rejected_reason, approved_by, approved_at,
			   is_active, created_at, updated_at
		FROM companies
		WHERE id = $1
	`

	company := &Company{}
	err := r.db.QueryRow(query, id).Scan(
		&company.ID,
		&company.Name,
		&company.CUI,
		&company.RegistrationNumber,
		&company.IBAN,
		&company.BankName,
		&company.LegalAddress,
		&company.ContactEmail,
		&company.ContactPhone,
		&company.IDDocumentURL,
		&company.RegistrationDocumentURL,
		&company.IDDocumentVerified,
		&company.RegistrationDocumentVerified,
		&company.ApprovalStatus,
		&company.RejectedReason,
		&company.ApprovedBy,
		&company.ApprovedAt,
		&company.IsActive,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	return company, nil
}

// GetByCUI retrieves a company by CUI
func (r *CompanyRepository) GetByCUI(cui string) (*Company, error) {
	query := `
		SELECT id, name, cui, registration_number, iban, bank_name,
			   legal_address, contact_email, contact_phone,
			   id_document_url, registration_document_url,
			   id_document_verified, registration_document_verified,
			   approval_status, rejected_reason, approved_by, approved_at,
			   is_active, created_at, updated_at
		FROM companies
		WHERE cui = $1
	`

	company := &Company{}
	err := r.db.QueryRow(query, cui).Scan(
		&company.ID,
		&company.Name,
		&company.CUI,
		&company.RegistrationNumber,
		&company.IBAN,
		&company.BankName,
		&company.LegalAddress,
		&company.ContactEmail,
		&company.ContactPhone,
		&company.IDDocumentURL,
		&company.RegistrationDocumentURL,
		&company.IDDocumentVerified,
		&company.RegistrationDocumentVerified,
		&company.ApprovalStatus,
		&company.RejectedReason,
		&company.ApprovedBy,
		&company.ApprovedAt,
		&company.IsActive,
		&company.CreatedAt,
		&company.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get company by CUI: %w", err)
	}

	return company, nil
}

// Update updates a company
func (r *CompanyRepository) Update(company *Company) error {
	query := `
		UPDATE companies
		SET name = $1, iban = $2, bank_name = $3, legal_address = $4,
			contact_email = $5, contact_phone = $6, updated_at = CURRENT_TIMESTAMP
		WHERE id = $7
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		company.Name,
		company.IBAN,
		company.BankName,
		company.LegalAddress,
		company.ContactEmail,
		company.ContactPhone,
		company.ID,
	).Scan(&company.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("company not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update company: %w", err)
	}

	return nil
}

// UpdateDocumentURL updates document URL and verification status
func (r *CompanyRepository) UpdateDocumentURL(companyID, documentType, documentURL string) error {
	var query string
	if documentType == "ID" {
		query = `
			UPDATE companies
			SET id_document_url = $1, id_document_verified = false, updated_at = CURRENT_TIMESTAMP
			WHERE id = $2
		`
	} else if documentType == "REGISTRATION" {
		query = `
			UPDATE companies
			SET registration_document_url = $1, registration_document_verified = false, updated_at = CURRENT_TIMESTAMP
			WHERE id = $2
		`
	} else {
		return fmt.Errorf("invalid document type: %s", documentType)
	}

	result, err := r.db.Exec(query, documentURL, companyID)
	if err != nil {
		return fmt.Errorf("failed to update document URL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("company not found")
	}

	return nil
}

// UpdateApprovalStatus updates the approval status of a company
func (r *CompanyRepository) UpdateApprovalStatus(companyID, status, adminID string, reason sql.NullString) error {
	query := `
		UPDATE companies
		SET approval_status = $1, approved_by = $2, approved_at = CURRENT_TIMESTAMP,
			rejected_reason = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4
	`

	result, err := r.db.Exec(query, status, adminID, reason, companyID)
	if err != nil {
		return fmt.Errorf("failed to update approval status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("company not found")
	}

	return nil
}

// GetAll retrieves all companies
func (r *CompanyRepository) GetAll() ([]*Company, error) {
	query := `
		SELECT id, name, cui, registration_number, iban, bank_name,
			   legal_address, contact_email, contact_phone,
			   id_document_url, registration_document_url,
			   id_document_verified, registration_document_verified,
			   approval_status, rejected_reason, approved_by, approved_at,
			   is_active, created_at, updated_at
		FROM companies
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get all companies: %w", err)
	}
	defer rows.Close()

	var companies []*Company
	for rows.Next() {
		company := &Company{}
		err := rows.Scan(
			&company.ID,
			&company.Name,
			&company.CUI,
			&company.RegistrationNumber,
			&company.IBAN,
			&company.BankName,
			&company.LegalAddress,
			&company.ContactEmail,
			&company.ContactPhone,
			&company.IDDocumentURL,
			&company.RegistrationDocumentURL,
			&company.IDDocumentVerified,
			&company.RegistrationDocumentVerified,
			&company.ApprovalStatus,
			&company.RejectedReason,
			&company.ApprovedBy,
			&company.ApprovedAt,
			&company.IsActive,
			&company.CreatedAt,
			&company.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan company: %w", err)
		}
		companies = append(companies, company)
	}

	return companies, nil
}

// GetByApprovalStatus retrieves companies by approval status
func (r *CompanyRepository) GetByApprovalStatus(status string) ([]*Company, error) {
	query := `
		SELECT id, name, cui, registration_number, iban, bank_name,
			   legal_address, contact_email, contact_phone,
			   id_document_url, registration_document_url,
			   id_document_verified, registration_document_verified,
			   approval_status, rejected_reason, approved_by, approved_at,
			   is_active, created_at, updated_at
		FROM companies
		WHERE approval_status = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get companies by approval status: %w", err)
	}
	defer rows.Close()

	var companies []*Company
	for rows.Next() {
		company := &Company{}
		err := rows.Scan(
			&company.ID,
			&company.Name,
			&company.CUI,
			&company.RegistrationNumber,
			&company.IBAN,
			&company.BankName,
			&company.LegalAddress,
			&company.ContactEmail,
			&company.ContactPhone,
			&company.IDDocumentURL,
			&company.RegistrationDocumentURL,
			&company.IDDocumentVerified,
			&company.RegistrationDocumentVerified,
			&company.ApprovalStatus,
			&company.RejectedReason,
			&company.ApprovedBy,
			&company.ApprovedAt,
			&company.IsActive,
			&company.CreatedAt,
			&company.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan company: %w", err)
		}
		companies = append(companies, company)
	}

	return companies, nil
}

// GetCompanies retrieves companies with optional filters (status, search) and pagination
func (r *CompanyRepository) GetCompanies(limit, offset int, status *string, search *string) ([]*Company, error) {
	query := `
		SELECT id, name, cui, registration_number, iban, bank_name, legal_address,
		       contact_email, contact_phone, approval_status, rejected_reason,
		       is_active, created_at, updated_at
		FROM companies
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	// Filter by status if provided
	if status != nil && *status != "" {
		query += fmt.Sprintf(" AND approval_status = $%d", argCount)
		args = append(args, *status)
		argCount++
	}

	// Filter by search if provided
	if search != nil && *search != "" {
		searchPattern := "%" + *search + "%"
		query += fmt.Sprintf(" AND (name ILIKE $%d OR cui ILIKE $%d OR contact_email ILIKE $%d)", argCount, argCount, argCount)
		args = append(args, searchPattern)
		argCount++
	}

	// Add ordering and pagination
	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, limit, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query companies: %w", err)
	}
	defer rows.Close()

	var companies []*Company
	for rows.Next() {
		company := &Company{}
		err := rows.Scan(
			&company.ID,
			&company.Name,
			&company.CUI,
			&company.RegistrationNumber,
			&company.IBAN,
			&company.BankName,
			&company.LegalAddress,
			&company.ContactEmail,
			&company.ContactPhone,
			&company.ApprovalStatus,
			&company.RejectedReason,
			&company.IsActive,
			&company.CreatedAt,
			&company.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan company: %w", err)
		}
		companies = append(companies, company)
	}

	return companies, nil
}

// GetCompanyStats retrieves statistics for a company
func (r *CompanyRepository) GetCompanyStats(companyID string) (*CompanyStats, error) {
	query := `
		WITH team_stats AS (
			SELECT
				COUNT(*) FILTER (WHERE status = 'ACTIVE') as active_members,
				COUNT(*) as total_members
			FROM company_cleaners
			WHERE company_id = $1
		),
		booking_stats AS (
			SELECT
				COUNT(*) as total_bookings,
				COUNT(*) FILTER (WHERE status IN ('PENDING', 'CONFIRMED', 'IN_PROGRESS')) as active_bookings,
				COUNT(*) FILTER (WHERE status = 'COMPLETED') as completed_bookings,
				COALESCE(SUM(total_price) FILTER (WHERE status = 'COMPLETED'), 0) as total_revenue,
				COALESCE(SUM(total_price) FILTER (
					WHERE status = 'COMPLETED'
					AND EXTRACT(MONTH FROM completed_at) = EXTRACT(MONTH FROM CURRENT_DATE)
					AND EXTRACT(YEAR FROM completed_at) = EXTRACT(YEAR FROM CURRENT_DATE)
				), 0) as monthly_revenue
			FROM bookings
			WHERE cleaner_id IN (
				SELECT cleaner_id FROM company_cleaners WHERE company_id = $1
			)
		),
		rating_stats AS (
			SELECT AVG(client_rating)::float as avg_rating
			FROM bookings
			WHERE cleaner_id IN (
				SELECT cleaner_id FROM company_cleaners WHERE company_id = $1
			)
			AND client_rating IS NOT NULL
		)
		SELECT
			COALESCE(t.total_members, 0),
			COALESCE(t.active_members, 0),
			COALESCE(b.total_bookings, 0),
			COALESCE(b.active_bookings, 0),
			COALESCE(b.completed_bookings, 0),
			COALESCE(b.total_revenue, 0),
			COALESCE(b.monthly_revenue, 0),
			r.avg_rating
		FROM team_stats t
		CROSS JOIN booking_stats b
		CROSS JOIN rating_stats r
	`

	stats := &CompanyStats{}
	var avgRating sql.NullFloat64

	err := r.db.QueryRow(query, companyID).Scan(
		&stats.TotalTeamMembers,
		&stats.ActiveTeamMembers,
		&stats.TotalBookings,
		&stats.ActiveBookings,
		&stats.CompletedBookings,
		&stats.TotalRevenue,
		&stats.MonthlyRevenue,
		&avgRating,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get company stats: %w", err)
	}

	if avgRating.Valid {
		stats.AverageRating = &avgRating.Float64
	}

	return stats, nil
}
