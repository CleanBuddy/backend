-- Create payments table
CREATE TABLE IF NOT EXISTS payments (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,

    -- Relations
    booking_id TEXT NOT NULL REFERENCES bookings(id) ON DELETE RESTRICT,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,

    -- Payment provider details
    provider VARCHAR(50) NOT NULL, -- 'NETOPIA', 'MANUAL'
    provider_transaction_id VARCHAR(255), -- Transaction ID from payment provider
    provider_order_id VARCHAR(255), -- Order ID from payment provider

    -- Payment type and status
    payment_type VARCHAR(50) NOT NULL, -- 'PREAUTHORIZATION', 'CAPTURE', 'REFUND', 'CANCELLATION'
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING', -- 'PENDING', 'AUTHORIZED', 'CAPTURED', 'FAILED', 'REFUNDED', 'CANCELLED'

    -- Amounts (in RON)
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'RON',

    -- Card details (last 4 digits only, no sensitive data)
    card_last_four VARCHAR(4),
    card_brand VARCHAR(50), -- 'VISA', 'MASTERCARD', etc.

    -- Payment metadata
    error_code VARCHAR(100),
    error_message TEXT,
    provider_response JSONB, -- Full response from payment provider

    -- Timeline
    authorized_at TIMESTAMP WITH TIME ZONE,
    captured_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    refunded_at TIMESTAMP WITH TIME ZONE,

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_payments_booking_id ON payments(booking_id);
CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_provider_transaction_id ON payments(provider_transaction_id);
CREATE INDEX idx_payments_created_at ON payments(created_at DESC);

-- Trigger for updated_at
CREATE TRIGGER update_payments_updated_at
    BEFORE UPDATE ON payments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE payments IS 'Payment transactions for bookings with provider integration';
COMMENT ON COLUMN payments.provider IS 'Payment gateway: NETOPIA, MANUAL';
COMMENT ON COLUMN payments.payment_type IS 'Type: PREAUTHORIZATION (hold funds), CAPTURE (charge), REFUND, CANCELLATION';
COMMENT ON COLUMN payments.status IS 'Status: PENDING, AUTHORIZED, CAPTURED, FAILED, REFUNDED, CANCELLED';
COMMENT ON COLUMN payments.provider_response IS 'Full JSON response from payment provider for debugging';
