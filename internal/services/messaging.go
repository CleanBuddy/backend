package services

import (
	"database/sql"
	"fmt"

	"github.com/cleanbuddy/backend/internal/models"
)

// MessagingService handles messaging business logic
type MessagingService struct {
	messageRepo *models.MessageRepository
	bookingRepo *models.BookingRepository
	userRepo    *models.UserRepository
}

// NewMessagingService creates a new messaging service
func NewMessagingService(db *sql.DB) *MessagingService {
	return &MessagingService{
		messageRepo: models.NewMessageRepository(db),
		bookingRepo: models.NewBookingRepository(db),
		userRepo:    models.NewUserRepository(db),
	}
}

// SendMessage sends a message from sender to receiver for a booking
func (s *MessagingService) SendMessage(bookingID, senderID, receiverID, content string) (*models.Message, error) {
	// Validate booking exists and sender/receiver are involved
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	// Verify sender and receiver are part of the booking
	if !s.isUserInBooking(senderID, booking) {
		return nil, fmt.Errorf("sender is not part of this booking")
	}
	if !s.isUserInBooking(receiverID, booking) {
		return nil, fmt.Errorf("receiver is not part of this booking")
	}

	// Prevent sending messages to self
	if senderID == receiverID {
		return nil, fmt.Errorf("cannot send message to yourself")
	}

	// Create message
	message := &models.Message{
		BookingID:  bookingID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		IsRead:     false,
	}

	if err := s.messageRepo.Create(message); err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}

	return message, nil
}

// GetBookingMessages retrieves all messages for a booking (user must be involved)
func (s *MessagingService) GetBookingMessages(bookingID, userID string) ([]*models.Message, error) {
	// Verify user is part of the booking
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return nil, fmt.Errorf("booking not found")
	}

	if !s.isUserInBooking(userID, booking) {
		return nil, fmt.Errorf("unauthorized: not part of this booking")
	}

	messages, err := s.messageRepo.GetByBookingID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	return messages, nil
}

// GetMyConversations retrieves all conversations for a user
func (s *MessagingService) GetMyConversations(userID string) ([]*models.Message, error) {
	messages, err := s.messageRepo.GetConversationsByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversations: %w", err)
	}

	return messages, nil
}

// MarkMessagesAsRead marks all messages as read for a booking
func (s *MessagingService) MarkMessagesAsRead(bookingID, userID string) error {
	// Verify user is part of the booking
	booking, err := s.bookingRepo.GetByID(bookingID)
	if err != nil {
		return fmt.Errorf("failed to get booking: %w", err)
	}
	if booking == nil {
		return fmt.Errorf("booking not found")
	}

	if !s.isUserInBooking(userID, booking) {
		return fmt.Errorf("unauthorized: not part of this booking")
	}

	if err := s.messageRepo.MarkAllAsReadForBooking(bookingID, userID); err != nil {
		return fmt.Errorf("failed to mark messages as read: %w", err)
	}

	return nil
}

// GetUnreadCount gets unread message count for a user
func (s *MessagingService) GetUnreadCount(userID string) (int, error) {
	count, err := s.messageRepo.GetUnreadCountForUser(userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// GetUnreadCountForBooking gets unread message count for a booking
func (s *MessagingService) GetUnreadCountForBooking(bookingID, userID string) (int, error) {
	count, err := s.messageRepo.GetUnreadCountForBooking(bookingID, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count for booking: %w", err)
	}

	return count, nil
}

// isUserInBooking checks if a user is part of a booking (client or cleaner)
func (s *MessagingService) isUserInBooking(userID string, booking *models.Booking) bool {
	if booking.ClientID == userID {
		return true
	}
	if booking.CleanerID.Valid && booking.CleanerID.String == userID {
		return true
	}
	return false
}
