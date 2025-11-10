package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Dispute types
const (
	DisputeTypeQualityIssue = "QUALITY_ISSUE"
	DisputeTypeDamage       = "DAMAGE"
	DisputeTypeNoShow       = "NO_SHOW"
	DisputeTypePricing      = "PRICING"
	DisputeTypeOther        = "OTHER"
)

// Dispute statuses
const (
	DisputeStatusOpen        = "OPEN"
	DisputeStatusUnderReview = "UNDER_REVIEW"
	DisputeStatusResolved    = "RESOLVED"
	DisputeStatusClosed      = "CLOSED"
)

// Dispute resolution types
const (
	DisputeResolutionPartialRefund = "PARTIAL_REFUND"
	DisputeResolutionFullRefund    = "FULL_REFUND"
	DisputeResolutionReclean       = "RECLEAN"
	DisputeResolutionRejected      = "REJECTED"
)

// Dispute represents a customer dispute for a completed booking
type Dispute struct {
	ID          string
	BookingID   string
	CreatedBy   string
	AssignedTo  sql.NullString
	DisputeType string
	Status      string
	Description string

	// Resolution
	ResolutionType  sql.NullString
	ResolutionNotes sql.NullString
	RefundAmount    sql.NullFloat64
	ResolvedAt      sql.NullTime
	ResolvedBy      sql.NullString

	// Cleaner response
	CleanerResponse     sql.NullString
	CleanerRespondedAt  sql.NullTime

	CreatedAt time.Time
	UpdatedAt time.Time
}

// DisputeRepository handles dispute database operations
type DisputeRepository struct {
	db *sql.DB
}

// NewDisputeRepository creates a new dispute repository
func NewDisputeRepository(db *sql.DB) *DisputeRepository {
	return &DisputeRepository{db: db}
}

// Create creates a new dispute
func (r *DisputeRepository) Create(dispute *Dispute) error {
	// Generate UUID if not set
	if dispute.ID == "" {
		dispute.ID = uuid.New().String()
	}

	return r.db.QueryRow(`
		INSERT INTO disputes (
			id, booking_id, created_by, dispute_type, status, description
		) VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`, dispute.ID, dispute.BookingID, dispute.CreatedBy, dispute.DisputeType,
	   dispute.Status, dispute.Description).
		Scan(&dispute.CreatedAt, &dispute.UpdatedAt)
}

// GetByID gets a dispute by ID
func (r *DisputeRepository) GetByID(id string) (*Dispute, error) {
	dispute := &Dispute{}
	err := r.db.QueryRow(`
		SELECT id, booking_id, created_by, assigned_to, dispute_type, status, description,
		       resolution_type, resolution_notes, refund_amount, resolved_at, resolved_by,
		       cleaner_response, cleaner_responded_at, created_at, updated_at
		FROM disputes
		WHERE id = $1
	`, id).Scan(
		&dispute.ID, &dispute.BookingID, &dispute.CreatedBy, &dispute.AssignedTo,
		&dispute.DisputeType, &dispute.Status, &dispute.Description,
		&dispute.ResolutionType, &dispute.ResolutionNotes, &dispute.RefundAmount,
		&dispute.ResolvedAt, &dispute.ResolvedBy,
		&dispute.CleanerResponse, &dispute.CleanerRespondedAt,
		&dispute.CreatedAt, &dispute.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return dispute, nil
}

// GetByBookingID gets a dispute by booking ID
func (r *DisputeRepository) GetByBookingID(bookingID string) (*Dispute, error) {
	dispute := &Dispute{}
	err := r.db.QueryRow(`
		SELECT id, booking_id, created_by, assigned_to, dispute_type, status, description,
		       resolution_type, resolution_notes, refund_amount, resolved_at, resolved_by,
		       cleaner_response, cleaner_responded_at, created_at, updated_at
		FROM disputes
		WHERE booking_id = $1
	`, bookingID).Scan(
		&dispute.ID, &dispute.BookingID, &dispute.CreatedBy, &dispute.AssignedTo,
		&dispute.DisputeType, &dispute.Status, &dispute.Description,
		&dispute.ResolutionType, &dispute.ResolutionNotes, &dispute.RefundAmount,
		&dispute.ResolvedAt, &dispute.ResolvedBy,
		&dispute.CleanerResponse, &dispute.CleanerRespondedAt,
		&dispute.CreatedAt, &dispute.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return dispute, nil
}

// Update updates a dispute
func (r *DisputeRepository) Update(dispute *Dispute) error {
	_, err := r.db.Exec(`
		UPDATE disputes
		SET assigned_to = $2, status = $3, resolution_type = $4, resolution_notes = $5,
		    refund_amount = $6, resolved_at = $7, resolved_by = $8,
		    cleaner_response = $9, cleaner_responded_at = $10, updated_at = NOW()
		WHERE id = $1
	`, dispute.ID, dispute.AssignedTo, dispute.Status, dispute.ResolutionType,
	   dispute.ResolutionNotes, dispute.RefundAmount, dispute.ResolvedAt, dispute.ResolvedBy,
	   dispute.CleanerResponse, dispute.CleanerRespondedAt)
	return err
}

// GetAllByStatus gets all disputes by status
func (r *DisputeRepository) GetAllByStatus(status string, limit int) ([]*Dispute, error) {
	rows, err := r.db.Query(`
		SELECT id, booking_id, created_by, assigned_to, dispute_type, status, description,
		       resolution_type, resolution_notes, refund_amount, resolved_at, resolved_by,
		       cleaner_response, cleaner_responded_at, created_at, updated_at
		FROM disputes
		WHERE status = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var disputes []*Dispute
	for rows.Next() {
		dispute := &Dispute{}
		err := rows.Scan(
			&dispute.ID, &dispute.BookingID, &dispute.CreatedBy, &dispute.AssignedTo,
			&dispute.DisputeType, &dispute.Status, &dispute.Description,
			&dispute.ResolutionType, &dispute.ResolutionNotes, &dispute.RefundAmount,
			&dispute.ResolvedAt, &dispute.ResolvedBy,
			&dispute.CleanerResponse, &dispute.CleanerRespondedAt,
			&dispute.CreatedAt, &dispute.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		disputes = append(disputes, dispute)
	}

	return disputes, nil
}
