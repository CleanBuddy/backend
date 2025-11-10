package services

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/cleanbuddy/backend/internal/models"
)

// DisputeService handles dispute business logic
type DisputeService struct {
	disputeRepo    *models.DisputeRepository
	bookingRepo    *models.BookingRepository
	paymentService *PaymentService
	bookingService *BookingService
	emailService   *EmailService
}

// NewDisputeService creates a new dispute service
func NewDisputeService(db *sql.DB) *DisputeService {
	return &DisputeService{
		disputeRepo: models.NewDisputeRepository(db),
		bookingRepo: models.NewBookingRepository(db),
	}
}

// SetPaymentService sets the payment service (to avoid circular dependency)
func (s *DisputeService) SetPaymentService(paymentService *PaymentService) {
	s.paymentService = paymentService
}

// SetBookingService sets the booking service (to avoid circular dependency)
func (s *DisputeService) SetBookingService(bookingService *BookingService) {
	s.bookingService = bookingService
}

// SetEmailService sets the email service (to avoid circular dependency)
func (s *DisputeService) SetEmailService(emailService *EmailService) {
	s.emailService = emailService
}

// CreateDispute creates a new dispute for a booking
func (s *DisputeService) CreateDispute(bookingID, userID, disputeType, description string) (*models.Dispute, error) {
	// Get booking to validate
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Validate user is the client who made the booking
	if booking.ClientID != userID {
		return nil, fmt.Errorf("unauthorized: only the booking client can create a dispute")
	}

	// Validate booking is completed
	if booking.Status != models.BookingStatusCompleted {
		return nil, fmt.Errorf("can only dispute completed bookings")
	}

	// Check if dispute already exists for this booking
	existingDispute, err := s.disputeRepo.GetByBookingID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing dispute: %w", err)
	}
	if existingDispute != nil {
		return nil, fmt.Errorf("dispute already exists for this booking")
	}

	// Validate dispute window (7 days after completion)
	if booking.CompletedAt.Valid {
		daysSinceCompletion := time.Since(booking.CompletedAt.Time).Hours() / 24
		if daysSinceCompletion > 7 {
			return nil, fmt.Errorf("dispute window has expired (7 days after completion)")
		}
	}

	// Validate dispute type
	validTypes := map[string]bool{
		models.DisputeTypeQualityIssue: true,
		models.DisputeTypeDamage:       true,
		models.DisputeTypeNoShow:       true,
		models.DisputeTypePricing:      true,
		models.DisputeTypeOther:        true,
	}
	if !validTypes[disputeType] {
		return nil, fmt.Errorf("invalid dispute type")
	}

	// Create dispute
	dispute := &models.Dispute{
		BookingID:   bookingID,
		CreatedBy:   userID,
		DisputeType: disputeType,
		Status:      models.DisputeStatusOpen,
		Description: description,
	}

	if err := s.disputeRepo.Create(dispute); err != nil {
		return nil, fmt.Errorf("failed to create dispute: %w", err)
	}

	// Update booking status to DISPUTED
	booking.Status = models.BookingStatusDisputed
	if err := s.bookingRepo.Update(booking); err != nil {
		// Log error but don't fail - dispute was created
		fmt.Printf("Warning: failed to update booking status to DISPUTED: %v\n", err)
	}

	return dispute, nil
}

// GetDispute gets a dispute by ID with authorization check
func (s *DisputeService) GetDispute(disputeID, userID string) (*models.Dispute, error) {
	dispute, err := s.disputeRepo.GetByID(disputeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dispute: %w", err)
	}
	if dispute == nil {
		return nil, fmt.Errorf("dispute not found")
	}

	// Get booking to check authorization
	booking, err := s.bookingRepo.GetByID(dispute.BookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	// Check if user is authorized (client or cleaner involved in booking)
	isClient := booking.ClientID == userID
	isCleaner := false
	if booking.CleanerID.Valid {
		// Need to check if this user is the cleaner
		// For now, we'll allow if user created the dispute or is the client
		isCleaner = booking.CleanerID.String == userID
	}

	if !isClient && !isCleaner {
		return nil, fmt.Errorf("unauthorized: not authorized to view this dispute")
	}

	return dispute, nil
}

// GetDisputeByBookingID gets a dispute by booking ID
func (s *DisputeService) GetDisputeByBookingID(bookingID, userID string) (*models.Dispute, error) {
	// Get booking to check authorization
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Check authorization
	isClient := booking.ClientID == userID
	isCleaner := false
	if booking.CleanerID.Valid {
		isCleaner = booking.CleanerID.String == userID
	}

	if !isClient && !isCleaner {
		return nil, fmt.Errorf("unauthorized")
	}

	dispute, err := s.disputeRepo.GetByBookingID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dispute: %w", err)
	}

	return dispute, nil
}

// AddCleanerResponse allows cleaner to respond to dispute
func (s *DisputeService) AddCleanerResponse(disputeID, userID, response string) (*models.Dispute, error) {
	dispute, err := s.disputeRepo.GetByID(disputeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dispute: %w", err)
	}
	if dispute == nil {
		return nil, fmt.Errorf("dispute not found")
	}

	// Get booking to validate cleaner
	booking, err := s.bookingRepo.GetByID(dispute.BookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	// Verify user is the cleaner
	if !booking.CleanerID.Valid || booking.CleanerID.String != userID {
		return nil, fmt.Errorf("unauthorized: only the assigned cleaner can respond")
	}

	// Update dispute with cleaner response
	now := time.Now()
	dispute.CleanerResponse = sql.NullString{String: response, Valid: true}
	dispute.CleanerRespondedAt = sql.NullTime{Time: now, Valid: true}
	dispute.Status = models.DisputeStatusUnderReview

	if err := s.disputeRepo.Update(dispute); err != nil {
		return nil, fmt.Errorf("failed to update dispute: %w", err)
	}

	return dispute, nil
}

// ResolveDispute resolves a dispute (admin only - validation done in resolver)
func (s *DisputeService) ResolveDispute(disputeID, adminID, resolutionType, resolutionNotes string, refundAmount float64) (*models.Dispute, error) {
	dispute, err := s.disputeRepo.GetByID(disputeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dispute: %w", err)
	}
	if dispute == nil {
		return nil, fmt.Errorf("dispute not found")
	}

	// Validate resolution type
	validResolutions := map[string]bool{
		models.DisputeResolutionPartialRefund: true,
		models.DisputeResolutionFullRefund:    true,
		models.DisputeResolutionReclean:       true,
		models.DisputeResolutionRejected:      true,
	}
	if !validResolutions[resolutionType] {
		return nil, fmt.Errorf("invalid resolution type")
	}

	// Update dispute
	now := time.Now()
	dispute.Status = models.DisputeStatusResolved
	dispute.ResolutionType = sql.NullString{String: resolutionType, Valid: true}
	dispute.ResolutionNotes = sql.NullString{String: resolutionNotes, Valid: true}
	if refundAmount > 0 {
		dispute.RefundAmount = sql.NullFloat64{Float64: refundAmount, Valid: true}
	}
	dispute.ResolvedAt = sql.NullTime{Time: now, Valid: true}
	dispute.ResolvedBy = sql.NullString{String: adminID, Valid: true}
	dispute.AssignedTo = sql.NullString{String: adminID, Valid: true}

	if err := s.disputeRepo.Update(dispute); err != nil {
		return nil, fmt.Errorf("failed to update dispute: %w", err)
	}

	// Get booking for additional processing
	booking, err := s.bookingRepo.GetByID(dispute.BookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}

	// Process refund if needed
	if resolutionType == models.DisputeResolutionPartialRefund || resolutionType == models.DisputeResolutionFullRefund {
		if refundAmount > 0 && s.paymentService != nil {
			// Find the payment for this booking
			payments, err := s.paymentService.GetPaymentsByBooking(dispute.BookingID, booking.ClientID)
			if err == nil && len(payments) > 0 {
				// Find captured payment
				for _, payment := range payments {
					if payment.Status == models.PaymentStatusCaptured {
						// Process refund
						_, err := s.paymentService.RefundPayment(payment.ID, refundAmount, fmt.Sprintf("Dispute resolution: %s", resolutionType))
						if err != nil {
							// Log error but don't fail dispute resolution
							fmt.Printf("Warning: Failed to process refund for dispute %s: %v\n", disputeID, err)
						} else {
							fmt.Printf("✅ Processed refund of %.2f RON for dispute %s\n", refundAmount, disputeID)
						}
						break
					}
				}
			}
		}
	}

	// Create reclean booking if needed
	if resolutionType == models.DisputeResolutionReclean && s.bookingService != nil {
		// Create a new booking with same details as original
		// Set cleaner_payout to 0 since this is a reclean (cleaner doesn't get paid twice)
		// Platform absorbs the cost as service recovery

		// Note: This creates a new PENDING booking that admin must manually assign
		// In the future, could auto-assign to a different cleaner
		fmt.Printf("📋 Reclean resolution for dispute %s - Admin should create follow-up booking\n", disputeID)

		// TODO: Implement automatic reclean booking creation
		// For now, admin must manually create the reclean booking
		// This prevents edge cases and allows admin to choose a different cleaner
	}

	// Send notifications
	if s.emailService != nil {
		// Notify client about resolution
		go func() {
			refundInfo := ""
			if refundAmount > 0 {
				refundInfo = fmt.Sprintf("\nRefund Amount: %.2f RON", refundAmount)
			}

			message := fmt.Sprintf("Your dispute for booking #%s has been resolved.\n\nResolution Type: %s\nResolution Notes: %s%s\n\nIf you have any questions, please contact our support team.\n\nBest regards,\nCleanBuddy Team",
				dispute.BookingID,
				resolutionType,
				resolutionNotes,
				refundInfo,
			)

			// In production, this would get the client's email from the user record
			// For now, just log
			fmt.Printf("📧 Would send email to client about dispute resolution:\n%s\n", message)
			// s.emailService.SendEmail(clientEmail, "Dispute Resolved - CleanBuddy", message)
		}()

		// Notify cleaner about resolution
		go func() {
			message := fmt.Sprintf("A dispute for booking #%s has been resolved by admin.\n\nResolution Type: %s\nResolution Notes: %s\n\nIf you have any questions, please contact our support team.\n\nBest regards,\nCleanBuddy Team",
				dispute.BookingID,
				resolutionType,
				resolutionNotes,
			)

			fmt.Printf("📧 Would send email to cleaner about dispute resolution:\n%s\n", message)
			// s.emailService.SendEmail(cleanerEmail, "Dispute Update - CleanBuddy", message)
		}()
	}

	return dispute, nil
}

// GetOpenDisputes gets all open disputes (admin only)
func (s *DisputeService) GetOpenDisputes(limit int) ([]*models.Dispute, error) {
	return s.disputeRepo.GetAllByStatus(models.DisputeStatusOpen, limit)
}
