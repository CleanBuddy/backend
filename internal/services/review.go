package services

import (
	"database/sql"
	"fmt"

	"github.com/cleanbuddy/backend/internal/models"
)

// ReviewService handles review business logic
type ReviewService struct {
	reviewRepo  *models.ReviewRepository
	bookingRepo *models.BookingRepository
	cleanerRepo *models.CleanerRepository
}

// NewReviewService creates a new review service
func NewReviewService(db *sql.DB) *ReviewService {
	return &ReviewService{
		reviewRepo:  models.NewReviewRepository(db),
		bookingRepo: models.NewBookingRepository(db),
		cleanerRepo: models.NewCleanerRepository(db),
	}
}

// CreateReview creates a new review for a booking
func (s *ReviewService) CreateReview(bookingID, reviewerID string, rating int, comment string) (*models.Review, error) {
	// Validate rating
	if rating < 1 || rating > 5 {
		return nil, fmt.Errorf("rating must be between 1 and 5")
	}

	// Get booking details
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Validate booking is completed
	if booking.Status != models.BookingStatusCompleted {
		return nil, fmt.Errorf("can only review completed bookings")
	}

	// Check if review already exists
	existing, err := s.reviewRepo.GetByBookingID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing review: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("review already exists for this booking")
	}

	// Determine reviewer role and reviewee
	var reviewerRole models.ReviewerRole
	var revieweeID string

	if booking.ClientID == reviewerID {
		reviewerRole = models.ReviewerRoleClient
		if booking.CleanerID.Valid {
			revieweeID = booking.CleanerID.String
		} else {
			return nil, fmt.Errorf("no cleaner assigned to booking")
		}
	} else if booking.CleanerID.Valid && booking.CleanerID.String == reviewerID {
		reviewerRole = models.ReviewerRoleCleaner
		revieweeID = booking.ClientID
	} else {
		return nil, fmt.Errorf("unauthorized to review this booking")
	}

	// Create review
	review := &models.Review{
		BookingID:    bookingID,
		ReviewerID:   reviewerID,
		RevieweeID:   revieweeID,
		ReviewerRole: reviewerRole,
		Rating:       rating,
	}

	if comment != "" {
		review.Comment = sql.NullString{String: comment, Valid: true}
	}

	if err := s.reviewRepo.Create(review); err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	// Update reviewee's average rating (only for cleaners)
	if reviewerRole == models.ReviewerRoleClient {
		if err := s.updateCleanerRating(revieweeID); err != nil {
			// Log error but don't fail the review creation
			fmt.Printf("Warning: failed to update cleaner rating: %v\n", err)
		}
	}

	return review, nil
}

// GetReviewByBookingID gets a review by booking ID
func (s *ReviewService) GetReviewByBookingID(bookingID string) (*models.Review, error) {
	return s.reviewRepo.GetByBookingID(bookingID)
}

// GetCleanerReviews gets all reviews for a cleaner
func (s *ReviewService) GetCleanerReviews(cleanerID string, limit, offset int) ([]*models.Review, error) {
	return s.reviewRepo.GetByRevieweeID(cleanerID, limit, offset)
}

// updateCleanerRating updates a cleaner's average rating and total reviews
func (s *ReviewService) updateCleanerRating(cleanerID string) error {
	// Calculate new average rating
	avgRating, totalReviews, err := s.reviewRepo.CalculateAverageRating(cleanerID)
	if err != nil {
		return fmt.Errorf("failed to calculate average rating: %w", err)
	}

	// Get cleaner
	cleaner, err := s.cleanerRepo.GetByUserID(cleanerID)
	if err != nil {
		return fmt.Errorf("failed to get cleaner: %w", err)
	}
	if cleaner == nil {
		return fmt.Errorf("cleaner not found")
	}

	// Update cleaner stats
	if avgRating > 0 {
		cleaner.AverageRating = sql.NullFloat64{Float64: avgRating, Valid: true}
	}
	cleaner.TotalJobs = totalReviews // Using TotalJobs as a proxy for total reviews

	if err := s.cleanerRepo.Update(cleaner); err != nil {
		return fmt.Errorf("failed to update cleaner: %w", err)
	}

	return nil
}
