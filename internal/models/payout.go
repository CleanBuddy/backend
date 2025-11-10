package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const (
	PayoutStatusPending    = "PENDING"
	PayoutStatusProcessing = "PROCESSING"
	PayoutStatusSent       = "SENT"
	PayoutStatusFailed     = "FAILED"
	PayoutStatusCancelled  = "CANCELLED"
)

type Payout struct {
	ID                    string
	CleanerID             string
	PeriodStart           time.Time
	PeriodEnd             time.Time
	Status                string
	TotalBookings         int
	TotalEarnings         float64
	PlatformFees          float64
	NetAmount             float64
	IBAN                  sql.NullString
	TransferReference     sql.NullString
	SettlementInvoiceURL  sql.NullString
	PaidAt                sql.NullTime
	FailedReason          sql.NullString
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

type PayoutLineItem struct {
	ID              string
	PayoutID        string
	BookingID       string
	BookingDate     time.Time
	ServiceType     string
	BookingAmount   float64
	PlatformFeeRate float64
	PlatformFee     float64
	CleanerEarnings float64
	CreatedAt       time.Time
}

type PayoutRepository struct {
	db *sql.DB
}

func NewPayoutRepository(db *sql.DB) *PayoutRepository {
	return &PayoutRepository{db: db}
}

func (r *PayoutRepository) Create(payout *Payout) error {
	if payout.ID == "" {
		payout.ID = uuid.New().String()
	}

	query := `
		INSERT INTO payouts (
			id, cleaner_id, period_start, period_end, status,
			total_bookings, total_earnings, platform_fees, net_amount,
			iban, transfer_reference, settlement_invoice_url,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	return r.db.QueryRow(
		query,
		payout.ID,
		payout.CleanerID,
		payout.PeriodStart,
		payout.PeriodEnd,
		payout.Status,
		payout.TotalBookings,
		payout.TotalEarnings,
		payout.PlatformFees,
		payout.NetAmount,
		payout.IBAN,
		payout.TransferReference,
		payout.SettlementInvoiceURL,
	).Scan(&payout.CreatedAt, &payout.UpdatedAt)
}

func (r *PayoutRepository) GetByID(id string) (*Payout, error) {
	payout := &Payout{}
	query := `
		SELECT id, cleaner_id, period_start, period_end, status,
			   total_bookings, total_earnings, platform_fees, net_amount,
			   iban, transfer_reference, settlement_invoice_url,
			   paid_at, failed_reason, created_at, updated_at
		FROM payouts
		WHERE id = $1
	`
	err := r.db.QueryRow(query, id).Scan(
		&payout.ID,
		&payout.CleanerID,
		&payout.PeriodStart,
		&payout.PeriodEnd,
		&payout.Status,
		&payout.TotalBookings,
		&payout.TotalEarnings,
		&payout.PlatformFees,
		&payout.NetAmount,
		&payout.IBAN,
		&payout.TransferReference,
		&payout.SettlementInvoiceURL,
		&payout.PaidAt,
		&payout.FailedReason,
		&payout.CreatedAt,
		&payout.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return payout, nil
}

func (r *PayoutRepository) GetByCleanerID(cleanerID string, limit, offset int) ([]*Payout, error) {
	query := `
		SELECT id, cleaner_id, period_start, period_end, status,
			   total_bookings, total_earnings, platform_fees, net_amount,
			   iban, transfer_reference, settlement_invoice_url,
			   paid_at, failed_reason, created_at, updated_at
		FROM payouts
		WHERE cleaner_id = $1
		ORDER BY period_start DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(query, cleanerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payouts []*Payout
	for rows.Next() {
		payout := &Payout{}
		if err := rows.Scan(
			&payout.ID,
			&payout.CleanerID,
			&payout.PeriodStart,
			&payout.PeriodEnd,
			&payout.Status,
			&payout.TotalBookings,
			&payout.TotalEarnings,
			&payout.PlatformFees,
			&payout.NetAmount,
			&payout.IBAN,
			&payout.TransferReference,
			&payout.SettlementInvoiceURL,
			&payout.PaidAt,
			&payout.FailedReason,
			&payout.CreatedAt,
			&payout.UpdatedAt,
		); err != nil {
			return nil, err
		}
		payouts = append(payouts, payout)
	}
	return payouts, nil
}

func (r *PayoutRepository) GetByStatus(status string) ([]*Payout, error) {
	query := `
		SELECT id, cleaner_id, period_start, period_end, status,
			   total_bookings, total_earnings, platform_fees, net_amount,
			   iban, transfer_reference, settlement_invoice_url,
			   paid_at, failed_reason, created_at, updated_at
		FROM payouts
		WHERE status = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payouts []*Payout
	for rows.Next() {
		payout := &Payout{}
		if err := rows.Scan(
			&payout.ID,
			&payout.CleanerID,
			&payout.PeriodStart,
			&payout.PeriodEnd,
			&payout.Status,
			&payout.TotalBookings,
			&payout.TotalEarnings,
			&payout.PlatformFees,
			&payout.NetAmount,
			&payout.IBAN,
			&payout.TransferReference,
			&payout.SettlementInvoiceURL,
			&payout.PaidAt,
			&payout.FailedReason,
			&payout.CreatedAt,
			&payout.UpdatedAt,
		); err != nil {
			return nil, err
		}
		payouts = append(payouts, payout)
	}
	return payouts, nil
}

// GetAll returns all payouts with pagination
func (r *PayoutRepository) GetAll(limit, offset int) ([]*Payout, error) {
	query := `
		SELECT id, cleaner_id, period_start, period_end, status,
			total_bookings, total_earnings, platform_fees, net_amount,
			iban, transfer_reference, settlement_invoice_url,
			paid_at, failed_reason, created_at, updated_at
		FROM payouts
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payouts []*Payout
	for rows.Next() {
		payout := &Payout{}
		if err := rows.Scan(
			&payout.ID,
			&payout.CleanerID,
			&payout.PeriodStart,
			&payout.PeriodEnd,
			&payout.Status,
			&payout.TotalBookings,
			&payout.TotalEarnings,
			&payout.PlatformFees,
			&payout.NetAmount,
			&payout.IBAN,
			&payout.TransferReference,
			&payout.SettlementInvoiceURL,
			&payout.PaidAt,
			&payout.FailedReason,
			&payout.CreatedAt,
			&payout.UpdatedAt,
		); err != nil {
			return nil, err
		}
		payouts = append(payouts, payout)
	}
	return payouts, nil
}

func (r *PayoutRepository) Update(payout *Payout) error {
	query := `
		UPDATE payouts
		SET status = $2, iban = $3, transfer_reference = $4,
			settlement_invoice_url = $5, paid_at = $6, failed_reason = $7,
			updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`
	return r.db.QueryRow(
		query,
		payout.ID,
		payout.Status,
		payout.IBAN,
		payout.TransferReference,
		payout.SettlementInvoiceURL,
		payout.PaidAt,
		payout.FailedReason,
	).Scan(&payout.UpdatedAt)
}

// PayoutLineItem Repository
type PayoutLineItemRepository struct {
	db *sql.DB
}

func NewPayoutLineItemRepository(db *sql.DB) *PayoutLineItemRepository {
	return &PayoutLineItemRepository{db: db}
}

func (r *PayoutLineItemRepository) Create(item *PayoutLineItem) error {
	if item.ID == "" {
		item.ID = uuid.New().String()
	}

	query := `
		INSERT INTO payout_line_items (
			id, payout_id, booking_id, booking_date, service_type,
			booking_amount, platform_fee_rate, platform_fee, cleaner_earnings,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
		RETURNING created_at
	`
	return r.db.QueryRow(
		query,
		item.ID,
		item.PayoutID,
		item.BookingID,
		item.BookingDate,
		item.ServiceType,
		item.BookingAmount,
		item.PlatformFeeRate,
		item.PlatformFee,
		item.CleanerEarnings,
	).Scan(&item.CreatedAt)
}

func (r *PayoutLineItemRepository) GetByPayoutID(payoutID string) ([]*PayoutLineItem, error) {
	query := `
		SELECT id, payout_id, booking_id, booking_date, service_type,
			   booking_amount, platform_fee_rate, platform_fee, cleaner_earnings,
			   created_at
		FROM payout_line_items
		WHERE payout_id = $1
		ORDER BY booking_date ASC
	`
	rows, err := r.db.Query(query, payoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []*PayoutLineItem
	for rows.Next() {
		item := &PayoutLineItem{}
		if err := rows.Scan(
			&item.ID,
			&item.PayoutID,
			&item.BookingID,
			&item.BookingDate,
			&item.ServiceType,
			&item.BookingAmount,
			&item.PlatformFeeRate,
			&item.PlatformFee,
			&item.CleanerEarnings,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
