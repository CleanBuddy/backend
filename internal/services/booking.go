package services

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/cleanbuddy/backend/internal/config"
	"github.com/cleanbuddy/backend/internal/models"
)

// BookingService handles booking business logic
type BookingService struct {
	bookingRepo     *models.BookingRepository
	cleanerRepo     *models.CleanerRepository
	addressRepo     *models.AddressRepository
	clientRepo      *models.ClientRepository
	userRepo        *models.UserRepository
	pricingService  *PricingService
	invoiceService  *InvoiceService
	paymentService  *PaymentService
	matchingService *CleanerMatchingService
	emailService    *EmailService
	cfg             *config.Config
}

// NewBookingService creates a new booking service
func NewBookingService(db *sql.DB, pricingService *PricingService, invoiceService *InvoiceService, emailService *EmailService) *BookingService {
	return &BookingService{
		bookingRepo:     models.NewBookingRepository(db),
		cleanerRepo:     models.NewCleanerRepository(db),
		addressRepo:     models.NewAddressRepository(db),
		clientRepo:      models.NewClientRepository(db),
		userRepo:        models.NewUserRepository(db),
		pricingService:  pricingService,
		invoiceService:  invoiceService,
		paymentService:  nil, // Will be set after PaymentService is created
		matchingService: nil, // Will be set after CleanerMatchingService is created
		emailService:    emailService,
		cfg:             config.Get(),
	}
}

// SetPaymentService sets the payment service (to break circular dependency)
func (s *BookingService) SetPaymentService(paymentService *PaymentService) {
	s.paymentService = paymentService
}

// SetMatchingService sets the cleaner matching service (to break circular dependency)
func (s *BookingService) SetMatchingService(matchingService *CleanerMatchingService) {
	s.matchingService = matchingService
}

// generateReservationCode generates a unique reservation code in format CB-YYYY-XXXXXX
func (s *BookingService) generateReservationCode() (string, error) {
	year := time.Now().Year()

	// Generate 6 random alphanumeric characters (uppercase)
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 6)

	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random code: %w", err)
		}
		code[i] = charset[num.Int64()]
	}

	return fmt.Sprintf("CB-%d-%s", year, string(code)), nil
}

// CreateBooking creates a new booking
func (s *BookingService) CreateBooking(
	clientID string,
	addressID string,
	serviceType models.ServiceType,
	areaSqm int,
	estimatedHours int,
	scheduledDate time.Time,
	scheduledTime time.Time,
	includesDeepCleaning bool,
	includesWindows bool,
	numberOfWindows int,
	includesCarpet bool,
	carpetAreaSqm int,
	includesFridge bool,
	includesOven bool,
	includesBalcony bool,
	specialInstructions string,
	accessInstructions string,
	supplies string, // Required: "client_provides" or "cleaner_provides"
	timePreferences string,
	frequency string,
) (*models.Booking, error) {
	// Validate supplies
	if supplies != "client_provides" && supplies != "cleaner_provides" {
		return nil, fmt.Errorf("supplies must be either 'client_provides' or 'cleaner_provides'")
	}
	// Validate address belongs to client
	address, err := s.addressRepo.GetByID(addressID)
	if err != nil {
		return nil, fmt.Errorf("failed to get address: %w", err)
	}
	if address == nil {
		return nil, fmt.Errorf("address not found")
	}
	if address.UserID != clientID {
		return nil, fmt.Errorf("address does not belong to client")
	}

	// Validate scheduling (only if fixed date/time is provided, not for flexible scheduling)
	if !scheduledDate.IsZero() && !scheduledTime.IsZero() {
		if err := s.validateScheduling(scheduledDate, scheduledTime); err != nil {
			return nil, err
		}
	}

	// Calculate pricing
	quote, err := s.pricingService.CalculatePrice(
		clientID,
		serviceType,
		areaSqm,
		estimatedHours,
		scheduledDate,
		scheduledTime,
		includesWindows,
		numberOfWindows,
		includesCarpet,
		carpetAreaSqm,
		includesFridge,
		includesOven,
		includesBalcony,
		false, // includesSupplies - will be added later
		frequency,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate price: %w", err)
	}

	// Generate reservation code
	reservationCode, err := s.generateReservationCode()
	if err != nil {
		return nil, fmt.Errorf("failed to generate reservation code: %w", err)
	}

	// Create booking
	booking := &models.Booking{
		ClientID:                clientID,
		AddressID:               addressID,
		ServiceType:             serviceType,
		EstimatedHours:          quote.EstimatedHours,
		ScheduledDate:           scheduledDate,
		ScheduledTime:           scheduledTime,
		IncludesDeepCleaning:    includesDeepCleaning,
		IncludesWindows:         includesWindows,
		IncludesCarpetCleaning:  includesCarpet,
		NumberOfWindows:         numberOfWindows,
		CarpetAreaSqm:           carpetAreaSqm,
		IncludesFridgeCleaning:  includesFridge,
		IncludesOvenCleaning:    includesOven,
		IncludesBalconyCleaning: includesBalcony,
		BasePrice:               quote.BasePrice,
		AddonsPrice:             quote.AddonsPrice,
		TotalPrice:              quote.TotalPrice,
		PlatformFee:             quote.PlatformFee,
		CleanerPayout:           quote.CleanerPayout,
		DiscountApplied:         quote.Discount,
		Status:                  models.BookingStatusPending,
		ReservationCode:         sql.NullString{String: reservationCode, Valid: true},
	}

	if areaSqm > 0 {
		booking.AreaSqm = sql.NullInt32{Int32: int32(areaSqm), Valid: true}
	}

	if specialInstructions != "" {
		booking.SpecialInstructions = sql.NullString{String: specialInstructions, Valid: true}
	}

	if accessInstructions != "" {
		booking.AccessInstructions = sql.NullString{String: accessInstructions, Valid: true}
	}

	// Supplies is required (already validated)
	booking.Supplies = sql.NullString{String: supplies, Valid: true}

	if timePreferences != "" {
		booking.TimePreferences = sql.NullString{String: timePreferences, Valid: true}
	}

	if frequency != "" {
		booking.Frequency = sql.NullString{String: frequency, Valid: true}
	}

	if err := s.bookingRepo.Create(booking); err != nil {
		return nil, fmt.Errorf("failed to create booking: %w", err)
	}

	// Send booking confirmation email to client (async, don't fail if email fails)
	go func() {
		ctx := context.Background()
		user, err := s.userRepo.GetByID(clientID)
		if err == nil && user != nil && user.Email.Valid {
			clientName := "Client"
			if user.FirstName.Valid && user.LastName.Valid {
				clientName = user.FirstName.String + " " + user.LastName.String
			} else if user.FirstName.Valid {
				clientName = user.FirstName.String
			}

			_ = s.emailService.SendBookingConfirmationEmail(
				ctx,
				user.Email.String,
				clientName,
				booking.ID,
				string(booking.ServiceType),
				booking.ScheduledDate.Format("02 January 2006"),
				booking.ScheduledTime.Format("15:04"),
				booking.TotalPrice,
			)
		}
	}()

	// Trigger intelligent cleaner matching algorithm (async)
	if s.matchingService != nil {
		go func() {
			// Find top 3 best matching cleaners
			matches, err := s.matchingService.MatchCleanersForBooking(booking.ID, 3)
			if err != nil {
				fmt.Printf("Warning: cleaner matching failed for booking %s: %v\n", booking.ID, err)
				return
			}

			if len(matches) == 0 {
				fmt.Printf("Warning: no suitable cleaners found for booking %s\n", booking.ID)
				return
			}

			// Log the top matches
			fmt.Printf("Found %d matching cleaners for booking %s:\n", len(matches), booking.ID)
			for i, match := range matches {
				fmt.Printf("  #%d: Cleaner %s (Score: %.1f/100) - %s\n",
					i+1, match.Cleaner.ID, match.Score, match.ReasonBreakdown)
			}

			// Notify top matched cleaners about the opportunity
			err = s.matchingService.NotifyMatchedCleaners(booking.ID, 3)
			if err != nil {
				fmt.Printf("Warning: failed to notify cleaners for booking %s: %v\n", booking.ID, err)
			}

			// Optional: Auto-assign the best cleaner if score is excellent (>80)
			// Uncomment to enable auto-assignment
			/*
			bestMatch := matches[0]
			if bestMatch.Score >= 80.0 {
				_, err := s.matchingService.AutoAssignBestCleaner(booking.ID)
				if err != nil {
					fmt.Printf("Warning: failed to auto-assign cleaner for booking %s: %v\n", booking.ID, err)
				} else {
					fmt.Printf("Auto-assigned cleaner %s to booking %s (score: %.1f)\n",
						bestMatch.Cleaner.ID, booking.ID, bestMatch.Score)
				}
			}
			*/
		}()
	}

	return booking, nil
}

// validateScheduling validates booking date and time
func (s *BookingService) validateScheduling(scheduledDate time.Time, scheduledTime time.Time) error {
	now := time.Now()

	// Combine date and time
	scheduled := time.Date(
		scheduledDate.Year(), scheduledDate.Month(), scheduledDate.Day(),
		scheduledTime.Hour(), scheduledTime.Minute(), 0, 0,
		scheduledDate.Location(),
	)

	// Check minimum advance booking
	minAdvance := time.Duration(s.cfg.Booking.MinAdvanceBookingHours) * time.Hour
	if scheduled.Before(now.Add(minAdvance)) {
		return fmt.Errorf("booking must be at least %d hours in advance", s.cfg.Booking.MinAdvanceBookingHours)
	}

	// Check maximum advance booking
	maxAdvance := time.Duration(s.cfg.Booking.MaxAdvanceBookingDays) * 24 * time.Hour
	if scheduled.After(now.Add(maxAdvance)) {
		return fmt.Errorf("booking cannot be more than %d days in advance", s.cfg.Booking.MaxAdvanceBookingDays)
	}

	// Check service hours
	hour := scheduledTime.Hour()
	if hour < s.cfg.Business.ServiceStartHour || hour >= s.cfg.Business.ServiceEndHour {
		return fmt.Errorf("service hours are between %d:00 and %d:00",
			s.cfg.Business.ServiceStartHour, s.cfg.Business.ServiceEndHour)
	}

	return nil
}

// GetBooking gets a booking by ID with ownership check
// AdminGetBooking gets a booking by ID without authorization (admin only)
func (s *BookingService) AdminGetBooking(bookingID string) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}
	return booking, nil
}

func (s *BookingService) GetBooking(bookingID string, userID string) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Check ownership (client or assigned cleaner)
	if booking.ClientID != userID {
		// Check if user is the assigned cleaner
		if booking.CleanerID.Valid {
			cleaner, err := s.cleanerRepo.GetByID(booking.CleanerID.String)
			if err != nil || cleaner == nil || cleaner.UserID != userID {
				return nil, fmt.Errorf("unauthorized")
			}
		} else {
			return nil, fmt.Errorf("unauthorized")
		}
	}

	return booking, nil
}

// GetClientBookings gets all bookings for a client
func (s *BookingService) GetClientBookings(clientID string) ([]*models.Booking, error) {
	return s.bookingRepo.GetByClientID(clientID)
}

// ConfirmBooking confirms a booking (PENDING → CONFIRMED)
func (s *BookingService) ConfirmBooking(bookingID string, cleanerID string) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Check status
	if booking.Status != models.BookingStatusPending {
		return nil, fmt.Errorf("booking is not in PENDING status")
	}

	// Verify cleaner is assigned
	if !booking.CleanerID.Valid || booking.CleanerID.String != cleanerID {
		return nil, fmt.Errorf("unauthorized")
	}

	// Confirm booking
	booking.Status = models.BookingStatusConfirmed
	now := time.Now()
	booking.ConfirmedAt = sql.NullTime{Time: now, Valid: true}

	if err := s.bookingRepo.Update(booking); err != nil {
		return nil, fmt.Errorf("failed to confirm booking: %w", err)
	}

	// Notify client about confirmation
	s.notifyBookingConfirmed(booking)

	return booking, nil
}

// CancelBooking cancels a booking
func (s *BookingService) CancelBooking(bookingID string, userID string, reason string) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Check if user can cancel (must be client or assigned cleaner)
	canCancel := booking.ClientID == userID
	if booking.CleanerID.Valid {
		// userID is user_id, need to check if this cleaner's user_id matches
		cleaner, err := s.cleanerRepo.GetByID(booking.CleanerID.String)
		if err == nil && cleaner != nil && cleaner.UserID == userID {
			canCancel = true
		}
	}
	if !canCancel {
		return nil, fmt.Errorf("unauthorized to cancel this booking")
	}

	// Validate current state - can only cancel if PENDING, CONFIRMED, or IN_PROGRESS
	if booking.Status != models.BookingStatusPending &&
	   booking.Status != models.BookingStatusConfirmed &&
	   booking.Status != models.BookingStatusInProgress {
		return nil, fmt.Errorf("cannot cancel booking in %s status", booking.Status)
	}

	// Check if already cancelled or completed
	if booking.Status == models.BookingStatusCancelled {
		return nil, fmt.Errorf("booking is already cancelled")
	}
	if booking.Status == models.BookingStatusCompleted {
		return nil, fmt.Errorf("cannot cancel completed booking")
	}

	// Check cancellation policy (free cancellation window)
	scheduled := time.Date(
		booking.ScheduledDate.Year(), booking.ScheduledDate.Month(), booking.ScheduledDate.Day(),
		booking.ScheduledTime.Hour(), booking.ScheduledTime.Minute(), 0, 0,
		booking.ScheduledDate.Location(),
	)

	freeWindow := time.Duration(s.cfg.Booking.CancellationFreeHours) * time.Hour
	if time.Now().Add(freeWindow).After(scheduled) {
		// TODO: Apply cancellation fee (future feature)
		// For now, we just note it in the reason
		reason = fmt.Sprintf("[Late cancellation] %s", reason)
	}

	// Update booking
	now := time.Now()
	booking.Status = models.BookingStatusCancelled
	booking.CancelledAt = sql.NullTime{Time: now, Valid: true}
	booking.CancelledBy = sql.NullString{String: userID, Valid: true}
	booking.CancellationReason = sql.NullString{String: reason, Valid: true}

	if err := s.bookingRepo.Update(booking); err != nil {
		return nil, fmt.Errorf("failed to cancel booking: %w", err)
	}

	// Send cancellation email to client and cleaner (async)
	go func() {
		ctx := context.Background()

		// Notify client
		clientUser, err := s.userRepo.GetByID(booking.ClientID)
		if err == nil && clientUser != nil && clientUser.Email.Valid {
			clientName := "Client"
			if clientUser.FirstName.Valid {
				clientName = clientUser.FirstName.String
			}

			cancelledBy := "You"
			if userID != booking.ClientID {
				cancelledBy = "the cleaner"
			}

			_ = s.emailService.SendBookingCancelledEmail(
				ctx,
				clientUser.Email.String,
				clientName,
				booking.ID,
				cancelledBy,
				reason,
			)
		}

		// Notify cleaner if assigned
		if booking.CleanerID.Valid {
			cleaner, err := s.cleanerRepo.GetByID(booking.CleanerID.String)
			if err == nil && cleaner != nil {
				cleanerUser, err := s.userRepo.GetByID(cleaner.UserID)
				if err == nil && cleanerUser != nil && cleanerUser.Email.Valid {
					cleanerName := "Cleaner"
					if cleanerUser.FirstName.Valid {
						cleanerName = cleanerUser.FirstName.String
					}

					cancelledBy := "the client"
					if userID == cleaner.UserID {
						cancelledBy = "You"
					}

					_ = s.emailService.SendBookingCancelledEmail(
						ctx,
						cleanerUser.Email.String,
						cleanerName,
						booking.ID,
						cancelledBy,
						reason,
					)
				}
			}
		}
	}()

	// TODO: Process refund if applicable

	return booking, nil
}

// AssignCleaner assigns a cleaner to a booking (admin or auto-match)
func (s *BookingService) AssignCleaner(bookingID string, cleanerID string) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Check booking status
	if booking.Status != models.BookingStatusPending {
		return nil, fmt.Errorf("booking is not in PENDING status")
	}

	// Verify cleaner exists and is approved
	cleaner, err := s.cleanerRepo.GetByUserID(cleanerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaner: %w", err)
	}
	if cleaner == nil {
		return nil, fmt.Errorf("cleaner not found")
	}
	if cleaner.ApprovalStatus != models.ApprovalStatusApproved {
		return nil, fmt.Errorf("cleaner is not approved")
	}
	if !cleaner.IsActive || !cleaner.IsAvailable {
		return nil, fmt.Errorf("cleaner is not available")
	}

	// Assign cleaner
	booking.CleanerID = sql.NullString{String: cleanerID, Valid: true}
	booking.Status = models.BookingStatusConfirmed
	now := time.Now()
	booking.ConfirmedAt = sql.NullTime{Time: now, Valid: true}

	if err := s.bookingRepo.Update(booking); err != nil {
		return nil, fmt.Errorf("failed to assign cleaner: %w", err)
	}

	// Notify client and cleaner about assignment
	s.notifyCleanerAssigned(booking)

	return booking, nil
}

// StartBooking marks a booking as in progress
func (s *BookingService) StartBooking(bookingID string, cleanerID string) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Verify cleaner - cleanerID parameter is user_id, need to resolve to cleaner.ID
	if !booking.CleanerID.Valid {
		return nil, fmt.Errorf("unauthorized")
	}

	cleaner, err := s.cleanerRepo.GetByID(booking.CleanerID.String)
	if err != nil || cleaner == nil || cleaner.UserID != cleanerID {
		return nil, fmt.Errorf("unauthorized")
	}

	// Check status
	if booking.Status != models.BookingStatusConfirmed {
		return nil, fmt.Errorf("booking is not in CONFIRMED status")
	}

	// Start booking
	booking.Status = models.BookingStatusInProgress
	now := time.Now()
	booking.StartedAt = sql.NullTime{Time: now, Valid: true}

	if err := s.bookingRepo.Update(booking); err != nil {
		return nil, fmt.Errorf("failed to start booking: %w", err)
	}

	return booking, nil
}

// CompleteBooking marks a booking as completed
func (s *BookingService) CompleteBooking(bookingID string, cleanerID string) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Verify cleaner - cleanerID parameter is user_id, need to resolve to cleaner.ID
	if !booking.CleanerID.Valid {
		return nil, fmt.Errorf("unauthorized")
	}

	cleaner, err := s.cleanerRepo.GetByID(booking.CleanerID.String)
	if err != nil || cleaner == nil || cleaner.UserID != cleanerID {
		return nil, fmt.Errorf("unauthorized")
	}

	// Check status
	if booking.Status != models.BookingStatusInProgress {
		return nil, fmt.Errorf("booking is not in IN_PROGRESS status")
	}

	// Complete booking
	booking.Status = models.BookingStatusCompleted
	now := time.Now()
	booking.CompletedAt = sql.NullTime{Time: now, Valid: true}

	if err := s.bookingRepo.Update(booking); err != nil {
		return nil, fmt.Errorf("failed to complete booking: %w", err)
	}

	// Auto-create invoice for completed booking
	if s.invoiceService != nil {
		_, err := s.invoiceService.CreateInvoiceForBooking(bookingID)
		if err != nil {
			// Log error but don't fail the completion
			fmt.Printf("Warning: failed to create invoice for booking %s: %v\n", bookingID, err)
		}
	}

	// Send booking completed email to client (async)
	go func() {
		ctx := context.Background()
		clientUser, err := s.userRepo.GetByID(booking.ClientID)
		if err == nil && clientUser != nil && clientUser.Email.Valid {
			clientName := "Client"
			if clientUser.FirstName.Valid {
				clientName = clientUser.FirstName.String
			}

			_ = s.emailService.SendBookingCompletedEmail(
				ctx,
				clientUser.Email.String,
				clientName,
				booking.ID,
				booking.TotalPrice,
				"https://cleanbuddy.ro/client/bookings/"+booking.ID+"/review",
			)
		}
	}()

	// Trigger payment capture for authorized payments
	if s.paymentService != nil {
		payments, err := s.paymentService.GetPaymentsByBooking(bookingID, booking.ClientID)
		if err == nil && len(payments) > 0 {
			// Find authorized payment
			for _, payment := range payments {
				if payment.Status == models.PaymentStatusAuthorized {
					_, err := s.paymentService.CapturePayment(payment.ID)
					if err != nil {
						// Log error but don't fail the completion
						fmt.Printf("Warning: failed to capture payment %s for booking %s: %v\n", payment.ID, bookingID, err)
					}
					break // Only capture the first authorized payment
				}
			}
		}
	}

	// Update cleaner stats (total jobs, earnings)
	if booking.CleanerID.Valid {
		cleaner, err := s.cleanerRepo.GetByID(booking.CleanerID.String)
		if err == nil && cleaner != nil {
			cleaner.TotalJobs++
			cleaner.TotalEarnings += booking.CleanerPayout
			if err := s.cleanerRepo.Update(cleaner); err != nil {
				// Log error but don't fail the completion
				fmt.Printf("Warning: failed to update cleaner stats for %s: %v\n", cleaner.ID, err)
			}
		}
	}

	// Update client stats (total bookings, total spent)
	client, err := s.clientRepo.GetByUserID(booking.ClientID)
	if err == nil && client != nil {
		client.TotalBookings++
		client.TotalSpent += booking.TotalPrice
		if err := s.clientRepo.Update(client); err != nil {
			// Log error but don't fail the completion
			fmt.Printf("Warning: failed to update client stats for %s: %v\n", client.ID, err)
		}
	}

	return booking, nil
}

// GetAvailableJobs returns bookings available for cleaners to accept
func (s *BookingService) GetAvailableJobs(cleanerID string, city string, limit, offset int) ([]*models.Booking, error) {
	// Verify cleaner exists and is approved
	cleaner, err := s.cleanerRepo.GetByUserID(cleanerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaner: %w", err)
	}
	if cleaner == nil {
		return nil, fmt.Errorf("cleaner not found")
	}
	if cleaner.ApprovalStatus != models.ApprovalStatusApproved {
		return nil, fmt.Errorf("cleaner is not approved")
	}

	return s.bookingRepo.GetAvailableJobs(city, limit, offset)
}

// AcceptBooking allows a cleaner to accept a job
func (s *BookingService) AcceptBooking(bookingID string, cleanerID string) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Check booking is available (no cleaner assigned and PENDING status)
	if booking.CleanerID.Valid {
		return nil, fmt.Errorf("booking already has an assigned cleaner")
	}
	if booking.Status != models.BookingStatusPending {
		return nil, fmt.Errorf("booking is not in PENDING status")
	}

	// Verify cleaner exists and is approved
	cleaner, err := s.cleanerRepo.GetByUserID(cleanerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaner: %w", err)
	}
	if cleaner == nil {
		return nil, fmt.Errorf("cleaner not found")
	}
	if cleaner.ApprovalStatus != models.ApprovalStatusApproved {
		return nil, fmt.Errorf("cleaner is not approved")
	}
	if !cleaner.IsActive || !cleaner.IsAvailable {
		return nil, fmt.Errorf("cleaner is not available")
	}

	// Assign cleaner and confirm booking
	booking.CleanerID = sql.NullString{String: cleaner.ID, Valid: true}
	booking.Status = models.BookingStatusConfirmed
	now := time.Now()
	booking.ConfirmedAt = sql.NullTime{Time: now, Valid: true}

	if err := s.bookingRepo.Update(booking); err != nil {
		return nil, fmt.Errorf("failed to accept booking: %w", err)
	}

	// Send booking accepted email to client (async)
	go func() {
		ctx := context.Background()
		user, err := s.userRepo.GetByID(booking.ClientID)
		if err == nil && user != nil && user.Email.Valid {
			clientName := "Client"
			if user.FirstName.Valid && user.LastName.Valid {
				clientName = user.FirstName.String + " " + user.LastName.String
			} else if user.FirstName.Valid {
				clientName = user.FirstName.String
			}

			// Get cleaner info
			cleanerUser, err := s.userRepo.GetByID(cleanerID)
			cleanerName := "Your cleaner"
			if err == nil && cleanerUser != nil {
				if cleanerUser.FirstName.Valid {
					cleanerName = cleanerUser.FirstName.String
				}
			}

			_ = s.emailService.SendBookingAcceptedEmail(
				ctx,
				user.Email.String,
				clientName,
				cleanerName,
				booking.ID,
				booking.ScheduledDate.Format("02 January 2006"),
				booking.ScheduledTime.Format("15:04"),
			)
		}
	}()

	return booking, nil
}

// AcceptBookingWithTime allows a cleaner to accept a job and optionally set the scheduled time
func (s *BookingService) AcceptBookingWithTime(bookingID string, cleanerID string, scheduledDate *time.Time, scheduledTime *time.Time) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Check booking is available (no cleaner assigned and PENDING status)
	if booking.CleanerID.Valid {
		return nil, fmt.Errorf("booking already has an assigned cleaner")
	}
	if booking.Status != models.BookingStatusPending {
		return nil, fmt.Errorf("booking is not in PENDING status")
	}

	// Verify cleaner exists and is approved
	cleaner, err := s.cleanerRepo.GetByUserID(cleanerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleaner: %w", err)
	}
	if cleaner == nil {
		return nil, fmt.Errorf("cleaner not found")
	}
	if cleaner.ApprovalStatus != models.ApprovalStatusApproved {
		return nil, fmt.Errorf("cleaner is not approved")
	}
	if !cleaner.IsActive || !cleaner.IsAvailable {
		return nil, fmt.Errorf("cleaner is not available")
	}

	// If scheduledDate and scheduledTime are provided, update the booking
	if scheduledDate != nil && scheduledTime != nil {
		booking.ScheduledDate = *scheduledDate
		booking.ScheduledTime = *scheduledTime
	} else if scheduledDate != nil {
		booking.ScheduledDate = *scheduledDate
	} else if scheduledTime != nil {
		booking.ScheduledTime = *scheduledTime
	}
	// If neither is provided, keep existing values (or they remain as zero values)

	// Assign cleaner and confirm booking
	booking.CleanerID = sql.NullString{String: cleaner.ID, Valid: true}
	booking.Status = models.BookingStatusConfirmed
	now := time.Now()
	booking.ConfirmedAt = sql.NullTime{Time: now, Valid: true}

	if err := s.bookingRepo.Update(booking); err != nil {
		return nil, fmt.Errorf("failed to accept booking: %w", err)
	}

	// Send booking accepted email to client (async)
	go func() {
		ctx := context.Background()
		user, err := s.userRepo.GetByID(booking.ClientID)
		if err == nil && user != nil && user.Email.Valid {
			clientName := "Client"
			if user.FirstName.Valid && user.LastName.Valid {
				clientName = user.FirstName.String + " " + user.LastName.String
			} else if user.FirstName.Valid {
				clientName = user.FirstName.String
			}

			// Get cleaner info
			cleanerUser, err := s.userRepo.GetByID(cleanerID)
			cleanerName := "Your cleaner"
			if err == nil && cleanerUser != nil {
				if cleanerUser.FirstName.Valid {
					cleanerName = cleanerUser.FirstName.String
				}
			}

			_ = s.emailService.SendBookingAcceptedEmail(
				ctx,
				user.Email.String,
				clientName,
				cleanerName,
				booking.ID,
				booking.ScheduledDate.Format("02 January 2006"),
				booking.ScheduledTime.Format("15:04"),
			)
		}
	}()

	return booking, nil
}

// DeclineBooking allows a cleaner to decline a job
func (s *BookingService) DeclineBooking(bookingID string, cleanerID string, reason string) (bool, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return false, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return false, fmt.Errorf("booking not found")
	}

	// Verify booking is available
	if booking.CleanerID.Valid {
		return false, fmt.Errorf("booking already has an assigned cleaner")
	}
	if booking.Status != models.BookingStatusPending {
		return false, fmt.Errorf("booking is not in PENDING status")
	}

	// TODO: Optionally log the decline in a separate table for analytics
	// For now, we just return success

	return true, nil
}

// GetClientBookingsFiltered gets bookings for a client with filter
func (s *BookingService) GetClientBookingsFiltered(clientID string, filter string) ([]*models.Booking, error) {
	return s.bookingRepo.GetByClientIDFiltered(clientID, filter)
}

// AdminReassignBooking allows admin to reassign a booking to a different cleaner
func (s *BookingService) AdminReassignBooking(bookingID string, cleanerID string) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	// Verify cleaner exists and is approved
	cleaner, err := s.cleanerRepo.GetByID(cleanerID)
	if err != nil {
		return nil, fmt.Errorf("cleaner not found: %w", err)
	}

	if cleaner.ApprovalStatus != models.ApprovalStatusApproved {
		return nil, fmt.Errorf("cleaner is not approved")
	}

	// Update the booking
	booking.CleanerID = sql.NullString{String: cleanerID, Valid: true}
	if err := s.bookingRepo.Update(booking); err != nil {
		return nil, fmt.Errorf("failed to reassign booking: %w", err)
	}

	return booking, nil
}

// AdminCancelBooking allows admin to cancel any booking with a reason
func (s *BookingService) AdminCancelBooking(bookingID string, reason string) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	if booking.Status == models.BookingStatusCancelled {
		return nil, fmt.Errorf("booking is already cancelled")
	}

	now := time.Now()
	booking.Status = models.BookingStatusCancelled
	booking.CancellationReason = sql.NullString{String: reason, Valid: true}
	booking.CancelledAt = sql.NullTime{Time: now, Valid: true}
	// Use a special admin user ID to indicate admin cancellation
	booking.CancelledBy = sql.NullString{String: "admin", Valid: true}

	if err := s.bookingRepo.Update(booking); err != nil {
		return nil, fmt.Errorf("failed to cancel booking: %w", err)
	}

	return booking, nil
}

// AdminUpdateBookingStatus allows admin to update booking status
func (s *BookingService) AdminUpdateBookingStatus(bookingID string, status models.BookingStatus) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	booking.Status = status

	// Set appropriate timestamps based on status
	now := time.Now()
	switch status {
	case models.BookingStatusConfirmed:
		if !booking.ConfirmedAt.Valid {
			booking.ConfirmedAt = sql.NullTime{Time: now, Valid: true}
		}
	case models.BookingStatusInProgress:
		if !booking.StartedAt.Valid {
			booking.StartedAt = sql.NullTime{Time: now, Valid: true}
		}
	case models.BookingStatusCompleted:
		if !booking.CompletedAt.Valid {
			booking.CompletedAt = sql.NullTime{Time: now, Valid: true}
		}
	case models.BookingStatusCancelled:
		if !booking.CancelledAt.Valid {
			booking.CancelledAt = sql.NullTime{Time: now, Valid: true}
		}
	}

	if err := s.bookingRepo.Update(booking); err != nil {
		return nil, fmt.Errorf("failed to update booking status: %w", err)
	}

	return booking, nil
}

// AdminEditBooking allows admin to edit booking details
func (s *BookingService) AdminEditBooking(bookingID string, scheduledDate *time.Time, scheduledTime *time.Time, serviceType *models.ServiceType, estimatedHours *int, areaSqm *int, specialInstructions *string, accessInstructions *string) (*models.Booking, error) {
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	// Update fields if provided
	if scheduledDate != nil {
		booking.ScheduledDate = *scheduledDate
	}
	if scheduledTime != nil {
		booking.ScheduledTime = *scheduledTime
	}
	if serviceType != nil {
		booking.ServiceType = *serviceType
	}
	if estimatedHours != nil {
		booking.EstimatedHours = *estimatedHours
	}
	if areaSqm != nil {
		booking.AreaSqm = sql.NullInt32{Int32: int32(*areaSqm), Valid: true}
	}
	if specialInstructions != nil {
		booking.SpecialInstructions = sql.NullString{String: *specialInstructions, Valid: true}
	}
	if accessInstructions != nil {
		booking.AccessInstructions = sql.NullString{String: *accessInstructions, Valid: true}
	}

	if err := s.bookingRepo.Update(booking); err != nil {
		return nil, fmt.Errorf("failed to update booking: %w", err)
	}

	return booking, nil
}

// GetBookingsByCleanerIDs gets all bookings for multiple cleaners with optional filter
func (s *BookingService) GetBookingsByCleanerIDs(cleanerIDs []string, filter *string) ([]*models.Booking, error) {
	if len(cleanerIDs) == 0 {
		return []*models.Booking{}, nil
	}

	bookings, err := s.bookingRepo.GetByCleanerIDs(cleanerIDs, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get bookings by cleaner IDs: %w", err)
	}

	return bookings, nil
}

// ExpireOldPendingBookings automatically expires PENDING bookings older than the specified hours
// This is called by a scheduler to auto-expire bookings that haven't been accepted
func (s *BookingService) ExpireOldPendingBookings(expirationHours int) (int, error) {
	// Get all pending bookings older than the expiration time
	expiredBookings, err := s.bookingRepo.GetExpiredPendingBookings(expirationHours)
	if err != nil {
		return 0, fmt.Errorf("failed to get expired bookings: %w", err)
	}

	expiredCount := 0
	for _, booking := range expiredBookings {
		// Cancel the booking with auto-expiration reason
		booking.Status = models.BookingStatusCancelled
		booking.CancelledAt = sql.NullTime{Time: time.Now(), Valid: true}
		booking.CancellationReason = sql.NullString{
			String: fmt.Sprintf("Booking auto-expired after %d hours without acceptance", expirationHours),
			Valid:  true,
		}
		// Set CancelledBy to NULL for system-initiated cancellations (no foreign key violation)
		booking.CancelledBy = sql.NullString{String: "", Valid: false}

		if err := s.bookingRepo.Update(booking); err != nil {
			// Log error but continue with other bookings
			fmt.Printf("Failed to expire booking %s: %v\n", booking.ID, err)
			continue
		}

		expiredCount++

		// Notify client about booking expiration
		s.notifyBookingExpired(booking)
	}

	return expiredCount, nil
}

// Notification helper methods

// notifyBookingConfirmed sends confirmation notification to client
func (s *BookingService) notifyBookingConfirmed(booking *models.Booking) {
	if s.emailService == nil {
		return
	}

	go func() {
		message := fmt.Sprintf("Your booking #%s has been confirmed!\n\nScheduled Date: %s\nScheduled Time: %s\nService Type: %s\nTotal Price: %.2f RON\n\nYou will receive cleaner details shortly.\n\nBest regards,\nCleanBuddy Team",
			booking.ID,
			booking.ScheduledDate.Format("2006-01-02"),
			booking.ScheduledTime.Format("15:04"),
			booking.ServiceType,
			booking.TotalPrice,
		)
		fmt.Printf("📧 Would send booking confirmation email to client:\n%s\n", message)
		// s.emailService.SendEmail(clientEmail, "Booking Confirmed - CleanBuddy", message)
	}()
}

// notifyCleanerAssigned sends notifications to both client and cleaner
func (s *BookingService) notifyCleanerAssigned(booking *models.Booking) {
	if s.emailService == nil {
		return
	}

	// Notify client
	go func() {
		message := fmt.Sprintf("Great news! A cleaner has been assigned to your booking.\n\nBooking ID: #%s\nScheduled Date: %s\nScheduled Time: %s\n\nYour cleaner will arrive at the scheduled time.\n\nBest regards,\nCleanBuddy Team",
			booking.ID,
			booking.ScheduledDate.Format("2006-01-02"),
			booking.ScheduledTime.Format("15:04"),
		)
		fmt.Printf("📧 Would send cleaner assignment email to client:\n%s\n", message)
		// s.emailService.SendEmail(clientEmail, "Cleaner Assigned - CleanBuddy", message)
	}()

	// Notify cleaner
	go func() {
		message := fmt.Sprintf("You have been assigned a new booking!\n\nBooking ID: #%s\nScheduled Date: %s\nScheduled Time: %s\nService Type: %s\n\nPlease review the booking details in your dashboard.\n\nBest regards,\nCleanBuddy Team",
			booking.ID,
			booking.ScheduledDate.Format("2006-01-02"),
			booking.ScheduledTime.Format("15:04"),
			booking.ServiceType,
		)
		fmt.Printf("📧 Would send new booking email to cleaner:\n%s\n", message)
		// s.emailService.SendEmail(cleanerEmail, "New Booking Assigned - CleanBuddy", message)
	}()
}

// notifyBookingExpired sends expiration notification to client
func (s *BookingService) notifyBookingExpired(booking *models.Booking) {
	if s.emailService == nil {
		return
	}

	go func() {
		message := fmt.Sprintf("Your booking #%s has been automatically cancelled due to no cleaner availability.\n\nScheduled Date: %s\n\nWe apologize for the inconvenience. You have not been charged. Please try booking again or contact our support team.\n\nBest regards,\nCleanBuddy Team",
			booking.ID,
			booking.ScheduledDate.Format("2006-01-02"),
		)
		fmt.Printf("📧 Would send booking expired email to client:\n%s\n", message)
		// s.emailService.SendEmail(clientEmail, "Booking Cancelled - CleanBuddy", message)
	}()
}

// GetCleanerBookingsAdmin gets bookings for a specific cleaner (admin only)
func (s *BookingService) GetCleanerBookingsAdmin(cleanerID string, limit, offset int, status *models.BookingStatus, search *string) ([]*models.Booking, error) {
	return s.bookingRepo.GetByCleanerID(cleanerID, limit, offset, status, search)
}
