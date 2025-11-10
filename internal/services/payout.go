package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/cleanbuddy/backend/internal/models"
)

type PayoutService struct {
	payoutRepo   *models.PayoutRepository
	lineItemRepo *models.PayoutLineItemRepository
	bookingRepo  *models.BookingRepository
	cleanerRepo  *models.CleanerRepository
	userRepo     *models.UserRepository
	emailService *EmailService
}

func NewPayoutService(db *sql.DB, emailService *EmailService) *PayoutService {
	return &PayoutService{
		payoutRepo:   models.NewPayoutRepository(db),
		lineItemRepo: models.NewPayoutLineItemRepository(db),
		bookingRepo:  models.NewBookingRepository(db),
		cleanerRepo:  models.NewCleanerRepository(db),
		userRepo:     models.NewUserRepository(db),
		emailService: emailService,
	}
}

// GenerateMonthlyPayouts creates payout records for all cleaners for a given month
func (s *PayoutService) GenerateMonthlyPayouts(year int, month time.Month) ([]*models.Payout, error) {
	// Calculate period
	periodStart := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second) // Last second of the month

	// Get all completed bookings for the period
	bookings, err := s.bookingRepo.GetCompletedBookingsByPeriod(periodStart, periodEnd)
	if err != nil {
		return nil, fmt.Errorf("failed to get completed bookings: %w", err)
	}

	// Group bookings by cleaner
	cleanerBookings := make(map[string][]*models.Booking)
	for _, booking := range bookings {
		if booking.CleanerID.Valid {
			cleanerID := booking.CleanerID.String
			cleanerBookings[cleanerID] = append(cleanerBookings[cleanerID], booking)
		}
	}

	// Generate payout for each cleaner
	var payouts []*models.Payout
	for cleanerID, bookings := range cleanerBookings {
		// Get cleaner to retrieve user_id
		cleaner, err := s.cleanerRepo.GetByID(cleanerID)
		if err != nil {
			return nil, fmt.Errorf("failed to get cleaner %s: %w", cleanerID, err)
		}

		payout, err := s.calculatePayoutForCleaner(cleaner.UserID, bookings, periodStart, periodEnd)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate payout for cleaner %s: %w", cleanerID, err)
		}

		// Create payout record
		// Note: IBAN validation happens when marking payout as SENT
		if err := s.payoutRepo.Create(payout); err != nil {
			return nil, fmt.Errorf("failed to create payout: %w", err)
		}

		// Create line items
		for _, booking := range bookings {
			lineItem := s.createLineItem(payout.ID, booking)
			if err := s.lineItemRepo.Create(lineItem); err != nil {
				return nil, fmt.Errorf("failed to create line item: %w", err)
			}
		}

		payouts = append(payouts, payout)
	}

	return payouts, nil
}

// calculatePayoutForCleaner calculates earnings for a cleaner from their bookings
// userID is the user_id from the cleaners table (which references users.id)
func (s *PayoutService) calculatePayoutForCleaner(userID string, bookings []*models.Booking, periodStart, periodEnd time.Time) (*models.Payout, error) {
	var totalEarnings float64
	var platformFees float64
	totalBookings := len(bookings)

	for _, booking := range bookings {
		// Calculate platform fee rate
		// Repeat customers get reduced fee: 10% → 2%
		platformFeeRate := 10.0 // Default 10% for first-time customers

		// Check if this is a repeat customer
		isRepeatCustomer, err := s.isRepeatCustomer(booking.ClientID, booking.CompletedAt.Time)
		if err == nil && isRepeatCustomer {
			platformFeeRate = 2.0 // Reduced fee for repeat customers
		}

		// Platform fee calculation
		platformFee := booking.TotalPrice * (platformFeeRate / 100.0)

		totalEarnings += booking.TotalPrice
		platformFees += platformFee
	}

	netAmount := totalEarnings - platformFees

	payout := &models.Payout{
		CleanerID:     userID, // This is actually user_id per the schema
		PeriodStart:   periodStart,
		PeriodEnd:     periodEnd,
		Status:        models.PayoutStatusPending,
		TotalBookings: totalBookings,
		TotalEarnings: totalEarnings,
		PlatformFees:  platformFees,
		NetAmount:     netAmount,
	}

	return payout, nil
}

// createLineItem creates a payout line item from a booking
func (s *PayoutService) createLineItem(payoutID string, booking *models.Booking) *models.PayoutLineItem {
	// Calculate platform fee (same logic as above)
	platformFeeRate := 10.0 // Default 10% for first-time customers

	// Check if this is a repeat customer
	isRepeatCustomer, err := s.isRepeatCustomer(booking.ClientID, booking.CompletedAt.Time)
	if err == nil && isRepeatCustomer {
		platformFeeRate = 2.0 // Reduced fee for repeat customers
	}

	platformFee := booking.TotalPrice * (platformFeeRate / 100.0)
	cleanerEarnings := booking.TotalPrice - platformFee

	return &models.PayoutLineItem{
		PayoutID:        payoutID,
		BookingID:       booking.ID,
		BookingDate:     booking.ScheduledDate,
		ServiceType:     string(booking.ServiceType),
		BookingAmount:   booking.TotalPrice,
		PlatformFeeRate: platformFeeRate,
		PlatformFee:     platformFee,
		CleanerEarnings: cleanerEarnings,
	}
}

// GetPayoutsByCleanerID returns paginated payouts for a cleaner
func (s *PayoutService) GetPayoutsByCleanerID(cleanerID string, limit, offset int) ([]*models.Payout, error) {
	return s.payoutRepo.GetByCleanerID(cleanerID, limit, offset)
}

// GetPayoutWithLineItems returns a payout with its line items
func (s *PayoutService) GetPayoutWithLineItems(payoutID string) (*models.Payout, []*models.PayoutLineItem, error) {
	payout, err := s.payoutRepo.GetByID(payoutID)
	if err != nil {
		return nil, nil, err
	}
	if payout == nil {
		return nil, nil, fmt.Errorf("payout not found")
	}

	lineItems, err := s.lineItemRepo.GetByPayoutID(payoutID)
	if err != nil {
		return nil, nil, err
	}

	return payout, lineItems, nil
}

// MarkPayoutAsSent updates payout status to SENT
func (s *PayoutService) MarkPayoutAsSent(payoutID, transferReference string) error {
	payout, err := s.payoutRepo.GetByID(payoutID)
	if err != nil {
		return err
	}
	if payout == nil {
		return fmt.Errorf("payout not found")
	}

	// CRITICAL: Validate cleaner has IBAN before marking as sent
	if err := s.validateCleanerIBAN(payout.CleanerID); err != nil {
		return fmt.Errorf("cannot send payout: %w", err)
	}

	now := time.Now()
	payout.Status = models.PayoutStatusSent
	payout.TransferReference = sql.NullString{String: transferReference, Valid: true}
	payout.PaidAt = sql.NullTime{Time: now, Valid: true}

	if err := s.payoutRepo.Update(payout); err != nil {
		return err
	}

	// Send payout processed email to cleaner (async)
	go func() {
		ctx := context.Background()
		user, err := s.userRepo.GetByID(payout.CleanerID)
		if err == nil && user != nil && user.Email.Valid {
			cleanerName := "Cleaner"
			if user.FirstName.Valid {
				cleanerName = user.FirstName.String
			}

			periodStr := fmt.Sprintf("%s - %s",
				payout.PeriodStart.Format("02 January 2006"),
				payout.PeriodEnd.Format("02 January 2006"))

			transferRef := "N/A"
			if payout.TransferReference.Valid {
				transferRef = payout.TransferReference.String
			}

			_ = s.emailService.SendPayoutProcessedEmail(
				ctx,
				user.Email.String,
				cleanerName,
				payout.NetAmount,
				periodStr,
				transferRef,
			)
		}
	}()

	return nil
}

// MarkPayoutAsFailed updates payout status to FAILED
func (s *PayoutService) MarkPayoutAsFailed(payoutID, reason string) error {
	payout, err := s.payoutRepo.GetByID(payoutID)
	if err != nil {
		return err
	}
	if payout == nil {
		return fmt.Errorf("payout not found")
	}

	payout.Status = models.PayoutStatusFailed
	payout.FailedReason = sql.NullString{String: reason, Valid: true}

	return s.payoutRepo.Update(payout)
}

// GetPendingPayouts returns all payouts with PENDING status
func (s *PayoutService) GetPendingPayouts() ([]*models.Payout, error) {
	return s.payoutRepo.GetByStatus(models.PayoutStatusPending)
}

// GetPayoutsByStatus returns payouts filtered by status with optional pagination
// If status is empty, returns all payouts
func (s *PayoutService) GetPayoutsByStatus(status string, limit, offset *int) ([]*models.Payout, error) {
	if status == "" {
		// Return all payouts with pagination
		l := 100
		o := 0
		if limit != nil {
			l = *limit
		}
		if offset != nil {
			o = *offset
		}
		return s.payoutRepo.GetAll(l, o)
	}

	// Return payouts by status
	return s.payoutRepo.GetByStatus(status)
}

// isRepeatCustomer checks if a client has completed bookings before the given date
// Returns true if the client has at least one prior completed booking
func (s *PayoutService) isRepeatCustomer(clientID string, currentBookingDate time.Time) (bool, error) {
	// Get all completed bookings for this client before the current booking date
	previousBookings, err := s.bookingRepo.GetCompletedBookingsByClientBeforeDate(clientID, currentBookingDate)
	if err != nil {
		return false, fmt.Errorf("failed to check repeat customer status: %w", err)
	}

	// If there are any previous completed bookings, this is a repeat customer
	return len(previousBookings) > 0, nil
}

// validateCleanerIBAN checks if a cleaner has a valid IBAN for payouts
// Returns error if IBAN is missing or invalid
func (s *PayoutService) validateCleanerIBAN(cleanerID string) error {
	cleaner, err := s.cleanerRepo.GetByID(cleanerID)
	if err != nil {
		return fmt.Errorf("failed to get cleaner: %w", err)
	}
	if cleaner == nil {
		return fmt.Errorf("cleaner not found")
	}

	// Check if IBAN is present
	if !cleaner.IBAN.Valid || cleaner.IBAN.String == "" {
		return fmt.Errorf("cleaner does not have an IBAN configured - payouts cannot be sent")
	}

	// Basic Romanian IBAN validation: RO + 2 digits + 24 alphanumeric characters (total 28)
	iban := cleaner.IBAN.String
	if len(iban) < 24 || len(iban) > 34 {
		return fmt.Errorf("invalid IBAN length: %s", iban)
	}

	if iban[0:2] != "RO" {
		return fmt.Errorf("IBAN must start with 'RO' for Romanian accounts")
	}

	return nil
}

// GetCleanerPayouts gets payouts for a specific cleaner (admin only)
func (s *PayoutService) GetCleanerPayouts(cleanerID string, limit int) ([]*models.Payout, error) {
	offset := 0
	return s.payoutRepo.GetByCleanerID(cleanerID, limit, offset)
}
