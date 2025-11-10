package services

import (
	"database/sql"
	"fmt"

	"github.com/cleanbuddy/backend/internal/models"
	"github.com/cleanbuddy/backend/internal/utils"
)

// CompanyService handles business logic for companies
type CompanyService struct {
	companyRepo        *models.CompanyRepository
	companyAdminRepo   *models.CompanyAdminRepository
	companyCleanerRepo *models.CompanyCleanerRepository
	cleanerRepo        *models.CleanerRepository
}

// NewCompanyService creates a new company service
func NewCompanyService(db *sql.DB) *CompanyService {
	return &CompanyService{
		companyRepo:        models.NewCompanyRepository(db),
		companyAdminRepo:   models.NewCompanyAdminRepository(db),
		companyCleanerRepo: models.NewCompanyCleanerRepository(db),
		cleanerRepo:        models.NewCleanerRepository(db),
	}
}

// CreateCompany creates a new company and makes the user its owner
func (s *CompanyService) CreateCompany(
	userID string,
	name string,
	cui string,
	registrationNumber *string,
	iban *string,
	bankName *string,
	legalAddress *string,
	contactEmail *string,
	contactPhone *string,
) (*models.Company, error) {
	// Check if CUI already exists
	existing, err := s.companyRepo.GetByCUI(cui)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing CUI: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("company with CUI %s already exists", cui)
	}

	// Create company
	company := &models.Company{
		Name: name,
		CUI:  cui,
	}

	if registrationNumber != nil {
		company.RegistrationNumber = sql.NullString{String: *registrationNumber, Valid: true}
	}
	if iban != nil {
		// Encrypt IBAN before storing
		encryptedIBAN, err := utils.EncryptIBAN(*iban)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt IBAN: %w", err)
		}
		company.IBAN = sql.NullString{String: encryptedIBAN, Valid: true}
	}
	if bankName != nil {
		company.BankName = sql.NullString{String: *bankName, Valid: true}
	}
	if legalAddress != nil {
		company.LegalAddress = sql.NullString{String: *legalAddress, Valid: true}
	}
	if contactEmail != nil {
		company.ContactEmail = sql.NullString{String: *contactEmail, Valid: true}
	}
	if contactPhone != nil {
		company.ContactPhone = sql.NullString{String: *contactPhone, Valid: true}
	}

	err = s.companyRepo.Create(company)
	if err != nil {
		return nil, fmt.Errorf("failed to create company: %w", err)
	}

	// Make user the owner
	companyAdmin := &models.CompanyAdmin{
		CompanyID: company.ID,
		UserID:    userID,
		Role:      models.CompanyAdminRoleOwner,
	}

	err = s.companyAdminRepo.Create(companyAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to create company admin: %w", err)
	}

	return company, nil
}

// UpdateCompany updates company information
func (s *CompanyService) UpdateCompany(
	companyID string,
	userID string,
	name *string,
	iban *string,
	bankName *string,
	legalAddress *string,
	contactEmail *string,
	contactPhone *string,
) (*models.Company, error) {
	// Check ownership
	isAdmin, err := s.companyAdminRepo.IsAdmin(companyID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check admin status: %w", err)
	}
	if !isAdmin {
		return nil, fmt.Errorf("unauthorized: you must be a company admin")
	}

	// Get company
	company, err := s.companyRepo.GetByID(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}

	// Update fields
	if name != nil {
		company.Name = *name
	}
	if iban != nil {
		// Encrypt IBAN before storing
		encryptedIBAN, err := utils.EncryptIBAN(*iban)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt IBAN: %w", err)
		}
		company.IBAN = sql.NullString{String: encryptedIBAN, Valid: true}
	}
	if bankName != nil {
		company.BankName = sql.NullString{String: *bankName, Valid: true}
	}
	if legalAddress != nil {
		company.LegalAddress = sql.NullString{String: *legalAddress, Valid: true}
	}
	if contactEmail != nil {
		company.ContactEmail = sql.NullString{String: *contactEmail, Valid: true}
	}
	if contactPhone != nil {
		company.ContactPhone = sql.NullString{String: *contactPhone, Valid: true}
	}

	// Update in database
	err = s.companyRepo.Update(company)
	if err != nil {
		return nil, fmt.Errorf("failed to update company: %w", err)
	}

	return company, nil
}

// UploadCompanyDocument uploads a company document
func (s *CompanyService) UploadCompanyDocument(companyID, userID, documentType, fileURL string) (*models.Company, error) {
	// Check ownership
	isAdmin, err := s.companyAdminRepo.IsAdmin(companyID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check admin status: %w", err)
	}
	if !isAdmin {
		return nil, fmt.Errorf("unauthorized: you must be a company admin")
	}

	// Update document URL
	err = s.companyRepo.UpdateDocumentURL(companyID, documentType, fileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to upload document: %w", err)
	}

	// Get updated company
	company, err := s.companyRepo.GetByID(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	return company, nil
}

// AddCleanerToCompany adds a cleaner to a company
func (s *CompanyService) AddCleanerToCompany(companyID, cleanerID, userID string) (*models.CompanyCleaner, error) {
	// Check ownership
	isAdmin, err := s.companyAdminRepo.IsAdmin(companyID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check admin status: %w", err)
	}
	if !isAdmin {
		return nil, fmt.Errorf("unauthorized: you must be a company admin")
	}

	// Verify cleaner exists and is approved
	cleaner, err := s.cleanerRepo.GetByID(cleanerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaner: %w", err)
	}
	if cleaner == nil {
		return nil, fmt.Errorf("cleaner not found")
	}
	if cleaner.ApprovalStatus != models.ApprovalStatusApproved {
		return nil, fmt.Errorf("cleaner must be approved before joining a company")
	}

	// Check if already exists
	existing, err := s.companyCleanerRepo.GetByCompanyAndCleaner(companyID, cleanerID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing relationship: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("cleaner is already associated with this company")
	}

	// Create relationship
	companyCleaner := &models.CompanyCleaner{
		CompanyID: companyID,
		CleanerID: cleanerID,
		Status:    models.CompanyCleanerStatusActive,
	}

	err = s.companyCleanerRepo.Create(companyCleaner)
	if err != nil {
		return nil, fmt.Errorf("failed to add cleaner to company: %w", err)
	}

	return companyCleaner, nil
}

// RemoveCleanerFromCompany removes a cleaner from a company
func (s *CompanyService) RemoveCleanerFromCompany(companyID, cleanerID, userID string) error {
	// Check ownership
	isAdmin, err := s.companyAdminRepo.IsAdmin(companyID, userID)
	if err != nil {
		return fmt.Errorf("failed to check admin status: %w", err)
	}
	if !isAdmin {
		return fmt.Errorf("unauthorized: you must be a company admin")
	}

	// Get relationship
	companyCleaner, err := s.companyCleanerRepo.GetByCompanyAndCleaner(companyID, cleanerID)
	if err != nil {
		return fmt.Errorf("failed to get company cleaner: %w", err)
	}
	if companyCleaner == nil {
		return fmt.Errorf("cleaner is not associated with this company")
	}

	// Check for active bookings
	hasBookings, err := s.companyCleanerRepo.HasActiveBookings(cleanerID)
	if err != nil {
		return fmt.Errorf("failed to check active bookings: %w", err)
	}
	if hasBookings {
		return fmt.Errorf("cannot remove cleaner with active bookings")
	}

	// Delete relationship
	err = s.companyCleanerRepo.Delete(companyCleaner.ID)
	if err != nil {
		return fmt.Errorf("failed to remove cleaner from company: %w", err)
	}

	return nil
}

// GetMyCompanies gets all companies a user administers
func (s *CompanyService) GetMyCompanies(userID string) ([]*models.Company, error) {
	// Get company admins for user
	admins, err := s.companyAdminRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user companies: %w", err)
	}

	// Get companies
	var companies []*models.Company
	for _, admin := range admins {
		company, err := s.companyRepo.GetByID(admin.CompanyID)
		if err != nil {
			return nil, fmt.Errorf("failed to get company: %w", err)
		}
		if company != nil {
			companies = append(companies, company)
		}
	}

	return companies, nil
}

// GetUserCompanies gets all companies a user administers (alias for GetMyCompanies)
func (s *CompanyService) GetUserCompanies(userID string) ([]*models.Company, error) {
	return s.GetMyCompanies(userID)
}

// GetCompany gets a company by ID
func (s *CompanyService) GetCompany(companyID string) (*models.Company, error) {
	company, err := s.companyRepo.GetByID(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}

	return company, nil
}

// GetCompanyTeam gets all cleaners for a company
func (s *CompanyService) GetCompanyTeam(companyID, userID string) ([]*models.CompanyCleaner, error) {
	// Check ownership
	isAdmin, err := s.companyAdminRepo.IsAdmin(companyID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check admin status: %w", err)
	}
	if !isAdmin {
		return nil, fmt.Errorf("unauthorized: you must be a company admin")
	}

	// Get cleaners
	cleaners, err := s.companyCleanerRepo.GetByCompanyID(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company team: %w", err)
	}

	return cleaners, nil
}

// ApproveCompany approves a company (admin only)
func (s *CompanyService) ApproveCompany(companyID, adminID string) (*models.Company, error) {
	err := s.companyRepo.UpdateApprovalStatus(companyID, models.CompanyApprovalStatusApproved, adminID, sql.NullString{})
	if err != nil {
		return nil, fmt.Errorf("failed to approve company: %w", err)
	}

	company, err := s.companyRepo.GetByID(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	return company, nil
}

// RejectCompany rejects a company (admin only)
func (s *CompanyService) RejectCompany(companyID, adminID, reason string) (*models.Company, error) {
	err := s.companyRepo.UpdateApprovalStatus(
		companyID,
		models.CompanyApprovalStatusRejected,
		adminID,
		sql.NullString{String: reason, Valid: true},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to reject company: %w", err)
	}

	company, err := s.companyRepo.GetByID(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company: %w", err)
	}

	return company, nil
}

// GetPendingCompanies gets all pending companies (admin only)
func (s *CompanyService) GetPendingCompanies() ([]*models.Company, error) {
	companies, err := s.companyRepo.GetByApprovalStatus(models.CompanyApprovalStatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending companies: %w", err)
	}

	return companies, nil
}

// GetCompanyStats retrieves statistics for a company
func (s *CompanyService) GetCompanyStats(companyID string) (*models.CompanyStats, error) {
	stats, err := s.companyRepo.GetCompanyStats(companyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get company stats: %w", err)
	}

	return stats, nil
}

// GetCompanies returns a list of companies with optional filters
func (s *CompanyService) GetCompanies(limit, offset int, status *string, search *string) ([]*models.Company, error) {
	return s.companyRepo.GetCompanies(limit, offset, status, search)
}
