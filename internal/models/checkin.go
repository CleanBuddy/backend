package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Checkin represents a check-in/check-out record for a booking
type Checkin struct {
	ID                 string
	BookingID          string
	CleanerID          string
	CheckInTime        sql.NullTime
	CheckInLatitude    sql.NullFloat64
	CheckInLongitude   sql.NullFloat64
	CheckOutTime       sql.NullTime
	CheckOutLatitude   sql.NullFloat64
	CheckOutLongitude  sql.NullFloat64
	TotalHoursWorked   sql.NullFloat64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// CheckinRepository handles checkin database operations
type CheckinRepository struct {
	db *sql.DB
}

// NewCheckinRepository creates a new checkin repository
func NewCheckinRepository(db *sql.DB) *CheckinRepository {
	return &CheckinRepository{db: db}
}

// Create creates a new checkin record
func (r *CheckinRepository) Create(checkin *Checkin) error {
	// Generate UUID if not set
	if checkin.ID == "" {
		checkin.ID = uuid.New().String()
	}

	return r.db.QueryRow(`
		INSERT INTO checkins (id, booking_id, cleaner_id, check_in_time, check_in_latitude, check_in_longitude)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at
	`, checkin.ID, checkin.BookingID, checkin.CleanerID, checkin.CheckInTime, checkin.CheckInLatitude, checkin.CheckInLongitude).
		Scan(&checkin.CreatedAt, &checkin.UpdatedAt)
}

// GetByBookingID gets a checkin by booking ID
func (r *CheckinRepository) GetByBookingID(bookingID string) (*Checkin, error) {
	checkin := &Checkin{}
	err := r.db.QueryRow(`
		SELECT id, booking_id, cleaner_id,
		       check_in_time, check_in_latitude, check_in_longitude,
		       check_out_time, check_out_latitude, check_out_longitude,
		       total_hours_worked, created_at, updated_at
		FROM checkins
		WHERE booking_id = $1
	`, bookingID).Scan(
		&checkin.ID, &checkin.BookingID, &checkin.CleanerID,
		&checkin.CheckInTime, &checkin.CheckInLatitude, &checkin.CheckInLongitude,
		&checkin.CheckOutTime, &checkin.CheckOutLatitude, &checkin.CheckOutLongitude,
		&checkin.TotalHoursWorked, &checkin.CreatedAt, &checkin.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return checkin, nil
}

// Update updates a checkin record
func (r *CheckinRepository) Update(checkin *Checkin) error {
	_, err := r.db.Exec(`
		UPDATE checkins
		SET check_in_time = $2, check_in_latitude = $3, check_in_longitude = $4,
		    check_out_time = $5, check_out_latitude = $6, check_out_longitude = $7,
		    total_hours_worked = $8, updated_at = NOW()
		WHERE id = $1
	`, checkin.ID, checkin.CheckInTime, checkin.CheckInLatitude, checkin.CheckInLongitude,
		checkin.CheckOutTime, checkin.CheckOutLatitude, checkin.CheckOutLongitude,
		checkin.TotalHoursWorked)
	return err
}
