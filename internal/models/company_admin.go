package models

import (
	"database/sql"
	"fmt"
	"time"
)

// CompanyAdminRole constants
const (
	CompanyAdminRoleOwner = "OWNER"
	CompanyAdminRoleAdmin = "ADMIN"
)

// CompanyAdmin represents a company administrator
type CompanyAdmin struct {
	ID        string
	CompanyID string
	UserID    string
	Role      string
	CreatedAt time.Time
}

// CompanyAdminRepository handles database operations for company admins
type CompanyAdminRepository struct {
	db *sql.DB
}

// NewCompanyAdminRepository creates a new company admin repository
func NewCompanyAdminRepository(db *sql.DB) *CompanyAdminRepository {
	return &CompanyAdminRepository{db: db}
}

// Create creates a new company admin record
func (r *CompanyAdminRepository) Create(companyAdmin *CompanyAdmin) error {
	query := `
		INSERT INTO company_admins (company_id, user_id, role)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	err := r.db.QueryRow(
		query,
		companyAdmin.CompanyID,
		companyAdmin.UserID,
		companyAdmin.Role,
	).Scan(&companyAdmin.ID, &companyAdmin.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create company admin: %w", err)
	}

	return nil
}

// GetByCompanyID retrieves all admins for a company
func (r *CompanyAdminRepository) GetByCompanyID(companyID string) ([]*CompanyAdmin, error) {
	query := `
		SELECT id, company_id, user_id, role, created_at
		FROM company_admins
		WHERE company_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.Query(query, companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company admins: %w", err)
	}
	defer rows.Close()

	var admins []*CompanyAdmin
	for rows.Next() {
		admin := &CompanyAdmin{}
		err := rows.Scan(
			&admin.ID,
			&admin.CompanyID,
			&admin.UserID,
			&admin.Role,
			&admin.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan company admin: %w", err)
		}
		admins = append(admins, admin)
	}

	return admins, nil
}

// GetByUserID retrieves all companies a user administers
func (r *CompanyAdminRepository) GetByUserID(userID string) ([]*CompanyAdmin, error) {
	query := `
		SELECT id, company_id, user_id, role, created_at
		FROM company_admins
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user's companies: %w", err)
	}
	defer rows.Close()

	var admins []*CompanyAdmin
	for rows.Next() {
		admin := &CompanyAdmin{}
		err := rows.Scan(
			&admin.ID,
			&admin.CompanyID,
			&admin.UserID,
			&admin.Role,
			&admin.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan company admin: %w", err)
		}
		admins = append(admins, admin)
	}

	return admins, nil
}

// GetByCompanyAndUser retrieves a specific company admin record
func (r *CompanyAdminRepository) GetByCompanyAndUser(companyID, userID string) (*CompanyAdmin, error) {
	query := `
		SELECT id, company_id, user_id, role, created_at
		FROM company_admins
		WHERE company_id = $1 AND user_id = $2
	`

	admin := &CompanyAdmin{}
	err := r.db.QueryRow(query, companyID, userID).Scan(
		&admin.ID,
		&admin.CompanyID,
		&admin.UserID,
		&admin.Role,
		&admin.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get company admin: %w", err)
	}

	return admin, nil
}

// Delete removes a company admin
func (r *CompanyAdminRepository) Delete(id string) error {
	query := `DELETE FROM company_admins WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete company admin: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("company admin not found")
	}

	return nil
}

// IsAdmin checks if a user is an admin of a company
func (r *CompanyAdminRepository) IsAdmin(companyID, userID string) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM company_admins
		WHERE company_id = $1 AND user_id = $2
	`

	var isAdmin bool
	err := r.db.QueryRow(query, companyID, userID).Scan(&isAdmin)
	if err != nil {
		return false, fmt.Errorf("failed to check if user is admin: %w", err)
	}

	return isAdmin, nil
}
