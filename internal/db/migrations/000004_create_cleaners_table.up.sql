-- Cleaners table for cleaner profiles
CREATE TABLE IF NOT EXISTS cleaners (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::TEXT,
    user_id TEXT NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,

    -- Personal Information
    phone_number VARCHAR(20) NOT NULL,
    date_of_birth DATE,

    -- Address Information
    street_address VARCHAR(255),
    city VARCHAR(100),
    county VARCHAR(100),
    postal_code VARCHAR(20),

    -- Experience & Skills
    years_of_experience INTEGER DEFAULT 0,
    bio TEXT,
    specializations JSONB DEFAULT '[]'::jsonb,
    languages JSONB DEFAULT '["ro"]'::jsonb,

    -- KYC Documents (stored in GCP Cloud Storage)
    id_document_url TEXT,
    id_document_verified BOOLEAN DEFAULT false,
    background_check_url TEXT,
    background_check_verified BOOLEAN DEFAULT false,
    profile_photo_url TEXT,

    -- Ratings & Stats
    average_rating DECIMAL(3, 2),
    total_jobs INTEGER NOT NULL DEFAULT 0,
    total_earnings DECIMAL(10, 2) NOT NULL DEFAULT 0.00,

    -- Status
    approval_status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_available BOOLEAN NOT NULL DEFAULT false,

    -- Admin notes
    admin_notes TEXT,
    approved_by TEXT REFERENCES users(id),
    approved_at TIMESTAMP WITH TIME ZONE,

    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Add indexes
CREATE INDEX idx_cleaners_user_id ON cleaners(user_id);
CREATE INDEX idx_cleaners_approval_status ON cleaners(approval_status);
CREATE INDEX idx_cleaners_is_active ON cleaners(is_active);
CREATE INDEX idx_cleaners_is_available ON cleaners(is_available);
CREATE INDEX idx_cleaners_city ON cleaners(city);

-- Add trigger for updated_at
CREATE TRIGGER set_cleaners_updated_at
    BEFORE UPDATE ON cleaners
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Add constraint for approval_status enum
ALTER TABLE cleaners
    ADD CONSTRAINT cleaners_approval_status_check
    CHECK (approval_status IN ('PENDING', 'APPROVED', 'REJECTED'));

-- Comment on table
COMMENT ON TABLE cleaners IS 'Cleaner profiles with KYC verification and performance metrics';
