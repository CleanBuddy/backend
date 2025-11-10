package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cleanbuddy/backend/internal/config"
	"github.com/cleanbuddy/backend/internal/models"
)

// PaymentService handles payment processing
type PaymentService struct {
	paymentRepo *models.PaymentRepository
	bookingRepo *models.BookingRepository
	cfg         *config.Config
}

// NewPaymentService creates a new payment service
func NewPaymentService(db *sql.DB) *PaymentService {
	return &PaymentService{
		paymentRepo: models.NewPaymentRepository(db),
		bookingRepo: models.NewBookingRepository(db),
		cfg:         config.Get(),
	}
}

// PreauthorizePayment creates a payment preauthorization for a booking
// This holds the funds on the customer's card without capturing them
func (s *PaymentService) PreauthorizePayment(
	bookingID string,
	userID string,
	amount float64,
	provider models.PaymentProvider,
) (*models.Payment, error) {
	// Validate booking exists and belongs to user
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}
	if booking.ClientID != userID {
		return nil, fmt.Errorf("booking does not belong to user")
	}

	// Create payment record
	payment := &models.Payment{
		BookingID:   bookingID,
		UserID:      userID,
		Provider:    provider,
		PaymentType: models.PaymentTypePreauthorization,
		Status:      models.PaymentStatusPending,
		Amount:      amount,
		Currency:    "RON",
	}

	// Call payment provider based on type
	switch provider {
	case models.PaymentProviderNetopia:
		return s.netopiaPreauthorize(payment)
	case models.PaymentProviderManual:
		return s.manualPreauthorize(payment)
	default:
		return nil, fmt.Errorf("unsupported payment provider: %s", provider)
	}
}

// CapturePayment captures a preauthorized payment
// This actually charges the customer's card
func (s *PaymentService) CapturePayment(paymentID string) (*models.Payment, error) {
	// Get payment
	payment, err := s.paymentRepo.GetByID(paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	if payment == nil {
		return nil, fmt.Errorf("payment not found")
	}

	// Validate status
	if payment.Status != models.PaymentStatusAuthorized {
		return nil, fmt.Errorf("payment is not authorized (status: %s)", payment.Status)
	}

	// Call payment provider
	switch payment.Provider {
	case models.PaymentProviderNetopia:
		return s.netopiaCapture(payment)
	case models.PaymentProviderManual:
		return s.manualCapture(payment)
	default:
		return nil, fmt.Errorf("unsupported payment provider: %s", payment.Provider)
	}
}

// RefundPayment refunds a captured payment
func (s *PaymentService) RefundPayment(paymentID string, amount float64, reason string) (*models.Payment, error) {
	// Get original payment
	originalPayment, err := s.paymentRepo.GetByID(paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	if originalPayment == nil {
		return nil, fmt.Errorf("payment not found")
	}

	// Validate status
	if originalPayment.Status != models.PaymentStatusCaptured {
		return nil, fmt.Errorf("payment is not captured (status: %s)", originalPayment.Status)
	}

	// Validate amount
	if amount > originalPayment.Amount {
		return nil, fmt.Errorf("refund amount cannot exceed original amount")
	}

	// Create refund payment record
	refundPayment := &models.Payment{
		BookingID:   originalPayment.BookingID,
		UserID:      originalPayment.UserID,
		Provider:    originalPayment.Provider,
		PaymentType: models.PaymentTypeRefund,
		Status:      models.PaymentStatusPending,
		Amount:      amount,
		Currency:    originalPayment.Currency,
	}

	// Call payment provider
	switch originalPayment.Provider {
	case models.PaymentProviderNetopia:
		return s.netopiaRefund(refundPayment, originalPayment)
	case models.PaymentProviderManual:
		return s.manualRefund(refundPayment)
	default:
		return nil, fmt.Errorf("unsupported payment provider: %s", originalPayment.Provider)
	}
}

// CancelPreauthorization cancels a preauthorized payment
func (s *PaymentService) CancelPreauthorization(paymentID string) (*models.Payment, error) {
	// Get payment
	payment, err := s.paymentRepo.GetByID(paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	if payment == nil {
		return nil, fmt.Errorf("payment not found")
	}

	// Validate status
	if payment.Status != models.PaymentStatusAuthorized {
		return nil, fmt.Errorf("payment is not authorized (status: %s)", payment.Status)
	}

	// Call payment provider
	switch payment.Provider {
	case models.PaymentProviderNetopia:
		return s.netopiaCancel(payment)
	case models.PaymentProviderManual:
		return s.manualCancel(payment)
	default:
		return nil, fmt.Errorf("unsupported payment provider: %s", payment.Provider)
	}
}

// GetPaymentsByBooking retrieves all payments for a booking
func (s *PaymentService) GetPaymentsByBooking(bookingID string, userID string) ([]*models.Payment, error) {
	// Validate booking belongs to user
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}
	if booking.ClientID != userID {
		return nil, fmt.Errorf("booking does not belong to user")
	}

	return s.paymentRepo.GetByBookingID(bookingID)
}

// GetPayment retrieves a single payment by ID with authorization check
func (s *PaymentService) GetPayment(paymentID string, userID string) (*models.Payment, error) {
	// Get payment
	payment, err := s.paymentRepo.GetByID(paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	if payment == nil {
		return nil, fmt.Errorf("payment not found")
	}

	// Validate booking belongs to user
	booking, err := s.bookingRepo.GetByID(payment.BookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}
	if booking.ClientID != userID {
		return nil, fmt.Errorf("payment does not belong to user")
	}

	return payment, nil
}

// --- Netopia Payment Provider Integration ---

func (s *PaymentService) netopiaPreauthorize(payment *models.Payment) (*models.Payment, error) {
	// TODO: Integrate with Netopia SDK
	// For now, this is a stub that simulates the flow

	// In production, this would:
	// 1. Call Netopia API to create preauthorization
	// 2. Get transaction ID and order ID from response
	// 3. Store card details (last 4 digits, brand)
	// 4. Return payment URL for 3DS authentication

	// Simulate successful preauthorization
	payment.ProviderTransactionID = sql.NullString{String: fmt.Sprintf("NETOPIA-TXN-%d", time.Now().Unix()), Valid: true}
	payment.ProviderOrderID = sql.NullString{String: fmt.Sprintf("ORDER-%d", time.Now().Unix()), Valid: true}
	payment.Status = models.PaymentStatusAuthorized
	payment.AuthorizedAt = sql.NullTime{Time: time.Now(), Valid: true}

	// Store mock response
	response := map[string]interface{}{
		"status": "authorized",
		"message": "Payment preauthorized successfully (MOCK)",
		"transaction_id": payment.ProviderTransactionID.String,
	}
	payment.ProviderResponse, _ = json.Marshal(response)

	// Save to database
	err := s.paymentRepo.Create(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return payment, nil
}

func (s *PaymentService) netopiaCapture(payment *models.Payment) (*models.Payment, error) {
	// TODO: Integrate with Netopia SDK for capture

	payment.Status = models.PaymentStatusCaptured
	payment.CapturedAt = sql.NullTime{Time: time.Now(), Valid: true}

	response := map[string]interface{}{
		"status": "captured",
		"message": "Payment captured successfully (MOCK)",
	}
	payment.ProviderResponse, _ = json.Marshal(response)

	err := s.paymentRepo.Update(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return payment, nil
}

func (s *PaymentService) netopiaRefund(refundPayment *models.Payment, originalPayment *models.Payment) (*models.Payment, error) {
	// TODO: Integrate with Netopia SDK for refund

	refundPayment.ProviderTransactionID = sql.NullString{String: fmt.Sprintf("NETOPIA-REFUND-%d", time.Now().Unix()), Valid: true}
	refundPayment.Status = models.PaymentStatusRefunded
	refundPayment.RefundedAt = sql.NullTime{Time: time.Now(), Valid: true}

	response := map[string]interface{}{
		"status": "refunded",
		"message": "Payment refunded successfully (MOCK)",
		"original_transaction_id": originalPayment.ProviderTransactionID.String,
	}
	refundPayment.ProviderResponse, _ = json.Marshal(response)

	err := s.paymentRepo.Create(refundPayment)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	return refundPayment, nil
}

func (s *PaymentService) netopiaCancel(payment *models.Payment) (*models.Payment, error) {
	// TODO: Integrate with Netopia SDK for cancellation

	payment.Status = models.PaymentStatusCancelled

	response := map[string]interface{}{
		"status": "cancelled",
		"message": "Preauthorization cancelled successfully (MOCK)",
	}
	payment.ProviderResponse, _ = json.Marshal(response)

	err := s.paymentRepo.Update(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return payment, nil
}

// --- Manual Payment Provider (for testing/admin) ---

func (s *PaymentService) manualPreauthorize(payment *models.Payment) (*models.Payment, error) {
	payment.ProviderTransactionID = sql.NullString{String: fmt.Sprintf("MANUAL-TXN-%d", time.Now().Unix()), Valid: true}
	payment.Status = models.PaymentStatusAuthorized
	payment.AuthorizedAt = sql.NullTime{Time: time.Now(), Valid: true}

	response := map[string]interface{}{
		"status": "authorized",
		"message": "Manual payment authorized",
	}
	payment.ProviderResponse, _ = json.Marshal(response)

	err := s.paymentRepo.Create(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	return payment, nil
}

func (s *PaymentService) manualCapture(payment *models.Payment) (*models.Payment, error) {
	payment.Status = models.PaymentStatusCaptured
	payment.CapturedAt = sql.NullTime{Time: time.Now(), Valid: true}

	response := map[string]interface{}{
		"status": "captured",
		"message": "Manual payment captured",
	}
	payment.ProviderResponse, _ = json.Marshal(response)

	err := s.paymentRepo.Update(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return payment, nil
}

func (s *PaymentService) manualRefund(refundPayment *models.Payment) (*models.Payment, error) {
	refundPayment.ProviderTransactionID = sql.NullString{String: fmt.Sprintf("MANUAL-REFUND-%d", time.Now().Unix()), Valid: true}
	refundPayment.Status = models.PaymentStatusRefunded
	refundPayment.RefundedAt = sql.NullTime{Time: time.Now(), Valid: true}

	response := map[string]interface{}{
		"status": "refunded",
		"message": "Manual payment refunded",
	}
	refundPayment.ProviderResponse, _ = json.Marshal(response)

	err := s.paymentRepo.Create(refundPayment)
	if err != nil {
		return nil, fmt.Errorf("failed to create refund: %w", err)
	}

	return refundPayment, nil
}

func (s *PaymentService) manualCancel(payment *models.Payment) (*models.Payment, error) {
	payment.Status = models.PaymentStatusCancelled

	response := map[string]interface{}{
		"status": "cancelled",
		"message": "Manual payment cancelled",
	}
	payment.ProviderResponse, _ = json.Marshal(response)

	err := s.paymentRepo.Update(payment)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return payment, nil
}
