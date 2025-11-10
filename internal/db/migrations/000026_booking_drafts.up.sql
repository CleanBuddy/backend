-- Create booking_drafts table for anonymous booking flow
-- Stores draft bookings before user authentication

CREATE TABLE IF NOT EXISTS booking_drafts (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,

    -- Anonymous identifier (before user creation)
    session_id VARCHAR(255) NOT NULL,

    -- User ID (NULL until authentication in step 5)
    user_id TEXT REFERENCES users(id) ON DELETE SET NULL,

    -- Booking data stored as JSONB for flexibility
    -- Structure:
    -- {
    --   "cleaning_type": "regular|deep_clean|move_in_out|post_renovation|office",
    --   "duration_hours": 3.5,
    --   "frequency": "one_time|weekly|biweekly|monthly",
    --   "address": {
    --     "street": "Str. Victoriei 45",
    --     "city": "București",
    --     "county": "București",
    --     "postal_code": "010063",
    --     "locality_id": 179,
    --     "apartment_details": "Apt 5, Floor 2",
    --     "latitude": 44.4392,
    --     "longitude": 26.0961
    --   },
    --   "scheduled_date": "2025-10-11",
    --   "time_slot": "morning|afternoon|evening",
    --   "scheduled_start_time": "2025-10-11T10:00:00Z",
    --   "extras": ["ironing", "windows", "fridge_oven", "balcony"],
    --   "supplies": "client_provides|cleaner_provides",
    --   "special_instructions": "Please use eco-friendly products",
    --   "includes_windows": true
    -- }
    draft_data JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Current wizard step (1-6)
    current_step INT NOT NULL DEFAULT 1 CHECK (current_step >= 1 AND current_step <= 6),

    -- Pricing snapshot (calculated by backend)
    estimated_price_ron DECIMAL(10,2),

    -- Price breakdown stored as JSONB
    -- {
    --   "base_price": 150.00,
    --   "extras_total": 60.00,
    --   "supplies_fee": 15.00,
    --   "frequency_discount": -22.50,
    --   "subtotal": 202.50,
    --   "vat": 38.48,
    --   "total": 240.98,
    --   "items": [
    --     {"description": "4 hours regular cleaning", "amount": 150.00},
    --     {"description": "Ironing service", "amount": 30.00},
    --     {"description": "Window cleaning", "amount": 40.00},
    --     {"description": "Cleaner provides supplies", "amount": 15.00},
    --     {"description": "Weekly frequency discount (10%)", "amount": -22.50}
    --   ]
    -- }
    price_breakdown JSONB,

    -- Conversion tracking
    converted_to_booking_id TEXT REFERENCES bookings(id) ON DELETE SET NULL,
    conversion_completed_at TIMESTAMPTZ,

    -- Expiration (24 hours from creation)
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '24 hours'),

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for fast lookups
CREATE INDEX idx_booking_drafts_session_id ON booking_drafts(session_id);
CREATE INDEX idx_booking_drafts_user_id ON booking_drafts(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_booking_drafts_expires_at ON booking_drafts(expires_at);
CREATE INDEX idx_booking_drafts_converted ON booking_drafts(converted_to_booking_id) WHERE converted_to_booking_id IS NOT NULL;

-- Trigger to update updated_at timestamp
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

-- Comments for documentation
COMMENT ON TABLE booking_drafts IS 'Stores anonymous booking drafts before user authentication (Step 1-4), converted to real bookings after auth (Step 5-6)';
COMMENT ON COLUMN booking_drafts.session_id IS 'Anonymous UUID stored in browser localStorage to track draft across steps';
COMMENT ON COLUMN booking_drafts.draft_data IS 'JSONB containing all booking details collected during wizard steps 1-4';
COMMENT ON COLUMN booking_drafts.current_step IS 'Current wizard step (1=service, 2=location, 3=extras, 4=review, 5=auth, 6=payment)';
COMMENT ON COLUMN booking_drafts.estimated_price_ron IS 'Calculated price snapshot for quick display';
COMMENT ON COLUMN booking_drafts.price_breakdown IS 'Detailed price calculation with line items';
COMMENT ON COLUMN booking_drafts.expires_at IS 'Drafts auto-expire after 24 hours to prevent database bloat';
