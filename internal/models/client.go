package models

import (
	"database/sql"
	"time"
)

// Client represents a client profile in the system
type Client struct {
	ID                       string
	UserID                   string
	PhoneNumber              sql.NullString
	PreferredLanguage        string
	NotificationPreferences  []byte // JSONB stored as bytes
	TotalBookings            int
	TotalSpent               float64
	AverageRating            sql.NullFloat64
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

// ClientRepository handles client database operations
type ClientRepository struct {
	db *sql.DB
}

// NewClientRepository creates a new client repository
func NewClientRepository(db *sql.DB) *ClientRepository {
	return &ClientRepository{db: db}
}

// GetByUserID finds a client by user ID
func (r *ClientRepository) GetByUserID(userID string) (*Client, error) {
	client := &Client{}
	err := r.db.QueryRow(`
		SELECT id, user_id, phone_number, preferred_language, notification_preferences,
		       total_bookings, total_spent, average_rating, created_at, updated_at
		FROM clients
		WHERE user_id = $1
	`, userID).Scan(
		&client.ID, &client.UserID, &client.PhoneNumber, &client.PreferredLanguage,
		&client.NotificationPreferences, &client.TotalBookings, &client.TotalSpent,
		&client.AverageRating, &client.CreatedAt, &client.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return client, nil
}

// GetByID finds a client by ID
func (r *ClientRepository) GetByID(id string) (*Client, error) {
	client := &Client{}
	err := r.db.QueryRow(`
		SELECT id, user_id, phone_number, preferred_language, notification_preferences,
		       total_bookings, total_spent, average_rating, created_at, updated_at
		FROM clients
		WHERE id = $1
	`, id).Scan(
		&client.ID, &client.UserID, &client.PhoneNumber, &client.PreferredLanguage,
		&client.NotificationPreferences, &client.TotalBookings, &client.TotalSpent,
		&client.AverageRating, &client.CreatedAt, &client.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return client, nil
}

// Create creates a new client profile
func (r *ClientRepository) Create(client *Client) error {
	return r.db.QueryRow(`
		INSERT INTO clients (user_id, phone_number, preferred_language)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`, client.UserID, client.PhoneNumber, client.PreferredLanguage).
		Scan(&client.ID, &client.CreatedAt, &client.UpdatedAt)
}

// Update updates a client profile
func (r *ClientRepository) Update(client *Client) error {
	_, err := r.db.Exec(`
		UPDATE clients
		SET phone_number = $2, preferred_language = $3, notification_preferences = $4
		WHERE id = $1
	`, client.ID, client.PhoneNumber, client.PreferredLanguage, client.NotificationPreferences)
	return err
}
