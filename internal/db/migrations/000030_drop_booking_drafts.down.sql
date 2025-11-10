-- Recreate booking_drafts table (rollback)
-- This is the reverse of the drop migration

CREATE TABLE IF NOT EXISTS booking_drafts (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    session_id VARCHAR(255) NOT NULL,
    user_id TEXT REFERENCES users(id) ON DELETE SET NULL,
    draft_data JSONB NOT NULL DEFAULT '{}'::jsonb,
    current_step INT NOT NULL DEFAULT 1 CHECK (current_step >= 1 AND current_step <= 6),
    estimated_price_ron DECIMAL(10,2),
    price_breakdown JSONB,
    converted_to_booking_id TEXT REFERENCES bookings(id) ON DELETE SET NULL,
    conversion_completed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '24 hours'),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_booking_drafts_session_id ON booking_drafts(session_id);
CREATE INDEX idx_booking_drafts_user_id ON booking_drafts(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_booking_drafts_expires_at ON booking_drafts(expires_at);
CREATE INDEX idx_booking_drafts_converted ON booking_drafts(converted_to_booking_id) WHERE converted_to_booking_id IS NOT NULL;

CREATE OR REPLACE FUNCTION update_booking_drafts_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER booking_drafts_updated_at
    BEFORE UPDATE ON booking_drafts
    FOR EACH ROW
    EXECUTE FUNCTION update_booking_drafts_updated_at();
