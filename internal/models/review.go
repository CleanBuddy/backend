package models

import (
	"database/sql"
	"time"
)

// ReviewerRole represents who is reviewing
type ReviewerRole string

const (
	ReviewerRoleClient  ReviewerRole = "CLIENT"
	ReviewerRoleCleaner ReviewerRole = "CLEANER"
)

// Review represents a review for a booking
type Review struct {
	ID           string
	BookingID    string
	ReviewerID   string
	RevieweeID   string
	ReviewerRole ReviewerRole
	Rating       int
	Comment      sql.NullString
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ReviewRepository handles review database operations
type ReviewRepository struct {
	db *sql.DB
}

// NewReviewRepository creates a new review repository
func NewReviewRepository(db *sql.DB) *ReviewRepository {
	return &ReviewRepository{db: db}
}

// Create creates a new review
func (r *ReviewRepository) Create(review *Review) error {
	return r.db.QueryRow(`
		INSERT INTO reviews (
			booking_id, reviewer_id, reviewee_id, reviewer_role,
			rating, comment
		)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`, review.BookingID, review.ReviewerID, review.RevieweeID, review.ReviewerRole,
		review.Rating, review.Comment).
		Scan(&review.ID, &review.CreatedAt, &review.UpdatedAt)
}

// GetByID finds a review by ID
func (r *ReviewRepository) GetByID(id string) (*Review, error) {
	review := &Review{}
	err := r.db.QueryRow(`
		SELECT id, booking_id, reviewer_id, reviewee_id, reviewer_role,
		       rating, comment, created_at, updated_at
		FROM reviews
		WHERE id = $1
	`, id).Scan(
		&review.ID, &review.BookingID, &review.ReviewerID, &review.RevieweeID, &review.ReviewerRole,
		&review.Rating, &review.Comment, &review.CreatedAt, &review.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return review, nil
}

// GetByBookingID finds a review by booking ID
func (r *ReviewRepository) GetByBookingID(bookingID string) (*Review, error) {
	review := &Review{}
	err := r.db.QueryRow(`
		SELECT id, booking_id, reviewer_id, reviewee_id, reviewer_role,
		       rating, comment, created_at, updated_at
		FROM reviews
		WHERE booking_id = $1
	`, bookingID).Scan(
		&review.ID, &review.BookingID, &review.ReviewerID, &review.RevieweeID, &review.ReviewerRole,
		&review.Rating, &review.Comment, &review.CreatedAt, &review.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return review, nil
}

// GetByRevieweeID returns all reviews for a reviewee (cleaner)
func (r *ReviewRepository) GetByRevieweeID(revieweeID string, limit, offset int) ([]*Review, error) {
	query := `
		SELECT id, booking_id, reviewer_id, reviewee_id, reviewer_role,
		       rating, comment, created_at, updated_at
		FROM reviews
		WHERE reviewee_id = $1
		ORDER BY created_at DESC
	`

	args := []interface{}{revieweeID}

	if limit > 0 {
		query += ` LIMIT $2`
		args = append(args, limit)
		if offset > 0 {
			query += ` OFFSET $3`
			args = append(args, offset)
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reviews := []*Review{}
	for rows.Next() {
		review := &Review{}
		err := rows.Scan(
			&review.ID, &review.BookingID, &review.ReviewerID, &review.RevieweeID, &review.ReviewerRole,
			&review.Rating, &review.Comment, &review.CreatedAt, &review.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}

	return reviews, rows.Err()
}

// CalculateAverageRating calculates the average rating for a reviewee
func (r *ReviewRepository) CalculateAverageRating(revieweeID string) (float64, int, error) {
	var avgRating sql.NullFloat64
	var totalReviews int

	err := r.db.QueryRow(`
		SELECT AVG(rating), COUNT(*)
		FROM reviews
		WHERE reviewee_id = $1
	`, revieweeID).Scan(&avgRating, &totalReviews)

	if err != nil {
		return 0, 0, err
	}

	if !avgRating.Valid {
		return 0, 0, nil
	}

	return avgRating.Float64, totalReviews, nil
}

// Update updates a review
func (r *ReviewRepository) Update(review *Review) error {
	_, err := r.db.Exec(`
		UPDATE reviews
		SET rating = $2, comment = $3, updated_at = NOW()
		WHERE id = $1
	`, review.ID, review.Rating, review.Comment)
	return err
}
