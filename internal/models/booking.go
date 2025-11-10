package models

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

// BookingStatus represents booking state
type BookingStatus string

const (
	BookingStatusPending        BookingStatus = "PENDING"
	BookingStatusConfirmed      BookingStatus = "CONFIRMED"
	BookingStatusInProgress     BookingStatus = "IN_PROGRESS"
	BookingStatusCompleted      BookingStatus = "COMPLETED"
	BookingStatusCancelled      BookingStatus = "CANCELLED"
	BookingStatusDisputed       BookingStatus = "DISPUTED"
	BookingStatusRefunded       BookingStatus = "REFUNDED"
	BookingStatusNoShowClient   BookingStatus = "NO_SHOW_CLIENT"
	BookingStatusNoShowCleaner  BookingStatus = "NO_SHOW_CLEANER"
)

// ServiceType represents type of cleaning service
type ServiceType string

const (
	ServiceTypeStandard       ServiceType = "STANDARD"
	ServiceTypeDeepCleaning   ServiceType = "DEEP_CLEANING"
	ServiceTypeOffice         ServiceType = "OFFICE"
	ServiceTypePostRenovation ServiceType = "POST_RENOVATION"
	ServiceTypeMoveInOut      ServiceType = "MOVE_IN_OUT"
)

// Booking represents a cleaning service booking
type Booking struct {
	ID              string
	ReservationCode sql.NullString // Human-friendly unique identifier (e.g., CB-2024-A1B2C3)
	ClientID        string
	CleanerID       sql.NullString
	AddressID       string

	// Service details
	ServiceType     ServiceType
	AreaSqm         sql.NullInt32
	EstimatedHours  int
	Frequency       sql.NullString // one_time, weekly, biweekly, monthly

	// Scheduling
	ScheduledDate     time.Time
	ScheduledTime     time.Time
	EstimatedEndTime  sql.NullTime
	TimePreferences   sql.NullString // JSONB: Client's available time slots

	// Add-ons
	IncludesDeepCleaning     bool
	IncludesWindows          bool
	IncludesCarpetCleaning   bool
	NumberOfWindows          int
	CarpetAreaSqm            int
	IncludesFridgeCleaning   bool
	IncludesOvenCleaning     bool
	IncludesBalconyCleaning  bool

	// Special instructions & supplies
	SpecialInstructions sql.NullString
	AccessInstructions  sql.NullString
	Supplies            sql.NullString // client_provides or cleaner_provides (NOT NULL)

	// Pricing
	BasePrice       float64
	AddonsPrice     float64
	TotalPrice      float64
	PlatformFee     float64
	CleanerPayout   float64
	DiscountApplied float64

	// State
	Status BookingStatus

	// Status timestamps
	ConfirmedAt sql.NullTime
	StartedAt   sql.NullTime
	CompletedAt sql.NullTime
	CancelledAt sql.NullTime

	// Cancellation
	CancellationReason sql.NullString
	CancelledBy        sql.NullString

	// Ratings
	ClientRating  sql.NullInt32
	ClientReview  sql.NullString
	CleanerRating sql.NullInt32
	CleanerReview sql.NullString

	// Timestamps
	CreatedAt time.Time
	UpdatedAt time.Time
}

// BookingRepository handles booking database operations
type BookingRepository struct {
	db *sql.DB
}

// NewBookingRepository creates a new booking repository
func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

// Create creates a new booking
func (r *BookingRepository) Create(booking *Booking) error {
	return r.db.QueryRow(`
		INSERT INTO bookings (
			client_id, address_id, service_type, area_sqm, estimated_hours, frequency,
			scheduled_date, scheduled_time, time_preferences,
			includes_deep_cleaning, includes_windows, includes_carpet_cleaning,
			number_of_windows, carpet_area_sqm,
			includes_fridge_cleaning, includes_oven_cleaning, includes_balcony_cleaning,
			special_instructions, access_instructions, supplies,
			base_price, addons_price, total_price, platform_fee, cleaner_payout, discount_applied,
			status, reservation_code
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28)
		RETURNING id, created_at, updated_at
	`, booking.ClientID, booking.AddressID, booking.ServiceType, booking.AreaSqm, booking.EstimatedHours, booking.Frequency,
		booking.ScheduledDate, booking.ScheduledTime, booking.TimePreferences,
		booking.IncludesDeepCleaning, booking.IncludesWindows, booking.IncludesCarpetCleaning,
		booking.NumberOfWindows, booking.CarpetAreaSqm,
		booking.IncludesFridgeCleaning, booking.IncludesOvenCleaning, booking.IncludesBalconyCleaning,
		booking.SpecialInstructions, booking.AccessInstructions, booking.Supplies,
		booking.BasePrice, booking.AddonsPrice, booking.TotalPrice, booking.PlatformFee, booking.CleanerPayout, booking.DiscountApplied,
		booking.Status, booking.ReservationCode).
		Scan(&booking.ID, &booking.CreatedAt, &booking.UpdatedAt)
}

// GetByID finds a booking by ID
func (r *BookingRepository) GetByID(id string) (*Booking, error) {
	booking := &Booking{}
	err := r.db.QueryRow(`
		SELECT id, client_id, cleaner_id, address_id,
		       service_type, area_sqm, estimated_hours, frequency,
		       scheduled_date, scheduled_time, estimated_end_time, time_preferences,
		       includes_deep_cleaning, includes_windows, includes_carpet_cleaning,
		       number_of_windows, carpet_area_sqm,
		       includes_fridge_cleaning, includes_oven_cleaning, includes_balcony_cleaning,
		       special_instructions, access_instructions, supplies,
		       base_price, addons_price, total_price, platform_fee, cleaner_payout, discount_applied,
		       status, reservation_code,
		       confirmed_at, started_at, completed_at, cancelled_at,
		       cancellation_reason, cancelled_by,
		       client_rating, client_review, cleaner_rating, cleaner_review,
		       created_at, updated_at
		FROM bookings
		WHERE id = $1
	`, id).Scan(
		&booking.ID, &booking.ClientID, &booking.CleanerID, &booking.AddressID,
		&booking.ServiceType, &booking.AreaSqm, &booking.EstimatedHours, &booking.Frequency,
		&booking.ScheduledDate, &booking.ScheduledTime, &booking.EstimatedEndTime, &booking.TimePreferences,
		&booking.IncludesDeepCleaning, &booking.IncludesWindows, &booking.IncludesCarpetCleaning,
		&booking.NumberOfWindows, &booking.CarpetAreaSqm,
		&booking.IncludesFridgeCleaning, &booking.IncludesOvenCleaning, &booking.IncludesBalconyCleaning,
		&booking.SpecialInstructions, &booking.AccessInstructions, &booking.Supplies,
		&booking.BasePrice, &booking.AddonsPrice, &booking.TotalPrice, &booking.PlatformFee, &booking.CleanerPayout, &booking.DiscountApplied,
		&booking.Status, &booking.ReservationCode,
		&booking.ConfirmedAt, &booking.StartedAt, &booking.CompletedAt, &booking.CancelledAt,
		&booking.CancellationReason, &booking.CancelledBy,
		&booking.ClientRating, &booking.ClientReview, &booking.CleanerRating, &booking.CleanerReview,
		&booking.CreatedAt, &booking.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return booking, nil
}

// GetByClientID returns all bookings for a client
func (r *BookingRepository) GetByClientID(clientID string) ([]*Booking, error) {
	rows, err := r.db.Query(`
		SELECT id, client_id, cleaner_id, address_id,
		       service_type, area_sqm, estimated_hours,
		       scheduled_date, scheduled_time, estimated_end_time,
		       time_preferences,
		       includes_deep_cleaning, includes_windows, includes_carpet_cleaning,
		       number_of_windows, carpet_area_sqm,
		       includes_fridge_cleaning, includes_oven_cleaning, includes_balcony_cleaning,
		       special_instructions, access_instructions,
		       base_price, addons_price, total_price, platform_fee, cleaner_payout, discount_applied,
		       status,
		       confirmed_at, started_at, completed_at, cancelled_at,
		       cancellation_reason, cancelled_by,
		       client_rating, client_review, cleaner_rating, cleaner_review,
		       created_at, updated_at
		FROM bookings
		WHERE client_id = $1
		ORDER BY scheduled_date DESC, scheduled_time DESC
	`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := []*Booking{}
	for rows.Next() {
		booking := &Booking{}
		err := rows.Scan(
			&booking.ID, &booking.ClientID, &booking.CleanerID, &booking.AddressID,
			&booking.ServiceType, &booking.AreaSqm, &booking.EstimatedHours,
			&booking.ScheduledDate, &booking.ScheduledTime, &booking.EstimatedEndTime,
			&booking.TimePreferences,
			&booking.IncludesDeepCleaning, &booking.IncludesWindows, &booking.IncludesCarpetCleaning,
			&booking.NumberOfWindows, &booking.CarpetAreaSqm,
			&booking.IncludesFridgeCleaning, &booking.IncludesOvenCleaning, &booking.IncludesBalconyCleaning,
			&booking.SpecialInstructions, &booking.AccessInstructions,
			&booking.BasePrice, &booking.AddonsPrice, &booking.TotalPrice, &booking.PlatformFee, &booking.CleanerPayout, &booking.DiscountApplied,
			&booking.Status,
			&booking.ConfirmedAt, &booking.StartedAt, &booking.CompletedAt, &booking.CancelledAt,
			&booking.CancellationReason, &booking.CancelledBy,
			&booking.ClientRating, &booking.ClientReview, &booking.CleanerRating, &booking.CleanerReview,
			&booking.CreatedAt, &booking.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	return bookings, rows.Err()
}

// Update updates a booking
func (r *BookingRepository) Update(booking *Booking) error {
	_, err := r.db.Exec(`
		UPDATE bookings
		SET cleaner_id = $2, service_type = $3, area_sqm = $4, estimated_hours = $5,
		    scheduled_date = $6, scheduled_time = $7, estimated_end_time = $8,
		    includes_deep_cleaning = $9, includes_windows = $10, includes_carpet_cleaning = $11,
		    number_of_windows = $12, carpet_area_sqm = $13,
		    special_instructions = $14, access_instructions = $15,
		    base_price = $16, addons_price = $17, total_price = $18, platform_fee = $19, cleaner_payout = $20, discount_applied = $21,
		    status = $22,
		    confirmed_at = $23, started_at = $24, completed_at = $25, cancelled_at = $26,
		    cancellation_reason = $27, cancelled_by = $28,
		    client_rating = $29, client_review = $30, cleaner_rating = $31, cleaner_review = $32
		WHERE id = $1
	`, booking.ID, booking.CleanerID, booking.ServiceType, booking.AreaSqm, booking.EstimatedHours,
		booking.ScheduledDate, booking.ScheduledTime, booking.EstimatedEndTime,
		booking.IncludesDeepCleaning, booking.IncludesWindows, booking.IncludesCarpetCleaning,
		booking.NumberOfWindows, booking.CarpetAreaSqm,
		booking.SpecialInstructions, booking.AccessInstructions,
		booking.BasePrice, booking.AddonsPrice, booking.TotalPrice, booking.PlatformFee, booking.CleanerPayout, booking.DiscountApplied,
		booking.Status,
		booking.ConfirmedAt, booking.StartedAt, booking.CompletedAt, booking.CancelledAt,
		booking.CancellationReason, booking.CancelledBy,
		booking.ClientRating, booking.ClientReview, booking.CleanerRating, booking.CleanerReview)
	return err
}

// GetPendingBookings returns all bookings waiting for cleaner assignment
func (r *BookingRepository) GetPendingBookings() ([]*Booking, error) {
	rows, err := r.db.Query(`
		SELECT id, client_id, cleaner_id, address_id,
		       service_type, area_sqm, estimated_hours,
		       scheduled_date, scheduled_time, estimated_end_time,
		       time_preferences,
		       includes_deep_cleaning, includes_windows, includes_carpet_cleaning,
		       number_of_windows, carpet_area_sqm,
		       special_instructions, access_instructions,
		       base_price, addons_price, total_price, platform_fee, cleaner_payout, discount_applied,
		       status,
		       confirmed_at, started_at, completed_at, cancelled_at,
		       cancellation_reason, cancelled_by,
		       client_rating, client_review, cleaner_rating, cleaner_review,
		       created_at, updated_at
		FROM bookings
		WHERE status = 'PENDING'
		ORDER BY created_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := []*Booking{}
	for rows.Next() {
		booking := &Booking{}
		err := rows.Scan(
			&booking.ID, &booking.ClientID, &booking.CleanerID, &booking.AddressID,
			&booking.ServiceType, &booking.AreaSqm, &booking.EstimatedHours,
			&booking.ScheduledDate, &booking.ScheduledTime, &booking.EstimatedEndTime,
			&booking.TimePreferences,
			&booking.IncludesDeepCleaning, &booking.IncludesWindows, &booking.IncludesCarpetCleaning,
			&booking.NumberOfWindows, &booking.CarpetAreaSqm,
			&booking.IncludesFridgeCleaning, &booking.IncludesOvenCleaning, &booking.IncludesBalconyCleaning,
			&booking.SpecialInstructions, &booking.AccessInstructions,
			&booking.BasePrice, &booking.AddonsPrice, &booking.TotalPrice, &booking.PlatformFee, &booking.CleanerPayout, &booking.DiscountApplied,
			&booking.Status,
			&booking.ConfirmedAt, &booking.StartedAt, &booking.CompletedAt, &booking.CancelledAt,
			&booking.CancellationReason, &booking.CancelledBy,
			&booking.ClientRating, &booking.ClientReview, &booking.CleanerRating, &booking.CleanerReview,
			&booking.CreatedAt, &booking.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	return bookings, rows.Err()
}

// GetAvailableJobs returns bookings available for cleaners to accept
func (r *BookingRepository) GetAvailableJobs(city string, limit, offset int) ([]*Booking, error) {
	query := `
		SELECT b.id, b.client_id, b.cleaner_id, b.address_id,
		       b.service_type, b.area_sqm, b.estimated_hours,
		       b.scheduled_date, b.scheduled_time, b.estimated_end_time,
		       b.time_preferences,
		       b.includes_deep_cleaning, b.includes_windows, b.includes_carpet_cleaning,
		       b.number_of_windows, b.carpet_area_sqm,
		       b.special_instructions, b.access_instructions,
		       b.base_price, b.addons_price, b.total_price, b.platform_fee, b.cleaner_payout, b.discount_applied,
		       b.status,
		       b.confirmed_at, b.started_at, b.completed_at, b.cancelled_at,
		       b.cancellation_reason, b.cancelled_by,
		       b.client_rating, b.client_review, b.cleaner_rating, b.cleaner_review,
		       b.created_at, b.updated_at
		FROM bookings b
		JOIN addresses a ON b.address_id = a.id
		WHERE b.cleaner_id IS NULL
		  AND b.status = 'PENDING'
	`

	args := []interface{}{}
	argIndex := 1

	if city != "" {
		query += fmt.Sprintf(` AND a.city = $%d`, argIndex)
		args = append(args, city)
		argIndex++
	}

	query += ` ORDER BY b.scheduled_date ASC, b.created_at ASC`

	if limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, argIndex)
		args = append(args, limit)
		argIndex++
	}

	if offset > 0 {
		query += fmt.Sprintf(` OFFSET $%d`, argIndex)
		args = append(args, offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := []*Booking{}
	for rows.Next() {
		booking := &Booking{}
		err := rows.Scan(
			&booking.ID, &booking.ClientID, &booking.CleanerID, &booking.AddressID,
			&booking.ServiceType, &booking.AreaSqm, &booking.EstimatedHours,
			&booking.ScheduledDate, &booking.ScheduledTime, &booking.EstimatedEndTime,
			&booking.TimePreferences,
			&booking.IncludesDeepCleaning, &booking.IncludesWindows, &booking.IncludesCarpetCleaning,
			&booking.NumberOfWindows, &booking.CarpetAreaSqm,
			&booking.IncludesFridgeCleaning, &booking.IncludesOvenCleaning, &booking.IncludesBalconyCleaning,
			&booking.SpecialInstructions, &booking.AccessInstructions,
			&booking.BasePrice, &booking.AddonsPrice, &booking.TotalPrice, &booking.PlatformFee, &booking.CleanerPayout, &booking.DiscountApplied,
			&booking.Status,
			&booking.ConfirmedAt, &booking.StartedAt, &booking.CompletedAt, &booking.CancelledAt,
			&booking.CancellationReason, &booking.CancelledBy,
			&booking.ClientRating, &booking.ClientReview, &booking.CleanerRating, &booking.CleanerReview,
			&booking.CreatedAt, &booking.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	return bookings, rows.Err()
}

// GetByClientIDFiltered returns bookings for a client with filter
func (r *BookingRepository) GetByClientIDFiltered(clientID string, filter string) ([]*Booking, error) {
	query := `
		SELECT id, client_id, cleaner_id, address_id,
		       service_type, area_sqm, estimated_hours,
		       scheduled_date, scheduled_time, estimated_end_time,
		       time_preferences,
		       includes_deep_cleaning, includes_windows, includes_carpet_cleaning,
		       number_of_windows, carpet_area_sqm,
		       includes_fridge_cleaning, includes_oven_cleaning, includes_balcony_cleaning,
		       special_instructions, access_instructions,
		       base_price, addons_price, total_price, platform_fee, cleaner_payout, discount_applied,
		       status,
		       confirmed_at, started_at, completed_at, cancelled_at,
		       cancellation_reason, cancelled_by,
		       client_rating, client_review, cleaner_rating, cleaner_review,
		       created_at, updated_at
		FROM bookings
		WHERE client_id = $1
	`

	switch filter {
	case "UPCOMING":
		query += ` AND status IN ('PENDING', 'CONFIRMED', 'IN_PROGRESS') AND scheduled_date >= CURRENT_DATE`
	case "PAST":
		query += ` AND (status IN ('COMPLETED', 'CANCELLED', 'NO_SHOW_CLIENT', 'NO_SHOW_CLEANER') OR scheduled_date < CURRENT_DATE)`
	case "ACTIVE":
		query += ` AND status = 'IN_PROGRESS'`
	// "ALL" or default: no additional filter
	}

	query += ` ORDER BY scheduled_date DESC, scheduled_time DESC`

	rows, err := r.db.Query(query, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bookings := []*Booking{}
	for rows.Next() {
		booking := &Booking{}
		err := rows.Scan(
			&booking.ID, &booking.ClientID, &booking.CleanerID, &booking.AddressID,
			&booking.ServiceType, &booking.AreaSqm, &booking.EstimatedHours,
			&booking.ScheduledDate, &booking.ScheduledTime, &booking.EstimatedEndTime,
			&booking.TimePreferences,
			&booking.IncludesDeepCleaning, &booking.IncludesWindows, &booking.IncludesCarpetCleaning,
			&booking.NumberOfWindows, &booking.CarpetAreaSqm,
			&booking.IncludesFridgeCleaning, &booking.IncludesOvenCleaning, &booking.IncludesBalconyCleaning,
			&booking.SpecialInstructions, &booking.AccessInstructions,
			&booking.BasePrice, &booking.AddonsPrice, &booking.TotalPrice, &booking.PlatformFee, &booking.CleanerPayout, &booking.DiscountApplied,
			&booking.Status,
			&booking.ConfirmedAt, &booking.StartedAt, &booking.CompletedAt, &booking.CancelledAt,
			&booking.CancellationReason, &booking.CancelledBy,
			&booking.ClientRating, &booking.ClientReview, &booking.CleanerRating, &booking.CleanerReview,
			&booking.CreatedAt, &booking.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	return bookings, rows.Err()
}

// GetCompletedBookingsByPeriod returns completed bookings within a date range
func (r *BookingRepository) GetCompletedBookingsByPeriod(startDate, endDate time.Time) ([]*Booking, error) {
	query := `
		SELECT id, client_id, address_id, cleaner_id,
			   service_type, area_sqm, estimated_hours,
			   scheduled_date, scheduled_time, estimated_end_time,
			   includes_deep_cleaning, includes_windows, includes_carpet_cleaning,
			   number_of_windows, carpet_area_sqm,
			   special_instructions, access_instructions,
			   base_price, addons_price, total_price, platform_fee, cleaner_payout, discount_applied,
			   status,
			   confirmed_at, started_at, completed_at, cancelled_at,
			   cancellation_reason, cancelled_by,
			   client_rating, client_review, cleaner_rating, cleaner_review,
			   created_at, updated_at
		FROM bookings
		WHERE status = $1 AND completed_at >= $2 AND completed_at <= $3
		ORDER BY completed_at ASC
	`
	rows, err := r.db.Query(query, BookingStatusCompleted, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*Booking
	for rows.Next() {
		booking := &Booking{}
		err := rows.Scan(
			&booking.ID, &booking.ClientID, &booking.AddressID, &booking.CleanerID, 
			&booking.ServiceType, &booking.AreaSqm, &booking.EstimatedHours,
			&booking.ScheduledDate, &booking.ScheduledTime, &booking.EstimatedEndTime,
			&booking.IncludesDeepCleaning, &booking.IncludesWindows, &booking.IncludesCarpetCleaning,
			&booking.NumberOfWindows, &booking.CarpetAreaSqm,
			&booking.SpecialInstructions, &booking.AccessInstructions,
			&booking.BasePrice, &booking.AddonsPrice, &booking.TotalPrice, &booking.PlatformFee, &booking.CleanerPayout, &booking.DiscountApplied,
			&booking.Status,
			&booking.ConfirmedAt, &booking.StartedAt, &booking.CompletedAt, &booking.CancelledAt,
			&booking.CancellationReason, &booking.CancelledBy,
			&booking.ClientRating, &booking.ClientReview, &booking.CleanerRating, &booking.CleanerReview,
			&booking.CreatedAt, &booking.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	return bookings, rows.Err()
}

// GetCompletedBookingsByClientBeforeDate gets all completed bookings for a client before a specific date
// Used to check if a customer is a repeat customer
func (r *BookingRepository) GetCompletedBookingsByClientBeforeDate(clientID string, beforeDate time.Time) ([]*Booking, error) {
	query := `
		SELECT id, client_id, address_id, cleaner_id,
			   service_type, area_sqm, estimated_hours,
			   scheduled_date, scheduled_time, estimated_end_time,
			   includes_deep_cleaning, includes_windows, includes_carpet_cleaning,
			   number_of_windows, carpet_area_sqm,
			   special_instructions, access_instructions,
			   base_price, addons_price, total_price, platform_fee, cleaner_payout, discount_applied,
			   status,
			   confirmed_at, started_at, completed_at, cancelled_at,
			   cancellation_reason, cancelled_by,
			   client_rating, client_review, cleaner_rating, cleaner_review,
			   created_at, updated_at
		FROM bookings
		WHERE client_id = $1 AND status = $2 AND completed_at < $3
		ORDER BY completed_at ASC
	`
	rows, err := r.db.Query(query, clientID, BookingStatusCompleted, beforeDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*Booking
	for rows.Next() {
		booking := &Booking{}
		err := rows.Scan(
			&booking.ID, &booking.ClientID, &booking.AddressID, &booking.CleanerID,
			&booking.ServiceType, &booking.AreaSqm, &booking.EstimatedHours,
			&booking.ScheduledDate, &booking.ScheduledTime, &booking.EstimatedEndTime,
			&booking.IncludesDeepCleaning, &booking.IncludesWindows, &booking.IncludesCarpetCleaning,
			&booking.NumberOfWindows, &booking.CarpetAreaSqm,
			&booking.SpecialInstructions, &booking.AccessInstructions,
			&booking.BasePrice, &booking.AddonsPrice, &booking.TotalPrice, &booking.PlatformFee, &booking.CleanerPayout, &booking.DiscountApplied,
			&booking.Status,
			&booking.ConfirmedAt, &booking.StartedAt, &booking.CompletedAt, &booking.CancelledAt,
			&booking.CancellationReason, &booking.CancelledBy,
			&booking.ClientRating, &booking.ClientReview, &booking.CleanerRating, &booking.CleanerReview,
			&booking.CreatedAt, &booking.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}

	return bookings, rows.Err()
}

// GetByCleanerIDs retrieves all bookings for multiple cleaners with optional filter
func (r *BookingRepository) GetByCleanerIDs(cleanerIDs []string, filter *string) ([]*Booking, error) {
	if len(cleanerIDs) == 0 {
		return []*Booking{}, nil
	}

	// Build the query with placeholders for the cleaner IDs
	query := `
		SELECT id, client_id, cleaner_id, address_id,
		       service_type, area_sqm, estimated_hours, scheduled_date, scheduled_time, estimated_end_time,
		       time_preferences,
		       includes_deep_cleaning, includes_windows, includes_carpet_cleaning,
		       number_of_windows, carpet_area_sqm,
		       includes_fridge_cleaning, includes_oven_cleaning, includes_balcony_cleaning,
		       special_instructions, access_instructions,
		       base_price, addons_price, total_price, platform_fee, cleaner_payout, discount_applied,
		       status, confirmed_at, started_at, completed_at, cancelled_at,
		       cancellation_reason, cancelled_by,
		       client_rating, client_review, cleaner_rating, cleaner_review,
		       created_at, updated_at
		FROM bookings
		WHERE cleaner_id = ANY($1)
	`

	// Apply filter if provided
	if filter != nil {
		switch *filter {
		case "UPCOMING":
			query += ` AND status IN ('PENDING', 'CONFIRMED', 'IN_PROGRESS') AND scheduled_date >= CURRENT_DATE`
		case "PAST":
			query += ` AND (status IN ('COMPLETED', 'CANCELLED', 'NO_SHOW_CLIENT', 'NO_SHOW_CLEANER') OR scheduled_date < CURRENT_DATE)`
		case "ACTIVE":
			query += ` AND status = 'IN_PROGRESS'`
		// "ALL" or default: no additional filter
		}
	}

	query += ` ORDER BY scheduled_date DESC, scheduled_time DESC`

	rows, err := r.db.Query(query, pq.Array(cleanerIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to query bookings by cleaner IDs: %w", err)
	}
	defer rows.Close()

	bookings := []*Booking{}
	for rows.Next() {
		booking := &Booking{}
		err := rows.Scan(
			&booking.ID, &booking.ClientID, &booking.CleanerID, &booking.AddressID,
			&booking.ServiceType, &booking.AreaSqm, &booking.EstimatedHours,
			&booking.ScheduledDate, &booking.ScheduledTime, &booking.EstimatedEndTime,
			&booking.TimePreferences,
			&booking.IncludesDeepCleaning, &booking.IncludesWindows, &booking.IncludesCarpetCleaning,
			&booking.NumberOfWindows, &booking.CarpetAreaSqm,
			&booking.IncludesFridgeCleaning, &booking.IncludesOvenCleaning, &booking.IncludesBalconyCleaning,
			&booking.SpecialInstructions, &booking.AccessInstructions,
			&booking.BasePrice, &booking.AddonsPrice, &booking.TotalPrice, &booking.PlatformFee, &booking.CleanerPayout, &booking.DiscountApplied,
			&booking.Status,
			&booking.ConfirmedAt, &booking.StartedAt, &booking.CompletedAt, &booking.CancelledAt,
			&booking.CancellationReason, &booking.CancelledBy,
			&booking.ClientRating, &booking.ClientReview, &booking.CleanerRating, &booking.CleanerReview,
			&booking.CreatedAt, &booking.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, booking)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bookings: %w", err)
	}

	return bookings, nil
}

// GetExpiredPendingBookings gets all PENDING bookings older than the specified duration
// Used for auto-expiring bookings that haven't been accepted within the timeout period
func (r *BookingRepository) GetExpiredPendingBookings(expirationHours int) ([]*Booking, error) {
	query := `
		SELECT id, client_id, cleaner_id, address_id,
		       service_type, area_sqm, estimated_hours, scheduled_date, scheduled_time, estimated_end_time,
		       includes_deep_cleaning, includes_windows, includes_carpet_cleaning,
		       number_of_windows, carpet_area_sqm, special_instructions, access_instructions,
		       base_price, addons_price, total_price, platform_fee, cleaner_payout, discount_applied,
		       status, confirmed_at, started_at, completed_at, cancelled_at,
		       cancellation_reason, cancelled_by,
		       client_rating, client_review, cleaner_rating, cleaner_review,
		       created_at, updated_at
		FROM bookings
		WHERE status = $1
		  AND created_at < NOW() - INTERVAL '1 hour' * $2
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query, BookingStatusPending, expirationHours)
	if err != nil {
		return nil, fmt.Errorf("failed to get expired bookings: %w", err)
	}
	defer rows.Close()

	var bookings []*Booking
	for rows.Next() {
		booking := &Booking{}
		err := rows.Scan(
			&booking.ID, &booking.ClientID, &booking.CleanerID, &booking.AddressID,
			&booking.ServiceType, &booking.AreaSqm, &booking.EstimatedHours,
			&booking.ScheduledDate, &booking.ScheduledTime, &booking.EstimatedEndTime,
			&booking.TimePreferences,
			&booking.IncludesDeepCleaning, &booking.IncludesWindows, &booking.IncludesCarpetCleaning,
			&booking.NumberOfWindows, &booking.CarpetAreaSqm,
			&booking.IncludesFridgeCleaning, &booking.IncludesOvenCleaning, &booking.IncludesBalconyCleaning,
			&booking.SpecialInstructions, &booking.AccessInstructions,
			&booking.BasePrice, &booking.AddonsPrice, &booking.TotalPrice, &booking.PlatformFee, &booking.CleanerPayout, &booking.DiscountApplied,
			&booking.Status,
			&booking.ConfirmedAt, &booking.StartedAt, &booking.CompletedAt, &booking.CancelledAt,
			&booking.CancellationReason, &booking.CancelledBy,
			&booking.ClientRating, &booking.ClientReview, &booking.CleanerRating, &booking.CleanerReview,
			&booking.CreatedAt, &booking.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan booking: %w", err)
		}
		bookings = append(bookings, booking)
	}

	return bookings, rows.Err()
}
// CountActiveBookingsByCleanerID counts active bookings (PENDING, CONFIRMED, IN_PROGRESS) for a cleaner
func (r *BookingRepository) CountActiveBookingsByCleanerID(cleanerID string) (int, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*)
		FROM bookings
		WHERE cleaner_id = $1
		  AND status IN ('PENDING', 'CONFIRMED', 'IN_PROGRESS')
	`, cleanerID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("failed to count active bookings: %w", err)
	}

	return count, nil
}
