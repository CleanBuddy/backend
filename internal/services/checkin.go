package services

import (
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/cleanbuddy/backend/internal/models"
)

// CheckinService handles check-in/check-out business logic
type CheckinService struct {
	checkinRepo      *models.CheckinRepository
	bookingRepo      *models.BookingRepository
	addressRepo      *models.AddressRepository
	cleanerRepo      *models.CleanerRepository
	bookingService   *BookingService
	geocodingService *GeocodingService
}

// NewCheckinService creates a new checkin service
func NewCheckinService(db *sql.DB, bookingService *BookingService) *CheckinService {
	return &CheckinService{
		checkinRepo:      models.NewCheckinRepository(db),
		bookingRepo:      models.NewBookingRepository(db),
		addressRepo:      models.NewAddressRepository(db),
		cleanerRepo:      models.NewCleanerRepository(db),
		bookingService:   bookingService,
		geocodingService: NewGeocodingService(),
	}
}

// CheckIn creates a check-in record and starts the booking
func (s *CheckinService) CheckIn(bookingID string, cleanerID string, latitude, longitude float64) (*models.Checkin, error) {
	// Get booking
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Get cleaner by user_id to validate authorization
	cleaner, err := s.cleanerRepo.GetByUserID(cleanerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaner: %w", err)
	}
	if cleaner == nil {
		return nil, fmt.Errorf("cleaner not found")
	}

	// Verify cleaner is assigned to this booking
	if !booking.CleanerID.Valid || booking.CleanerID.String != cleaner.ID {
		return nil, fmt.Errorf("unauthorized: cleaner not assigned to this booking")
	}

	// Verify booking is confirmed
	if booking.Status != models.BookingStatusConfirmed {
		return nil, fmt.Errorf("booking must be in CONFIRMED status to check in")
	}

	// Check if already checked in
	existingCheckin, err := s.checkinRepo.GetByBookingID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing checkin: %w", err)
	}
	if existingCheckin != nil && existingCheckin.CheckInTime.Valid {
		return nil, fmt.Errorf("already checked in for this booking")
	}

	// Get address and validate GPS location
	address, err := s.addressRepo.GetByID(booking.AddressID)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}
	if address == nil {
		return nil, fmt.Errorf("address not found")
	}

	// Geocode address if coordinates are not already stored
	if !address.Latitude.Valid || !address.Longitude.Valid {
		postalCode := ""
		if address.PostalCode.Valid {
			postalCode = address.PostalCode.String
		}

		geocodeResult, err := s.geocodingService.Geocode(
			address.StreetAddress,
			address.City,
			address.County,
			postalCode,
			address.Country,
		)
		if err != nil {
			// Log error but don't block check-in (geocoding is best-effort)
			fmt.Printf("Warning: failed to geocode address: %v\n", err)
		} else {
			// Update address with geocoded coordinates for future use
			address.Latitude = sql.NullFloat64{Float64: geocodeResult.Latitude, Valid: true}
			address.Longitude = sql.NullFloat64{Float64: geocodeResult.Longitude, Valid: true}
			if err := s.addressRepo.Update(address); err != nil {
				// Log error but don't block check-in
				fmt.Printf("Warning: failed to update address coordinates: %v\n", err)
			}
		}
	}

	// Validate GPS location if address has coordinates
	if address.Latitude.Valid && address.Longitude.Valid {
		distance := haversineDistance(
			latitude, longitude,
			address.Latitude.Float64, address.Longitude.Float64,
		)

		// Allow check-in within 200m of the address (flexible for GPS accuracy)
		const maxDistanceMeters = 200.0
		if distance > maxDistanceMeters {
			return nil, fmt.Errorf(
				"check-in location is too far from booking address (%.0fm away, max %0.fm allowed)",
				distance, maxDistanceMeters,
			)
		}
	}

	// Create check-in
	now := time.Now()
	checkin := &models.Checkin{
		BookingID:        bookingID,
		CleanerID:        cleaner.ID, // Use cleaner table ID, not user_id
		CheckInTime:      sql.NullTime{Time: now, Valid: true},
		CheckInLatitude:  sql.NullFloat64{Float64: latitude, Valid: true},
		CheckInLongitude: sql.NullFloat64{Float64: longitude, Valid: true},
	}

	if err := s.checkinRepo.Create(checkin); err != nil {
		return nil, fmt.Errorf("failed to create checkin: %w", err)
	}

	// Start the booking (pass user_id for StartBooking)
	if _, err := s.bookingService.StartBooking(bookingID, cleanerID); err != nil {
		return nil, fmt.Errorf("failed to start booking: %w", err)
	}

	return checkin, nil
}

// CheckOut updates check-out time and completes the booking
func (s *CheckinService) CheckOut(bookingID string, cleanerID string, latitude, longitude float64) (*models.Checkin, error) {
	// Get existing check-in
	checkin, err := s.checkinRepo.GetByBookingID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get checkin: %w", err)
	}
	if checkin == nil {
		return nil, fmt.Errorf("no check-in found for this booking")
	}

	// Get cleaner by user_id to validate authorization
	cleaner, err := s.cleanerRepo.GetByUserID(cleanerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaner: %w", err)
	}
	if cleaner == nil {
		return nil, fmt.Errorf("cleaner not found")
	}

	// Verify cleaner matches (checkin has cleaner table ID)
	if checkin.CleanerID != cleaner.ID {
		return nil, fmt.Errorf("unauthorized: cleaner mismatch")
	}

	// Verify not already checked out
	if checkin.CheckOutTime.Valid {
		return nil, fmt.Errorf("already checked out for this booking")
	}

	// Verify checked in
	if !checkin.CheckInTime.Valid {
		return nil, fmt.Errorf("must check in before checking out")
	}

	// Get booking
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Verify booking is in progress
	if booking.Status != models.BookingStatusInProgress {
		return nil, fmt.Errorf("booking must be in IN_PROGRESS status to check out")
	}

	// Calculate hours worked
	now := time.Now()
	hoursWorked := now.Sub(checkin.CheckInTime.Time).Hours()

	// Update check-out
	checkin.CheckOutTime = sql.NullTime{Time: now, Valid: true}
	checkin.CheckOutLatitude = sql.NullFloat64{Float64: latitude, Valid: true}
	checkin.CheckOutLongitude = sql.NullFloat64{Float64: longitude, Valid: true}
	checkin.TotalHoursWorked = sql.NullFloat64{Float64: hoursWorked, Valid: true}

	if err := s.checkinRepo.Update(checkin); err != nil {
		return nil, fmt.Errorf("failed to update checkin: %w", err)
	}

	// Complete the booking
	if _, err := s.bookingService.CompleteBooking(bookingID, cleanerID); err != nil {
		return nil, fmt.Errorf("failed to complete booking: %w", err)
	}

	return checkin, nil
}

// GetCheckinByBookingID gets a checkin by booking ID
func (s *CheckinService) GetCheckinByBookingID(bookingID string) (*models.Checkin, error) {
	return s.checkinRepo.GetByBookingID(bookingID)
}

// haversineDistance calculates the distance between two GPS coordinates in meters
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadiusMeters = 6371000

	// Convert to radians
	lat1Rad := lat1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	deltaLat := (lat2 - lat1) * math.Pi / 180
	deltaLon := (lon2 - lon1) * math.Pi / 180

	// Haversine formula
	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusMeters * c
}
