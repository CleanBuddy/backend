package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

const (
	PhotoTypeBefore          = "BEFORE"
	PhotoTypeAfter           = "AFTER"
	PhotoTypeDisputeEvidence = "DISPUTE_EVIDENCE"
)

type Photo struct {
	ID         string
	BookingID  string
	DisputeID  sql.NullString
	UploadedBy string
	PhotoType  string
	FilePath   string
	FileName   string
	FileSize   int
	MimeType   string
	URL        string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type PhotoRepository struct {
	db *sql.DB
}

func NewPhotoRepository(db *sql.DB) *PhotoRepository {
	return &PhotoRepository{db: db}
}

func (r *PhotoRepository) Create(photo *Photo) error {
	if photo.ID == "" {
		photo.ID = uuid.New().String()
	}

	query := `
		INSERT INTO photos (id, booking_id, dispute_id, uploaded_by, photo_type, file_path, file_name, file_size, mime_type, url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING created_at, updated_at
	`
	return r.db.QueryRow(
		query,
		photo.ID,
		photo.BookingID,
		photo.DisputeID,
		photo.UploadedBy,
		photo.PhotoType,
		photo.FilePath,
		photo.FileName,
		photo.FileSize,
		photo.MimeType,
		photo.URL,
	).Scan(&photo.CreatedAt, &photo.UpdatedAt)
}

func (r *PhotoRepository) GetByID(id string) (*Photo, error) {
	photo := &Photo{}
	query := `
		SELECT id, booking_id, dispute_id, uploaded_by, photo_type, file_path, file_name, file_size, mime_type, url, created_at, updated_at
		FROM photos
		WHERE id = $1
	`
	err := r.db.QueryRow(query, id).Scan(
		&photo.ID,
		&photo.BookingID,
		&photo.DisputeID,
		&photo.UploadedBy,
		&photo.PhotoType,
		&photo.FilePath,
		&photo.FileName,
		&photo.FileSize,
		&photo.MimeType,
		&photo.URL,
		&photo.CreatedAt,
		&photo.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return photo, nil
}

func (r *PhotoRepository) GetByBookingID(bookingID string) ([]*Photo, error) {
	query := `
		SELECT id, booking_id, dispute_id, uploaded_by, photo_type, file_path, file_name, file_size, mime_type, url, created_at, updated_at
		FROM photos
		WHERE booking_id = $1
		ORDER BY photo_type, created_at
	`
	rows, err := r.db.Query(query, bookingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []*Photo
	for rows.Next() {
		photo := &Photo{}
		if err := rows.Scan(
			&photo.ID,
			&photo.BookingID,
			&photo.DisputeID,
			&photo.UploadedBy,
			&photo.PhotoType,
			&photo.FilePath,
			&photo.FileName,
			&photo.FileSize,
			&photo.MimeType,
			&photo.URL,
			&photo.CreatedAt,
			&photo.UpdatedAt,
		); err != nil {
			return nil, err
		}
		photos = append(photos, photo)
	}
	return photos, nil
}

func (r *PhotoRepository) GetByDisputeID(disputeID string) ([]*Photo, error) {
	query := `
		SELECT id, booking_id, dispute_id, uploaded_by, photo_type, file_path, file_name, file_size, mime_type, url, created_at, updated_at
		FROM photos
		WHERE dispute_id = $1
		ORDER BY created_at
	`
	rows, err := r.db.Query(query, disputeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var photos []*Photo
	for rows.Next() {
		photo := &Photo{}
		if err := rows.Scan(
			&photo.ID,
			&photo.BookingID,
			&photo.DisputeID,
			&photo.UploadedBy,
			&photo.PhotoType,
			&photo.FilePath,
			&photo.FileName,
			&photo.FileSize,
			&photo.MimeType,
			&photo.URL,
			&photo.CreatedAt,
			&photo.UpdatedAt,
		); err != nil {
			return nil, err
		}
		photos = append(photos, photo)
	}
	return photos, nil
}

func (r *PhotoRepository) Delete(id string) error {
	query := `DELETE FROM photos WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
