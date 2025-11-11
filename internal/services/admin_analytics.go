package services

import (
	"database/sql"
	"fmt"
	"time"

	graphmodel "github.com/cleanbuddy/backend/internal/graph/model"
	"github.com/cleanbuddy/backend/internal/models"
)

type AdminAnalyticsService struct {
	db *sql.DB
}

func NewAdminAnalyticsService(db *sql.DB) *AdminAnalyticsService {
	return &AdminAnalyticsService{db: db}
}

// GetKPIs returns admin dashboard KPIs for a given period
func (s *AdminAnalyticsService) GetKPIs(period graphmodel.KPIPeriod) (*graphmodel.AdminKPIs, error) {
	startDate, endDate := s.getPeriodDates(period)

	// Get revenue metrics
	totalRevenue, platformFees, err := s.getRevenueMetrics(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get revenue metrics: %w", err)
	}

	// Get booking counts
	totalBookings, completedBookings, activeBookings, cancelledBookings, err := s.getBookingMetrics(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking metrics: %w", err)
	}

	// Get user counts
	activeCleaners, activeClients, err := s.getUserMetrics(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get user metrics: %w", err)
	}

	// Calculate average booking value
	averageBookingValue := 0.0
	if totalBookings > 0 {
		averageBookingValue = totalRevenue / float64(totalBookings)
	}

	// Get top cleaners
	topCleaners, err := s.getTopCleaners(startDate, endDate, 5)
	if err != nil {
		return nil, fmt.Errorf("failed to get top cleaners: %w", err)
	}

	return &graphmodel.AdminKPIs{
		Period:              period,
		TotalRevenue:        totalRevenue,
		PlatformFees:        platformFees,
		TotalBookings:       totalBookings,
		CompletedBookings:   completedBookings,
		ActiveBookings:      activeBookings,
		CancelledBookings:   cancelledBookings,
		ActiveCleaners:      activeCleaners,
		ActiveClients:       activeClients,
		AverageBookingValue: averageBookingValue,
		TopCleaners:         topCleaners,
	}, nil
}

func (s *AdminAnalyticsService) getPeriodDates(period graphmodel.KPIPeriod) (time.Time, time.Time) {
	now := time.Now().UTC()
	var startDate time.Time

	switch period {
	case graphmodel.KPIPeriodToday:
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	case graphmodel.KPIPeriodWeek:
		startDate = now.AddDate(0, 0, -7)
	case graphmodel.KPIPeriodMonth:
		startDate = now.AddDate(0, -1, 0)
	}

	return startDate, now
}

func (s *AdminAnalyticsService) getRevenueMetrics(startDate, endDate time.Time) (float64, float64, error) {
	query := `
		SELECT
			COALESCE(SUM(total_price), 0) as total_revenue,
			COALESCE(SUM(platform_fee), 0) as platform_fees
		FROM bookings
		WHERE created_at >= $1 AND created_at <= $2
			AND status IN ('CONFIRMED', 'IN_PROGRESS', 'COMPLETED')
	`

	var totalRevenue, platformFees float64
	err := s.db.QueryRow(query, startDate, endDate).Scan(&totalRevenue, &platformFees)
	if err != nil {
		return 0, 0, err
	}

	return totalRevenue, platformFees, nil
}

func (s *AdminAnalyticsService) getBookingMetrics(startDate, endDate time.Time) (int, int, int, int, error) {
	query := `
		SELECT
			COUNT(*) as total,
			COALESCE(SUM(CASE WHEN status = 'COMPLETED' THEN 1 ELSE 0 END), 0) as completed,
			COALESCE(SUM(CASE WHEN status IN ('CONFIRMED', 'IN_PROGRESS') THEN 1 ELSE 0 END), 0) as active,
			COALESCE(SUM(CASE WHEN status = 'CANCELLED' THEN 1 ELSE 0 END), 0) as cancelled
		FROM bookings
		WHERE created_at >= $1 AND created_at <= $2
	`

	var total, completed, active, cancelled int
	err := s.db.QueryRow(query, startDate, endDate).Scan(&total, &completed, &active, &cancelled)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	return total, completed, active, cancelled, nil
}

func (s *AdminAnalyticsService) getUserMetrics(startDate, endDate time.Time) (int, int, error) {
	// Active cleaners (with at least one booking in period)
	cleanersQuery := `
		SELECT COUNT(DISTINCT cleaner_id)
		FROM bookings
		WHERE created_at >= $1 AND created_at <= $2
			AND cleaner_id IS NOT NULL
	`

	var activeCleaners int
	err := s.db.QueryRow(cleanersQuery, startDate, endDate).Scan(&activeCleaners)
	if err != nil {
		return 0, 0, err
	}

	// Active clients (with at least one booking in period)
	clientsQuery := `
		SELECT COUNT(DISTINCT client_id)
		FROM bookings
		WHERE created_at >= $1 AND created_at <= $2
	`

	var activeClients int
	err = s.db.QueryRow(clientsQuery, startDate, endDate).Scan(&activeClients)
	if err != nil {
		return 0, 0, err
	}

	return activeCleaners, activeClients, nil
}

func (s *AdminAnalyticsService) getTopCleaners(startDate, endDate time.Time, limit int) ([]*graphmodel.TopCleanerStat, error) {
	query := `
		SELECT
			b.cleaner_id,
			u.first_name || ' ' || COALESCE(u.last_name, '') as cleaner_name,
			COUNT(b.id) as booking_count,
			COALESCE(SUM(b.cleaner_payout), 0) as total_earnings,
			c.average_rating
		FROM bookings b
		JOIN users u ON b.cleaner_id = u.id
		LEFT JOIN cleaners c ON c.user_id = u.id
		WHERE b.created_at >= $1 AND b.created_at <= $2
			AND b.status = 'COMPLETED'
			AND b.cleaner_id IS NOT NULL
		GROUP BY b.cleaner_id, u.first_name, u.last_name, c.average_rating
		ORDER BY total_earnings DESC
		LIMIT $3
	`

	rows, err := s.db.Query(query, startDate, endDate, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var topCleaners []*graphmodel.TopCleanerStat
	for rows.Next() {
		var stat graphmodel.TopCleanerStat
		var avgRating sql.NullFloat64

		err := rows.Scan(
			&stat.CleanerID,
			&stat.CleanerName,
			&stat.BookingCount,
			&stat.TotalEarnings,
			&avgRating,
		)
		if err != nil {
			return nil, err
		}

		if avgRating.Valid {
			stat.AverageRating = &avgRating.Float64
		}

		topCleaners = append(topCleaners, &stat)
	}

	return topCleaners, nil
}

// GetAllBookingsAdmin returns all bookings for admin with optional filters
func (s *AdminAnalyticsService) GetAllBookingsAdmin(limit, offset int, status *models.BookingStatus, search *string) ([]*models.Booking, error) {
	query := `
		SELECT id, reservation_code, client_id, address_id, cleaner_id,
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
		WHERE 1=1
	`

	args := []interface{}{}
	argCount := 1

	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, string(*status))
		argCount++
	}

	if search != nil && *search != "" {
		query += fmt.Sprintf(" AND (reservation_code ILIKE $%d OR special_instructions ILIKE $%d OR access_instructions ILIKE $%d)", argCount, argCount, argCount)
		args = append(args, "%"+*search+"%")
		argCount++
	}

	query += " ORDER BY created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
		argCount++
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
		argCount++
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []*models.Booking
	for rows.Next() {
		booking := &models.Booking{}
		err := rows.Scan(
			&booking.ID,
			&booking.ReservationCode,
			&booking.ClientID,
			&booking.AddressID,
			&booking.CleanerID,
			&booking.ServiceType,
			&booking.AreaSqm,
			&booking.EstimatedHours,
			&booking.ScheduledDate,
			&booking.ScheduledTime,
			&booking.EstimatedEndTime,
			&booking.IncludesDeepCleaning,
			&booking.IncludesWindows,
			&booking.IncludesCarpetCleaning,
			&booking.NumberOfWindows,
			&booking.CarpetAreaSqm,
			&booking.SpecialInstructions,
			&booking.AccessInstructions,
			&booking.BasePrice,
			&booking.AddonsPrice,
			&booking.TotalPrice,
			&booking.PlatformFee,
			&booking.CleanerPayout,
			&booking.DiscountApplied,
			&booking.Status,
			&booking.ConfirmedAt,
			&booking.StartedAt,
			&booking.CompletedAt,
			&booking.CancelledAt,
			&booking.CancellationReason,
			&booking.CancelledBy,
			&booking.ClientRating,
			&booking.ClientReview,
			&booking.CleanerRating,
			&booking.CleanerReview,
			&booking.CreatedAt,
			&booking.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		bookings = append(bookings, booking)
	}

	return bookings, nil
}

// GetPlatformStats returns platform-wide statistics for the landing page
// This is a public endpoint (no auth required)
func (s *AdminAnalyticsService) GetPlatformStats() (*graphmodel.PlatformStats, error) {
	stats := &graphmodel.PlatformStats{}

	// Get total approved cleaners
	err := s.db.QueryRow(`
		SELECT COUNT(*)
		FROM cleaners
		WHERE approval_status = 'APPROVED'
	`).Scan(&stats.TotalCleaners)
	if err != nil {
		return nil, fmt.Errorf("failed to count cleaners: %w", err)
	}

	// Get total completed bookings
	err = s.db.QueryRow(`
		SELECT COUNT(*)
		FROM bookings
		WHERE status = 'COMPLETED'
	`).Scan(&stats.TotalBookings)
	if err != nil {
		return nil, fmt.Errorf("failed to count bookings: %w", err)
	}

	// Get average rating from completed bookings with cleaner ratings
	var avgRating sql.NullFloat64
	err = s.db.QueryRow(`
		SELECT AVG(cleaner_rating)
		FROM bookings
		WHERE cleaner_rating IS NOT NULL
		AND cleaner_rating > 0
	`).Scan(&avgRating)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate average rating: %w", err)
	}

	if avgRating.Valid {
		stats.AverageRating = avgRating.Float64
	} else {
		stats.AverageRating = 0.0
	}

	// Get count of distinct cities served (from bookings)
	err = s.db.QueryRow(`
		SELECT COUNT(DISTINCT city)
		FROM addresses
		WHERE id IN (
			SELECT DISTINCT address_id
			FROM bookings
			WHERE status IN ('COMPLETED', 'IN_PROGRESS', 'CONFIRMED')
		)
	`).Scan(&stats.CitiesServed)
	if err != nil {
		return nil, fmt.Errorf("failed to count cities: %w", err)
	}

	return stats, nil
}
