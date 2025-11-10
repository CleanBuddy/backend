package services

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/cleanbuddy/backend/internal/models"
	"github.com/google/uuid"
)

type PhotoService struct {
	photoRepo  *models.PhotoRepository
	uploadPath string // Local path for development, will be GCP in production
}

func NewPhotoService(db *sql.DB, uploadPath string) *PhotoService {
	return &PhotoService{
		photoRepo:  models.NewPhotoRepository(db),
		uploadPath: uploadPath,
	}
}

// UploadPhoto handles file upload and database record creation
func (s *PhotoService) UploadPhoto(upload graphql.Upload, bookingID, userID, photoType string) (*models.Photo, error) {
	// Validate photo type
	if photoType != models.PhotoTypeBefore && photoType != models.PhotoTypeAfter && photoType != models.PhotoTypeDisputeEvidence {
		return nil, fmt.Errorf("invalid photo type: %s", photoType)
	}

	// Validate file size (max 10MB)
	if upload.Size > 10*1024*1024 {
		return nil, fmt.Errorf("file too large: maximum size is 10MB")
	}

	// Validate mime type
	if !isValidImageType(upload.ContentType) {
		return nil, fmt.Errorf("invalid file type: only JPG, PNG, and WebP images are allowed")
	}

	// Generate unique filename
	ext := filepath.Ext(upload.Filename)
	filename := fmt.Sprintf("%s_%s%s", uuid.New().String(), time.Now().Format("20060102150405"), ext)

	// Create directory structure: uploads/bookings/{bookingID}/
	dirPath := filepath.Join(s.uploadPath, "bookings", bookingID)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Full file path
	filePath := filepath.Join(dirPath, filename)

	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy uploaded file to destination
	if _, err := io.Copy(dst, upload.File); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Generate URL (for dev, this will be local; for prod, will be GCP URL)
	url := fmt.Sprintf("/uploads/bookings/%s/%s", bookingID, filename)

	// Create database record
	photo := &models.Photo{
		BookingID:  bookingID,
		UploadedBy: userID,
		PhotoType:  photoType,
		FilePath:   filePath,
		FileName:   upload.Filename,
		FileSize:   int(upload.Size),
		MimeType:   upload.ContentType,
		URL:        url,
	}

	if err := s.photoRepo.Create(photo); err != nil {
		// Clean up file if database insert fails
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save photo record: %w", err)
	}

	return photo, nil
}

// UploadDisputePhoto handles dispute evidence photo upload
func (s *PhotoService) UploadDisputePhoto(upload graphql.Upload, disputeID, userID string) (*models.Photo, error) {
	// Validate file size (max 10MB)
	if upload.Size > 10*1024*1024 {
		return nil, fmt.Errorf("file too large: maximum size is 10MB")
	}

	// Validate mime type
	if !isValidImageType(upload.ContentType) {
		return nil, fmt.Errorf("invalid file type: only JPG, PNG, and WebP images are allowed")
	}

	// Generate unique filename
	ext := filepath.Ext(upload.Filename)
	filename := fmt.Sprintf("%s_%s%s", uuid.New().String(), time.Now().Format("20060102150405"), ext)

	// Create directory structure: uploads/disputes/{disputeID}/
	dirPath := filepath.Join(s.uploadPath, "disputes", disputeID)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Full file path
	filePath := filepath.Join(dirPath, filename)

	// Create the file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy uploaded file to destination
	if _, err := io.Copy(dst, upload.File); err != nil {
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Generate URL
	url := fmt.Sprintf("/uploads/disputes/%s/%s", disputeID, filename)

	// Create database record
	photo := &models.Photo{
		DisputeID:  sql.NullString{String: disputeID, Valid: true},
		UploadedBy: userID,
		PhotoType:  models.PhotoTypeDisputeEvidence,
		FilePath:   filePath,
		FileName:   upload.Filename,
		FileSize:   int(upload.Size),
		MimeType:   upload.ContentType,
		URL:        url,
	}

	if err := s.photoRepo.Create(photo); err != nil {
		// Clean up file if database insert fails
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to save photo record: %w", err)
	}

	return photo, nil
}

func (s *PhotoService) GetPhotosByBookingID(bookingID string) ([]*models.Photo, error) {
	return s.photoRepo.GetByBookingID(bookingID)
}

func (s *PhotoService) GetPhotosByDisputeID(disputeID string) ([]*models.Photo, error) {
	return s.photoRepo.GetByDisputeID(disputeID)
}

func (s *PhotoService) DeletePhoto(photoID string) error {
	// Get photo to retrieve file path
	photo, err := s.photoRepo.GetByID(photoID)
	if err != nil {
		return err
	}
	if photo == nil {
		return fmt.Errorf("photo not found")
	}

	// Delete file from filesystem
	if err := os.Remove(photo.FilePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Delete database record
	return s.photoRepo.Delete(photoID)
}

func isValidImageType(mimeType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/webp",
	}
	mimeType = strings.ToLower(mimeType)
	for _, valid := range validTypes {
		if mimeType == valid {
			return true
		}
	}
	return false
}
