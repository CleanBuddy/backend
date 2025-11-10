-- Bookings table with state machine
CREATE TABLE IF NOT EXISTS bookings (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,

    -- Relationships
    client_id TEXT NOT NULL REFERENCES users(id),
    cleaner_id TEXT REFERENCES cleaners(id),
    address_id TEXT NOT NULL REFERENCES addresses(id),

    -- Service details
    service_type VARCHAR(50) NOT NULL,
    area_sqm INTEGER,
    estimated_hours INTEGER NOT NULL,

    -- Scheduling
    scheduled_date DATE NOT NULL,
    scheduled_time TIME NOT NULL,
    estimated_end_time TIME,

    -- Add-ons
    includes_deep_cleaning BOOLEAN NOT NULL DEFAULT false,
    includes_windows BOOLEAN NOT NULL DEFAULT false,
    includes_carpet_cleaning BOOLEAN NOT NULL DEFAULT false,
    number_of_windows INTEGER DEFAULT 0,
    carpet_area_sqm INTEGER DEFAULT 0,

    -- Special instructions
    special_instructions TEXT,
    access_instructions TEXT,

    -- Pricing
    base_price DECIMAL(10, 2) NOT NULL,
    addons_price DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    total_price DECIMAL(10, 2) NOT NULL,
    platform_fee DECIMAL(10, 2) NOT NULL,
    cleaner_payout DECIMAL(10, 2) NOT NULL,
    discount_applied DECIMAL(10, 2) DEFAULT 0.00,

    -- State machine
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING',

    -- Status timestamps
    confirmed_at TIMESTAMP WITH TIME ZONE,
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    cancelled_at TIMESTAMP WITH TIME ZONE,

    -- Cancellation
    cancellation_reason TEXT,
    cancelled_by TEXT REFERENCES users(id),

    -- Rating & Review
    client_rating INTEGER CHECK (client_rating >= 1 AND client_rating <= 5),
    client_review TEXT,
    cleaner_rating INTEGER CHECK (cleaner_rating >= 1 AND cleaner_rating <= 5),
    cleaner_review TEXT,

    -- Metadata
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Add indexes
CREATE INDEX idx_bookings_client_id ON bookings(client_id);
CREATE INDEX idx_bookings_cleaner_id ON bookings(cleaner_id);
CREATE INDEX idx_bookings_status ON bookings(status);
CREATE INDEX idx_bookings_scheduled_date ON bookings(scheduled_date);
CREATE INDEX idx_bookings_service_type ON bookings(service_type);
CREATE INDEX idx_bookings_created_at ON bookings(created_at DESC);

-- Add trigger for updated_at
CREATE TRIGGER set_bookings_updated_at
    BEFORE UPDATE ON bookings
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add constraint for status enum
ALTER TABLE bookings
    ADD CONSTRAINT bookings_status_check
    CHECK (status IN (
        'PENDING',           -- Initial state, waiting for cleaner assignment
        'CONFIRMED',         -- Cleaner assigned and confirmed
        'IN_PROGRESS',       -- Service is currently being performed
        'COMPLETED',         -- Service finished successfully
        'CANCELLED',         -- Booking was cancelled
        'NO_SHOW_CLIENT',    -- Client didn't show up
        'NO_SHOW_CLEANER'    -- Cleaner didn't show up
    ));

-- Add constraint for service_type enum (must match pricing_rules)
ALTER TABLE bookings
    ADD CONSTRAINT bookings_service_type_check
    CHECK (service_type IN ('STANDARD', 'DEEP_CLEANING', 'OFFICE', 'POST_RENOVATION', 'MOVE_IN_OUT'));

-- Comment on table
COMMENT ON TABLE bookings IS 'Bookings/jobs with state machine for tracking service lifecycle';
COMMENT ON COLUMN bookings.status IS 'State machine: PENDING → CONFIRMED → IN_PROGRESS → COMPLETED (or CANCELLED at any point)';
