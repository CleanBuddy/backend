package services

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/cleanbuddy/backend/internal/models"
	"github.com/cleanbuddy/backend/internal/utils"
)

// AddressService handles address business logic
type AddressService struct {
	addressRepo      *models.AddressRepository
	geocodingService *utils.GeocodingService
}

// NewAddressService creates a new address service
func NewAddressService(db *sql.DB) *AddressService {
	return &AddressService{
		addressRepo:      models.NewAddressRepository(db),
		geocodingService: utils.NewGeocodingService(),
	}
}

// GetUserAddresses gets all addresses for a user
func (s *AddressService) GetUserAddresses(userID string) ([]*models.Address, error) {
	addresses, err := s.addressRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses: %w", err)
	}
	return addresses, nil
}

// GetAddress gets a specific address by ID
func (s *AddressService) GetAddress(id string, userID string) (*models.Address, error) {
	address, err := s.addressRepo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}
	if address == nil {
		return nil, fmt.Errorf("address not found")
	}
	// Verify ownership
	if address.UserID != userID {
		return nil, fmt.Errorf("unauthorized")
	}
	return address, nil
}

// CreateAddress creates a new address
func (s *AddressService) CreateAddress(userID, label, streetAddress, apartment, city, county, postalCode, additionalInfo string, isDefault bool) (*models.Address, error) {
	address := &models.Address{
		UserID:        userID,
		Label:         label,
		StreetAddress: streetAddress,
		City:          city,
		County:        county,
		Country:       "RO",
		IsDefault:     isDefault,
	}

	if apartment != "" {
		address.Apartment = sql.NullString{String: apartment, Valid: true}
	}
	if postalCode != "" {
		address.PostalCode = sql.NullString{String: postalCode, Valid: true}
	}
	if additionalInfo != "" {
		address.AdditionalInfo = sql.NullString{String: additionalInfo, Valid: true}
	}

	// Geocode the address to get lat/long
	geocodeResult, err := s.geocodingService.GeocodeAddress(streetAddress, city, county, "România")
	if err != nil {
		// Log error but don't fail the address creation
		log.Printf("Warning: Failed to geocode address: %v", err)
	} else {
		// Store geocoding results
		address.Latitude = sql.NullFloat64{Float64: geocodeResult.Latitude, Valid: true}
		address.Longitude = sql.NullFloat64{Float64: geocodeResult.Longitude, Valid: true}
	}

	if err := s.addressRepo.Create(address); err != nil {
		return nil, fmt.Errorf("failed to create address: %w", err)
	}

	return address, nil
}

// UpdateAddress updates an address
func (s *AddressService) UpdateAddress(id, userID string, label, streetAddress, apartment, city, county, postalCode, additionalInfo *string, isDefault *bool) (*models.Address, error) {
	// Get existing address
	address, err := s.GetAddress(id, userID)
	if err != nil {
		return nil, err
	}

	// Track if address components changed (need re-geocoding)
	addressChanged := false

	// Update fields if provided
	if label != nil {
		address.Label = *label
	}
	if streetAddress != nil {
		address.StreetAddress = *streetAddress
		addressChanged = true
	}
	if apartment != nil {
		if *apartment != "" {
			address.Apartment = sql.NullString{String: *apartment, Valid: true}
		} else {
			address.Apartment = sql.NullString{Valid: false}
		}
	}
	if city != nil {
		address.City = *city
		addressChanged = true
	}
	if county != nil {
		address.County = *county
		addressChanged = true
	}
	if postalCode != nil {
		if *postalCode != "" {
			address.PostalCode = sql.NullString{String: *postalCode, Valid: true}
		} else {
			address.PostalCode = sql.NullString{Valid: false}
		}
	}
	if additionalInfo != nil {
		if *additionalInfo != "" {
			address.AdditionalInfo = sql.NullString{String: *additionalInfo, Valid: true}
		} else {
			address.AdditionalInfo = sql.NullString{Valid: false}
		}
	}
	if isDefault != nil {
		address.IsDefault = *isDefault
	}

	// Re-geocode if address components changed
	if addressChanged {
		geocodeResult, err := s.geocodingService.GeocodeAddress(address.StreetAddress, address.City, address.County, "România")
		if err != nil {
			// Log error but don't fail the update
			log.Printf("Warning: Failed to geocode updated address: %v", err)
		} else {
			// Update geocoding results
			address.Latitude = sql.NullFloat64{Float64: geocodeResult.Latitude, Valid: true}
			address.Longitude = sql.NullFloat64{Float64: geocodeResult.Longitude, Valid: true}
		}
	}

	if err := s.addressRepo.Update(address); err != nil {
		return nil, fmt.Errorf("failed to update address: %w", err)
	}

	return address, nil
}

// DeleteAddress deletes an address
func (s *AddressService) DeleteAddress(id, userID string) error {
	// Verify ownership
	address, err := s.GetAddress(id, userID)
	if err != nil {
		return err
	}
	if address == nil {
		return fmt.Errorf("address not found")
	}

	if err := s.addressRepo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete address: %w", err)
	}

	return nil
}
