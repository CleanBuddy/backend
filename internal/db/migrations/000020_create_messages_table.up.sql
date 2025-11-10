-- Messages table for client-cleaner communication
CREATE TABLE messages (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    booking_id TEXT NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
    sender_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    receiver_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for fast message retrieval
CREATE INDEX idx_messages_booking_id ON messages(booking_id);
CREATE INDEX idx_messages_sender_id ON messages(sender_id);
CREATE INDEX idx_messages_receiver_id ON messages(receiver_id);
CREATE INDEX idx_messages_created_at ON messages(created_at DESC);
CREATE INDEX idx_messages_unread ON messages(receiver_id, is_read) WHERE is_read = FALSE;

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_messages_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER messages_updated_at
BEFORE UPDATE ON messages
FOR EACH ROW
EXECUTE FUNCTION update_messages_updated_at();

-- Add check constraint to ensure sender and receiver are different
ALTER TABLE messages ADD CONSTRAINT messages_different_users
CHECK (sender_id != receiver_id);

COMMENT ON TABLE messages IS 'Messages exchanged between clients and cleaners for specific bookings';
COMMENT ON COLUMN messages.booking_id IS 'The booking this conversation is about';
COMMENT ON COLUMN messages.sender_id IS 'User who sent the message';
COMMENT ON COLUMN messages.receiver_id IS 'User who receives the message';
COMMENT ON COLUMN messages.content IS 'Message text content';
COMMENT ON COLUMN messages.is_read IS 'Whether the receiver has read this message';
