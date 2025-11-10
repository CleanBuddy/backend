package models

import (
	"database/sql"
	"time"
)

// Message represents a message between client and cleaner
type Message struct {
	ID         string    `json:"id"`
	BookingID  string    `json:"booking_id"`
	SenderID   string    `json:"sender_id"`
	ReceiverID string    `json:"receiver_id"`
	Content    string    `json:"content"`
	IsRead     bool      `json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// MessageRepository handles database operations for messages
type MessageRepository struct {
	db *sql.DB
}

// NewMessageRepository creates a new message repository
func NewMessageRepository(db *sql.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// Create creates a new message
func (r *MessageRepository) Create(message *Message) error {
	query := `
		INSERT INTO messages (booking_id, sender_id, receiver_id, content, is_read)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(
		query,
		message.BookingID,
		message.SenderID,
		message.ReceiverID,
		message.Content,
		message.IsRead,
	).Scan(&message.ID, &message.CreatedAt, &message.UpdatedAt)
}

// GetByID retrieves a message by ID
func (r *MessageRepository) GetByID(id string) (*Message, error) {
	message := &Message{}
	query := `
		SELECT id, booking_id, sender_id, receiver_id, content, is_read, created_at, updated_at
		FROM messages
		WHERE id = $1
	`
	err := r.db.QueryRow(query, id).Scan(
		&message.ID,
		&message.BookingID,
		&message.SenderID,
		&message.ReceiverID,
		&message.Content,
		&message.IsRead,
		&message.CreatedAt,
		&message.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return message, err
}

// GetByBookingID retrieves all messages for a booking (ordered chronologically)
func (r *MessageRepository) GetByBookingID(bookingID string) ([]*Message, error) {
	query := `
		SELECT id, booking_id, sender_id, receiver_id, content, is_read, created_at, updated_at
		FROM messages
		WHERE booking_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.Query(query, bookingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		message := &Message{}
		if err := rows.Scan(
			&message.ID,
			&message.BookingID,
			&message.SenderID,
			&message.ReceiverID,
			&message.Content,
			&message.IsRead,
			&message.CreatedAt,
			&message.UpdatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, rows.Err()
}

// GetConversationsByUserID retrieves all unique booking conversations for a user
func (r *MessageRepository) GetConversationsByUserID(userID string) ([]*Message, error) {
	// Get the most recent message for each booking where user is involved
	query := `
		SELECT DISTINCT ON (booking_id)
			id, booking_id, sender_id, receiver_id, content, is_read, created_at, updated_at
		FROM messages
		WHERE sender_id = $1 OR receiver_id = $1
		ORDER BY booking_id, created_at DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		message := &Message{}
		if err := rows.Scan(
			&message.ID,
			&message.BookingID,
			&message.SenderID,
			&message.ReceiverID,
			&message.Content,
			&message.IsRead,
			&message.CreatedAt,
			&message.UpdatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, rows.Err()
}

// MarkAsRead marks a message as read
func (r *MessageRepository) MarkAsRead(id string) error {
	query := `UPDATE messages SET is_read = TRUE WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// MarkAllAsReadForBooking marks all messages as read for a booking+receiver combination
func (r *MessageRepository) MarkAllAsReadForBooking(bookingID string, receiverID string) error {
	query := `
		UPDATE messages
		SET is_read = TRUE
		WHERE booking_id = $1 AND receiver_id = $2 AND is_read = FALSE
	`
	_, err := r.db.Exec(query, bookingID, receiverID)
	return err
}

// GetUnreadCountForUser counts unread messages for a user
func (r *MessageRepository) GetUnreadCountForUser(userID string) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM messages
		WHERE receiver_id = $1 AND is_read = FALSE
	`
	err := r.db.QueryRow(query, userID).Scan(&count)
	return count, err
}

// GetUnreadCountForBooking counts unread messages for a booking+receiver
func (r *MessageRepository) GetUnreadCountForBooking(bookingID string, receiverID string) (int, error) {
	var count int
	query := `
		SELECT COUNT(*)
		FROM messages
		WHERE booking_id = $1 AND receiver_id = $2 AND is_read = FALSE
	`
	err := r.db.QueryRow(query, bookingID, receiverID).Scan(&count)
	return count, err
}

// Delete deletes a message
func (r *MessageRepository) Delete(id string) error {
	query := `DELETE FROM messages WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
