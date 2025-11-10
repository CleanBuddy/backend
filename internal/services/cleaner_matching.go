package services

import (
	"database/sql"
	"fmt"
	"sort"
	"time"

	"github.com/cleanbuddy/backend/internal/models"
	"github.com/cleanbuddy/backend/internal/utils"
)

// CleanerMatchingService handles intelligent cleaner matching for bookings
type CleanerMatchingService struct {
	cleanerRepo       *models.CleanerRepository
	availabilityRepo  *models.AvailabilityRepository
	bookingRepo       *models.BookingRepository
	addressRepo       *models.AddressRepository
	emailService      *EmailService
}

// NewCleanerMatchingService creates a new cleaner matching service
func NewCleanerMatchingService(db *sql.DB, emailService *EmailService) *CleanerMatchingService {
	return &CleanerMatchingService{
		cleanerRepo:      models.NewCleanerRepository(db),
		availabilityRepo: models.NewAvailabilityRepository(db),
		bookingRepo:      models.NewBookingRepository(db),
		addressRepo:      models.NewAddressRepository(db),
		emailService:     emailService,
	}
}

// CleanerMatch represents a cleaner with their match score
type CleanerMatch struct {
	Cleaner           *models.Cleaner
	Score             float64
	DistanceScore     float64
	AvailabilityScore float64
	SkillScore        float64
	PerformanceScore  float64
	WorkloadScore     float64
	ReasonBreakdown   string
}

// MatchCleanersForBooking finds the best cleaners for a booking using intelligent scoring
func (s *CleanerMatchingService) MatchCleanersForBooking(bookingID string, limit int) ([]*CleanerMatch, error) {
	// Get booking details
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Get booking address
	address, err := s.addressRepo.GetByID(booking.AddressID)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}

	// Get all approved cleaners
	cleaners, err := s.cleanerRepo.GetByApprovalStatus(models.ApprovalStatusApproved)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaners: %w", err)
	}

	// Score each cleaner
	matches := make([]*CleanerMatch, 0)
	for _, cleaner := range cleaners {
		// Skip inactive cleaners
		if !cleaner.IsAvailable {
			continue
		}

		match := s.scoreCleanerForBooking(cleaner, booking, address)

		// Only include cleaners with score > 0 (i.e., they meet minimum requirements)
		if match.Score > 0 {
			matches = append(matches, match)
		}
	}

	// Sort by score (highest first)
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	// Return top N matches
	if len(matches) > limit {
		matches = matches[:limit]
	}

	return matches, nil
}

// scoreCleanerForBooking calculates a comprehensive match score (0-100)
func (s *CleanerMatchingService) scoreCleanerForBooking(
	cleaner *models.Cleaner,
	booking *models.Booking,
	address *models.Address,
) *CleanerMatch {
	match := &CleanerMatch{
		Cleaner: cleaner,
	}

	// 1. Location/Distance Score (30 points max)
	// Prioritize cleaners in the same city
	match.DistanceScore = s.calculateDistanceScore(cleaner, address)

	// 2. Availability Score (25 points max)
	// Check if cleaner is available at the requested time
	match.AvailabilityScore = s.calculateAvailabilityScore(cleaner, booking)

	// 3. Skill/Specialization Match Score (20 points max)
	// Match cleaner specializations with booking service type
	match.SkillScore = s.calculateSkillScore(cleaner, booking)

	// 4. Performance Score (15 points max)
	// Consider ratings and completion rate
	match.PerformanceScore = s.calculatePerformanceScore(cleaner)

	// 5. Workload Balance Score (10 points max)
	// Favor cleaners with fewer active bookings
	match.WorkloadScore = s.calculateWorkloadScore(cleaner)

	// Calculate total score
	match.Score = match.DistanceScore + match.AvailabilityScore +
	              match.SkillScore + match.PerformanceScore + match.WorkloadScore

	// Generate explanation
	match.ReasonBreakdown = fmt.Sprintf(
		"Distance: %.1f/30 | Availability: %.1f/25 | Skills: %.1f/20 | Performance: %.1f/15 | Workload: %.1f/10",
		match.DistanceScore, match.AvailabilityScore, match.SkillScore,
		match.PerformanceScore, match.WorkloadScore,
	)

	return match
}

// calculateDistanceScore scores based on location proximity (30 points max)
func (s *CleanerMatchingService) calculateDistanceScore(cleaner *models.Cleaner, address *models.Address) float64 {
	// Check if both cleaner and address have geolocation data
	hasCleanerCoords := cleaner.Latitude.Valid && cleaner.Longitude.Valid
	hasAddressCoords := address.Latitude.Valid && address.Longitude.Valid

	// If both have coordinates, use accurate distance calculation
	if hasCleanerCoords && hasAddressCoords {
		distance := utils.CalculateDistance(
			cleaner.Latitude.Float64, cleaner.Longitude.Float64,
			address.Latitude.Float64, address.Longitude.Float64,
		)

		// Score based on distance (30 points max)
		// 0-5 km: 30 points (perfect)
		// 5-10 km: 25 points (excellent)
		// 10-15 km: 20 points (good)
		// 15-25 km: 15 points (acceptable)
		// 25-40 km: 10 points (far)
		// 40+ km: 5 points (very far)

		if distance <= 5.0 {
			return 30.0
		} else if distance <= 10.0 {
			return 25.0
		} else if distance <= 15.0 {
			return 20.0
		} else if distance <= 25.0 {
			return 15.0
		} else if distance <= 40.0 {
			return 10.0
		} else {
			return 5.0
		}
	}

	// Fallback to city-based matching if no coordinates
	if !cleaner.City.Valid || address.City == "" {
		return 5.0 // Small score if city data missing
	}

	cleanerCity := cleaner.City.String
	addressCity := address.City

	// Exact city match = good score
	if cleanerCity == addressCity {
		return 25.0 // Slightly lower than perfect distance match
	}

	// Same county = partial points
	if cleaner.County.Valid && address.County != "" {
		if cleaner.County.String == address.County {
			return 15.0
		}
	}

	// Different location = low points
	return 5.0
}

// calculateAvailabilityScore checks if cleaner is available (25 points max)
func (s *CleanerMatchingService) calculateAvailabilityScore(cleaner *models.Cleaner, booking *models.Booking) float64 {
	// Get cleaner's availability
	availabilities, err := s.availabilityRepo.GetByCleanerID(cleaner.ID)
	if err != nil || len(availabilities) == 0 {
		return 0.0 // No availability = no match
	}

	// Extract day of week and time from booking
	dayOfWeek := int(booking.ScheduledDate.Weekday()) // 0 = Sunday, 1 = Monday, etc.
	bookingTime := booking.ScheduledTime
	bookingDate := booking.ScheduledDate

	// Check for matching availability
	hasRecurringMatch := false
	hasSpecificDateMatch := false

	for _, avail := range availabilities {
		if !avail.IsActive {
			continue
		}

		// Parse availability time strings (HH:MM format)
		startTime, err := parseTimeString(avail.StartTime)
		if err != nil {
			continue // Skip invalid time formats
		}
		endTime, err := parseTimeString(avail.EndTime)
		if err != nil {
			continue
		}

		// Check specific date availability
		if avail.Type == string(models.AvailabilityTypeOneTime) && avail.SpecificDate.Valid {
			specificDate := avail.SpecificDate.Time
			if isSameDay(specificDate, bookingDate) {
				if isTimeInRange(bookingTime, startTime, endTime) {
					hasSpecificDateMatch = true
					break
				}
			}
		}

		// Check recurring weekly availability
		if avail.Type == string(models.AvailabilityTypeRecurring) && avail.DayOfWeek.Valid {
			if int(avail.DayOfWeek.Int32) == dayOfWeek {
				if isTimeInRange(bookingTime, startTime, endTime) {
					hasRecurringMatch = true
				}
			}
		}
	}

	// Specific date match is better than recurring match
	if hasSpecificDateMatch {
		return 25.0 // Perfect availability
	}
	if hasRecurringMatch {
		return 20.0 // Good recurring availability
	}

	// Check if cleaner has any availability (flexible scheduling possible)
	if len(availabilities) > 0 {
		return 10.0 // Has availability, might be flexible
	}

	return 0.0 // No matching availability
}

// calculateSkillScore matches specializations with service type (20 points max)
func (s *CleanerMatchingService) calculateSkillScore(cleaner *models.Cleaner, booking *models.Booking) float64 {
	// Parse cleaner specializations from JSONB
	specializations, err := cleaner.ParseSpecializations()
	if err != nil {
		// If parsing fails, log error and return base score only
		fmt.Printf("Warning: Failed to parse specializations for cleaner %s: %v\n", cleaner.ID, err)
		return 10.0 // Base score for approved cleaner
	}

	serviceType := string(booking.ServiceType)
	score := 10.0 // Base score for approved cleaner

	// Map service types to specializations
	serviceToSpecialization := map[string]string{
		"STANDARD":    "Curățenie Standard",
		"DEEP":        "Curățenie Generală",
		"MOVE_IN_OUT": "Curățenie de Mutare",
		"POST_CONSTRUCTION": "După Renovare",
		"OFFICE":      "Birouri",
		"WINDOW":      "Geamuri",
	}

	requiredSpecialization, exists := serviceToSpecialization[serviceType]
	if !exists {
		return score
	}

	// Check if cleaner has matching specialization
	hasMatchingSpec := false
	for _, spec := range specializations {
		if spec == requiredSpecialization {
			score += 10.0 // Perfect skill match
			hasMatchingSpec = true
			break
		}
	}

	// Additional scoring for related skills
	if booking.IncludesWindows {
		for _, spec := range specializations {
			if spec == "Geamuri" {
				score += 2.0
				break
			}
		}
	}

	// Bonus for versatile cleaners (multiple specializations)
	if !hasMatchingSpec && len(specializations) >= 3 {
		score += 2.0 // Small bonus for experienced, versatile cleaners
	}

	// Cap at 20 points
	if score > 20.0 {
		score = 20.0
	}

	return score
}

// calculatePerformanceScore based on ratings and history (15 points max)
func (s *CleanerMatchingService) calculatePerformanceScore(cleaner *models.Cleaner) float64 {
	score := 0.0

	// Rating score (0-10 points)
	if cleaner.AverageRating.Valid {
		rating := cleaner.AverageRating.Float64
		// Linear scaling: 5.0 stars = 10 points, 3.0 stars = 5 points, etc.
		ratingScore := (rating / 5.0) * 10.0
		score += ratingScore
	} else {
		// No ratings yet = neutral score
		score += 5.0
	}

	// Experience score (0-5 points)
	totalJobs := cleaner.TotalJobs
	if totalJobs >= 100 {
		score += 5.0 // Veteran
	} else if totalJobs >= 50 {
		score += 4.0 // Very experienced
	} else if totalJobs >= 20 {
		score += 3.0 // Experienced
	} else if totalJobs >= 10 {
		score += 2.0 // Moderate
	} else if totalJobs >= 5 {
		score += 1.0 // New
	} else {
		score += 0.5 // Very new, but give them a chance
	}

	// Cap at 15 points
	if score > 15.0 {
		score = 15.0
	}

	return score
}

// calculateWorkloadScore favors cleaners with lighter workload (10 points max)
func (s *CleanerMatchingService) calculateWorkloadScore(cleaner *models.Cleaner) float64 {
	// Get count of active bookings for this cleaner (PENDING, CONFIRMED, IN_PROGRESS)
	activeBookings, err := s.bookingRepo.CountActiveBookingsByCleanerID(cleaner.ID)
	if err != nil {
		// If query fails, fall back to heuristic based on total jobs
		fmt.Printf("Warning: Failed to count active bookings for cleaner %s: %v\n", cleaner.ID, err)
		totalJobs := cleaner.TotalJobs
		if totalJobs == 0 {
			return 8.0
		} else if totalJobs < 10 {
			return 10.0
		} else if totalJobs < 50 {
			return 7.0
		} else {
			return 5.0
		}
	}

	// Score based on current active workload
	// Favor cleaners with availability, but give everyone a chance
	if activeBookings == 0 {
		// No active bookings = full availability
		if cleaner.TotalJobs == 0 {
			return 8.0 // New cleaner, slightly lower than experienced with no load
		}
		return 10.0 // Experienced cleaner with no current load = perfect
	} else if activeBookings == 1 {
		return 9.0 // Light load, very manageable
	} else if activeBookings == 2 {
		return 7.0 // Moderate load
	} else if activeBookings == 3 {
		return 5.0 // Busy but can handle one more
	} else if activeBookings == 4 {
		return 3.0 // Very busy
	} else {
		return 1.0 // Overloaded (5+ active bookings), still give a minimal chance
	}
}

// Helper functions

func isSameDay(date1, date2 time.Time) bool {
	y1, m1, d1 := date1.Date()
	y2, m2, d2 := date2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// parseTimeString parses time strings in HH:MM format
func parseTimeString(timeStr string) (time.Time, error) {
	// Parse time in HH:MM format
	t, err := time.Parse("15:04", timeStr)
	if err != nil {
		return time.Time{}, err
	}
	return t, nil
}

func isTimeInRange(checkTime, startTime, endTime time.Time) bool {
	// Extract hours and minutes
	checkHour := checkTime.Hour()
	checkMinute := checkTime.Minute()

	startHour := startTime.Hour()
	startMinute := startTime.Minute()

	endHour := endTime.Hour()
	endMinute := endTime.Minute()

	checkMinutes := checkHour*60 + checkMinute
	startMinutes := startHour*60 + startMinute
	endMinutes := endHour*60 + endMinute

	return checkMinutes >= startMinutes && checkMinutes <= endMinutes
}

// AutoAssignBestCleaner automatically assigns the best matching cleaner to a booking
func (s *CleanerMatchingService) AutoAssignBestCleaner(bookingID string) (*models.Cleaner, error) {
	// Find top 5 matches
	matches, err := s.MatchCleanersForBooking(bookingID, 5)
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no suitable cleaners found for booking")
	}

	// Get the best match
	bestMatch := matches[0]

	// Update booking with cleaner assignment
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, err
	}

	booking.CleanerID = sql.NullString{String: bestMatch.Cleaner.ID, Valid: true}
	booking.Status = models.BookingStatusConfirmed // Use CONFIRMED status for auto-assigned bookings

	if err := s.bookingRepo.Update(booking); err != nil {
		return nil, fmt.Errorf("failed to assign cleaner: %w", err)
	}

	return bestMatch.Cleaner, nil
}

// NotifyMatchedCleaners sends notifications to top matched cleaners (async)
func (s *CleanerMatchingService) NotifyMatchedCleaners(bookingID string, topN int) error {
	matches, err := s.MatchCleanersForBooking(bookingID, topN)
	if err != nil {
		return err
	}

	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return err
	}

	address, err := s.addressRepo.GetByID(booking.AddressID)
	if err != nil {
		return err
	}

	// Notify each matched cleaner asynchronously
	for i, match := range matches {
		go func(m *CleanerMatch, rank int) {
			// TODO: Send push notification or email
			fmt.Printf("Notifying cleaner %s (rank #%d, score: %.1f) about booking %s in %s\n",
				m.Cleaner.ID, rank+1, m.Score, bookingID, address.City)

			// Email notification would go here
			// s.emailService.SendJobOpportunityEmail(...)
		}(match, i)
	}

	return nil
}

// GetMatchScore calculates and returns the match score for a specific cleaner and booking
func (s *CleanerMatchingService) GetMatchScore(cleanerID, bookingID string) (*CleanerMatch, error) {
	cleaner, err := s.cleanerRepo.GetByID(cleanerID)
	if err != nil {
		return nil, err
	}

	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, err
	}

	address, err := s.addressRepo.GetByID(booking.AddressID)
	if err != nil {
		return nil, err
	}

	return s.scoreCleanerForBooking(cleaner, booking, address), nil
}
