package services

import (
	"database/sql"

	"github.com/cleanbuddy/backend/internal/utils"
)

// CleanerEncryptionWrapper wraps CleanerService to transparently encrypt/decrypt IBAN
type CleanerEncryptionWrapper struct {
	*CleanerService
	encryption *utils.EncryptionService
}

// NewCleanerEncryptionWrapper creates a wrapper with encryption support
func NewCleanerEncryptionWrapper(cleanerService *CleanerService, encryption *utils.EncryptionService) *CleanerEncryptionWrapper {
	return &CleanerEncryptionWrapper{
		CleanerService: cleanerService,
		encryption:     encryption,
	}
}

// SetIBAN encrypts and stores IBAN for a cleaner
func (w *CleanerEncryptionWrapper) SetIBAN(cleanerID string, iban string) error {
	// Validate IBAN format (Romanian IBAN: RO + 2 digits + up to 24 alphanumeric)
	if !isValidRomanianIBAN(iban) {
		return sql.ErrNoRows // Return error for invalid IBAN
	}

	// Encrypt IBAN
	encryptedIBAN, err := w.encryption.Encrypt(iban)
	if err != nil {
		return err
	}

	// Get cleaner
	cleaner, err := w.cleanerRepo.GetByID(cleanerID)
	if err != nil || cleaner == nil {
		return err
	}

	// Store encrypted IBAN
	cleaner.IBAN = sql.NullString{String: encryptedIBAN, Valid: true}

	// Update cleaner
	return w.cleanerRepo.Update(cleaner)
}

// GetIBAN retrieves and decrypts IBAN for a cleaner
func (w *CleanerEncryptionWrapper) GetIBAN(cleanerID string) (string, error) {
	// Get cleaner
	cleaner, err := w.cleanerRepo.GetByID(cleanerID)
	if err != nil || cleaner == nil {
		return "", err
	}

	// Check if IBAN exists
	if !cleaner.IBAN.Valid || cleaner.IBAN.String == "" {
		return "", nil // No IBAN set
	}

	// Try to decrypt
	plaintext, err := w.encryption.Decrypt(cleaner.IBAN.String)
	if err != nil {
		// If decryption fails, it might be plaintext (backward compatibility)
		// Return the plaintext value for migration period
		return cleaner.IBAN.String, nil
	}

	return plaintext, nil
}

// isValidRomanianIBAN validates Romanian IBAN format
func isValidRomanianIBAN(iban string) bool {
	// Romanian IBAN format: RO + 2 check digits + up to 24 alphanumeric characters
	// Example: RO49AAAA1B31007593840000
	if len(iban) < 24 || len(iban) > 34 {
		return false
	}
	if iban[0:2] != "RO" {
		return false
	}
	// Check if characters 2-3 are digits
	if iban[2] < '0' || iban[2] > '9' || iban[3] < '0' || iban[3] > '9' {
		return false
	}
	return true
}
