package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

// PaymentProvider represents the payment gateway
type PaymentProvider string

const (
	PaymentProviderNetopia PaymentProvider = "NETOPIA"
	PaymentProviderManual  PaymentProvider = "MANUAL"
)

// PaymentType represents the type of payment transaction
type PaymentType string

const (
	PaymentTypePreauthorization PaymentType = "PREAUTHORIZATION"
	PaymentTypeCapture          PaymentType = "CAPTURE"
	PaymentTypeRefund           PaymentType = "REFUND"
	PaymentTypeCancellation     PaymentType = "CANCELLATION"
)

// PaymentStatus represents the current status of a payment
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "PENDING"
	PaymentStatusAuthorized PaymentStatus = "AUTHORIZED"
	PaymentStatusCaptured   PaymentStatus = "CAPTURED"
	PaymentStatusFailed     PaymentStatus = "FAILED"
	PaymentStatusRefunded   PaymentStatus = "REFUNDED"
	PaymentStatusCancelled  PaymentStatus = "CANCELLED"
)

// Payment represents a payment transaction
type Payment struct {
	ID                     string
	BookingID              string
	UserID                 string
	Provider               PaymentProvider
	ProviderTransactionID  sql.NullString
	ProviderOrderID        sql.NullString
	PaymentType            PaymentType
	Status                 PaymentStatus
	Amount                 float64
	Currency               string
	CardLastFour           sql.NullString
	CardBrand              sql.NullString
	ErrorCode              sql.NullString
	ErrorMessage           sql.NullString
	ProviderResponse       json.RawMessage
	AuthorizedAt           sql.NullTime
	CapturedAt             sql.NullTime
	FailedAt               sql.NullTime
	RefundedAt             sql.NullTime
	CreatedAt              time.Time
	UpdatedAt              time.Time
}

// PaymentRepository handles database operations for payments
type PaymentRepository struct {
	db *sql.DB
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(db *sql.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create creates a new payment record
func (r *PaymentRepository) Create(payment *Payment) error {
	query := `
		INSERT INTO payments (
			id, booking_id, user_id, provider, provider_transaction_id, provider_order_id,
			payment_type, status, amount, currency, card_last_four, card_brand,
			error_code, error_message, provider_response,
			authorized_at, captured_at, failed_at, refunded_at
		) VALUES (
			COALESCE(NULLIF($1, ''), gen_random_uuid()::text),
			$2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		) RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(
		query,
		payment.ID,
		payment.BookingID,
		payment.UserID,
		payment.Provider,
		payment.ProviderTransactionID,
		payment.ProviderOrderID,
		payment.PaymentType,
		payment.Status,
		payment.Amount,
		payment.Currency,
		payment.CardLastFour,
		payment.CardBrand,
		payment.ErrorCode,
		payment.ErrorMessage,
		payment.ProviderResponse,
		payment.AuthorizedAt,
		payment.CapturedAt,
		payment.FailedAt,
		payment.RefundedAt,
	).Scan(&payment.ID, &payment.CreatedAt, &payment.UpdatedAt)
}

// GetByID retrieves a payment by ID
func (r *PaymentRepository) GetByID(id string) (*Payment, error) {
	payment := &Payment{}
	query := `
		SELECT
			id, booking_id, user_id, provider, provider_transaction_id, provider_order_id,
			payment_type, status, amount, currency, card_last_four, card_brand,
			error_code, error_message, provider_response,
			authorized_at, captured_at, failed_at, refunded_at,
			created_at, updated_at
		FROM payments
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&payment.ID,
		&payment.BookingID,
		&payment.UserID,
		&payment.Provider,
		&payment.ProviderTransactionID,
		&payment.ProviderOrderID,
		&payment.PaymentType,
		&payment.Status,
		&payment.Amount,
		&payment.Currency,
		&payment.CardLastFour,
		&payment.CardBrand,
		&payment.ErrorCode,
		&payment.ErrorMessage,
		&payment.ProviderResponse,
		&payment.AuthorizedAt,
		&payment.CapturedAt,
		&payment.FailedAt,
		&payment.RefundedAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return payment, nil
}

// GetByBookingID retrieves all payments for a booking
func (r *PaymentRepository) GetByBookingID(bookingID string) ([]*Payment, error) {
	query := `
		SELECT
			id, booking_id, user_id, provider, provider_transaction_id, provider_order_id,
			payment_type, status, amount, currency, card_last_four, card_brand,
			error_code, error_message, provider_response,
			authorized_at, captured_at, failed_at, refunded_at,
			created_at, updated_at
		FROM payments
		WHERE booking_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, bookingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*Payment
	for rows.Next() {
		payment := &Payment{}
		err := rows.Scan(
			&payment.ID,
			&payment.BookingID,
			&payment.UserID,
			&payment.Provider,
			&payment.ProviderTransactionID,
			&payment.ProviderOrderID,
			&payment.PaymentType,
			&payment.Status,
			&payment.Amount,
			&payment.Currency,
			&payment.CardLastFour,
			&payment.CardBrand,
			&payment.ErrorCode,
			&payment.ErrorMessage,
			&payment.ProviderResponse,
			&payment.AuthorizedAt,
			&payment.CapturedAt,
			&payment.FailedAt,
			&payment.RefundedAt,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}

	return payments, rows.Err()
}

// GetByProviderTransactionID retrieves a payment by provider transaction ID
func (r *PaymentRepository) GetByProviderTransactionID(providerTransactionID string) (*Payment, error) {
	payment := &Payment{}
	query := `
		SELECT
			id, booking_id, user_id, provider, provider_transaction_id, provider_order_id,
			payment_type, status, amount, currency, card_last_four, card_brand,
			error_code, error_message, provider_response,
			authorized_at, captured_at, failed_at, refunded_at,
			created_at, updated_at
		FROM payments
		WHERE provider_transaction_id = $1
	`

	err := r.db.QueryRow(query, providerTransactionID).Scan(
		&payment.ID,
		&payment.BookingID,
		&payment.UserID,
		&payment.Provider,
		&payment.ProviderTransactionID,
		&payment.ProviderOrderID,
		&payment.PaymentType,
		&payment.Status,
		&payment.Amount,
		&payment.Currency,
		&payment.CardLastFour,
		&payment.CardBrand,
		&payment.ErrorCode,
		&payment.ErrorMessage,
		&payment.ProviderResponse,
		&payment.AuthorizedAt,
		&payment.CapturedAt,
		&payment.FailedAt,
		&payment.RefundedAt,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return payment, nil
}

// Update updates a payment record
func (r *PaymentRepository) Update(payment *Payment) error {
	query := `
		UPDATE payments SET
			provider_transaction_id = $2,
			provider_order_id = $3,
			status = $4,
			card_last_four = $5,
			card_brand = $6,
			error_code = $7,
			error_message = $8,
			provider_response = $9,
			authorized_at = $10,
			captured_at = $11,
			failed_at = $12,
			refunded_at = $13,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at
	`

	return r.db.QueryRow(
		query,
		payment.ID,
		payment.ProviderTransactionID,
		payment.ProviderOrderID,
		payment.Status,
		payment.CardLastFour,
		payment.CardBrand,
		payment.ErrorCode,
		payment.ErrorMessage,
		payment.ProviderResponse,
		payment.AuthorizedAt,
		payment.CapturedAt,
		payment.FailedAt,
		payment.RefundedAt,
	).Scan(&payment.UpdatedAt)
}
