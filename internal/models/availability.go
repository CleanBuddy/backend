package models

import (
	"database/sql"
	"fmt"
	"time"
)

// AvailabilityType constants
const (
	AvailabilityTypeRecurring = "RECURRING"
	AvailabilityTypeOneTime   = "ONE_TIME"
	AvailabilityTypeBlocked   = "BLOCKED"
)

// Availability represents a cleaner's availability schedule
type Availability struct {
	ID           string
	CleanerID    string
	Type         string
	DayOfWeek    sql.NullInt32
	SpecificDate sql.NullTime
	StartTime    string
	EndTime      string
	IsActive     bool
	Notes        sql.NullString
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// AvailabilityRepository handles database operations for availability
type AvailabilityRepository struct {
	db *sql.DB
}

// NewAvailabilityRepository creates a new availability repository
func NewAvailabilityRepository(db *sql.DB) *AvailabilityRepository {
	return &AvailabilityRepository{db: db}
}

// Create creates a new availability record
func (r *AvailabilityRepository) Create(availability *Availability) error {
	query := `
		INSERT INTO availability (
			cleaner_id, type, day_of_week, specific_date,
			start_time, end_time, is_active, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	err := r.db.QueryRow(
		query,
		availability.CleanerID,
		availability.Type,
		availability.DayOfWeek,
		availability.SpecificDate,
		availability.StartTime,
		availability.EndTime,
		availability.IsActive,
		availability.Notes,
	).Scan(&availability.ID, &availability.CreatedAt, &availability.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create availability: %w", err)
	}

	return nil
}

// GetByID retrieves an availability record by ID
func (r *AvailabilityRepository) GetByID(id string) (*Availability, error) {
	query := `
		SELECT id, cleaner_id, type, day_of_week, specific_date,
			   start_time, end_time, is_active, notes, created_at, updated_at
		FROM availability
		WHERE id = $1
	`

	availability := &Availability{}
	err := r.db.QueryRow(query, id).Scan(
		&availability.ID,
		&availability.CleanerID,
		&availability.Type,
		&availability.DayOfWeek,
		&availability.SpecificDate,
		&availability.StartTime,
		&availability.EndTime,
		&availability.IsActive,
		&availability.Notes,
		&availability.CreatedAt,
		&availability.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get availability: %w", err)
	}

	return availability, nil
}

// GetByCleanerID retrieves all availability records for a cleaner
func (r *AvailabilityRepository) GetByCleanerID(cleanerID string) ([]*Availability, error) {
	query := `
		SELECT id, cleaner_id, type, day_of_week, specific_date,
			   start_time, end_time, is_active, notes, created_at, updated_at
		FROM availability
		WHERE cleaner_id = $1
		ORDER BY type, day_of_week, specific_date, start_time
	`

	rows, err := r.db.Query(query, cleanerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get availability by cleaner: %w", err)
	}
	defer rows.Close()

	var availabilities []*Availability
	for rows.Next() {
		availability := &Availability{}
		err := rows.Scan(
			&availability.ID,
			&availability.CleanerID,
			&availability.Type,
			&availability.DayOfWeek,
			&availability.SpecificDate,
			&availability.StartTime,
			&availability.EndTime,
			&availability.IsActive,
			&availability.Notes,
			&availability.CreatedAt,
			&availability.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan availability: %w", err)
		}
		availabilities = append(availabilities, availability)
	}

	return availabilities, nil
}

// Update updates an availability record
func (r *AvailabilityRepository) Update(availability *Availability) error {
	query := `
		UPDATE availability
		SET start_time = $1, end_time = $2, is_active = $3, notes = $4, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		RETURNING updated_at
	`

	err := r.db.QueryRow(
		query,
		availability.StartTime,
		availability.EndTime,
		availability.IsActive,
		availability.Notes,
		availability.ID,
	).Scan(&availability.UpdatedAt)

	if err == sql.ErrNoRows {
		return fmt.Errorf("availability not found")
	}
	if err != nil {
		return fmt.Errorf("failed to update availability: %w", err)
	}

	return nil
}

// Delete deletes an availability record
func (r *AvailabilityRepository) Delete(id string) error {
	query := `DELETE FROM availability WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete availability: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("availability not found")
	}

	return nil
}

// GetRecurringByDayOfWeek retrieves recurring availability for a specific day of week
func (r *AvailabilityRepository) GetRecurringByDayOfWeek(cleanerID string, dayOfWeek int) ([]*Availability, error) {
	query := `
		SELECT id, cleaner_id, type, day_of_week, specific_date,
			   start_time, end_time, is_active, notes, created_at, updated_at
		FROM availability
		WHERE cleaner_id = $1 AND type = $2 AND day_of_week = $3 AND is_active = true
		ORDER BY start_time
	`

	rows, err := r.db.Query(query, cleanerID, AvailabilityTypeRecurring, dayOfWeek)
	if err != nil {
		return nil, fmt.Errorf("failed to get recurring availability: %w", err)
	}
	defer rows.Close()

	var availabilities []*Availability
	for rows.Next() {
		availability := &Availability{}
		err := rows.Scan(
			&availability.ID,
			&availability.CleanerID,
			&availability.Type,
			&availability.DayOfWeek,
			&availability.SpecificDate,
			&availability.StartTime,
			&availability.EndTime,
			&availability.IsActive,
			&availability.Notes,
			&availability.CreatedAt,
			&availability.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan availability: %w", err)
		}
		availabilities = append(availabilities, availability)
	}

	return availabilities, nil
}

// GetByDateRange retrieves availability within a date range
func (r *AvailabilityRepository) GetByDateRange(cleanerID string, startDate, endDate time.Time) ([]*Availability, error) {
	query := `
		SELECT id, cleaner_id, type, day_of_week, specific_date,
			   start_time, end_time, is_active, notes, created_at, updated_at
		FROM availability
		WHERE cleaner_id = $1
		  AND is_active = true
		  AND (
		      (type = $2 AND specific_date >= $3 AND specific_date <= $4)
		      OR type = $5
		  )
		ORDER BY specific_date, start_time
	`

	rows, err := r.db.Query(
		query,
		cleanerID,
		AvailabilityTypeOneTime,
		startDate,
		endDate,
		AvailabilityTypeRecurring,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get availability by date range: %w", err)
	}
	defer rows.Close()

	var availabilities []*Availability
	for rows.Next() {
		availability := &Availability{}
		err := rows.Scan(
			&availability.ID,
			&availability.CleanerID,
			&availability.Type,
			&availability.DayOfWeek,
			&availability.SpecificDate,
			&availability.StartTime,
			&availability.EndTime,
			&availability.IsActive,
			&availability.Notes,
			&availability.CreatedAt,
			&availability.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan availability: %w", err)
		}
		availabilities = append(availabilities, availability)
	}

	return availabilities, nil
}

// CheckConflict checks if there's a conflicting availability slot
func (r *AvailabilityRepository) CheckConflict(cleanerID, availabilityType string, dayOfWeek sql.NullInt32, specificDate sql.NullTime, startTime, endTime string) (bool, error) {
	var query string
	var args []interface{}

	if availabilityType == AvailabilityTypeRecurring {
		query = `
			SELECT COUNT(*) > 0
			FROM availability
			WHERE cleaner_id = $1
			  AND type = $2
			  AND day_of_week = $3
			  AND is_active = true
			  AND (
			      (start_time <= $4 AND end_time > $4)
			      OR (start_time < $5 AND end_time >= $5)
			      OR (start_time >= $4 AND end_time <= $5)
			  )
		`
		args = []interface{}{cleanerID, availabilityType, dayOfWeek, startTime, endTime}
	} else {
		query = `
			SELECT COUNT(*) > 0
			FROM availability
			WHERE cleaner_id = $1
			  AND type = $2
			  AND specific_date = $3
			  AND is_active = true
			  AND (
			      (start_time <= $4 AND end_time > $4)
			      OR (start_time < $5 AND end_time >= $5)
			      OR (start_time >= $4 AND end_time <= $5)
			  )
		`
		args = []interface{}{cleanerID, availabilityType, specificDate, startTime, endTime}
	}

	var hasConflict bool
	err := r.db.QueryRow(query, args...).Scan(&hasConflict)
	if err != nil {
		return false, fmt.Errorf("failed to check conflict: %w", err)
	}

	return hasConflict, nil
}
