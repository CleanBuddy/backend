-- Create cleaner_applications table for progressive disclosure application flow
-- Stores cleaner applications from anonymous visitors through approval

CREATE TABLE IF NOT EXISTS cleaner_applications (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,

    -- Anonymous identifier (before user creation in steps 1-2)
    session_id VARCHAR(255),

    -- User ID (NULL until authentication in step 3)
    user_id TEXT REFERENCES users(id) ON DELETE SET NULL,

    -- Application data stored as JSONB for flexibility
    -- Structure:
    -- {
    --   "eligibility": {
    --     "age_18_plus": true,
    --     "work_right": "romanian_citizen|eu_citizen|work_permit",
    --     "experience": "professional|some|willing_to_learn"
    --   },
    --   "availability": {
    --     "hours_per_week": "5-10|10-20|20-30|30+",
    --     "areas": ["sector_1", "sector_2", "sector_3"],
    --     "days": ["monday", "tuesday", "wednesday", "thursday", "friday"],
    --     "time_slots": ["morning", "afternoon", "evening"],
    --     "estimated_monthly_earnings": 5500.00
    --   },
    --   "profile": {
    --     "photo_url": "https://storage.googleapis.com/...",
    --     "bio": "Professional cleaner with 5 years experience...",
    --     "languages": ["romanian", "english", "italian"],
    --     "equipment": ["vacuum", "mop", "products", "steam_cleaner"]
    --   },
    --   "legal": {
    --     "status": "pfa|srl|individual",
    --     "cif": "RO12345678",
    --     "cnp_encrypted": "encrypted_base64_string",
    --     "iban": "RO49AAAA1B31007593840000",
    --     "bank_name": "BCR"
    --   },
    --   "documents": {
    --     "cazier_url": "https://storage.googleapis.com/...",
    --     "id_front_url": "https://storage.googleapis.com/...",
    --     "id_back_url": "https://storage.googleapis.com/...",
    --     "insurance_url": "https://storage.googleapis.com/..."
    --   }
    -- }
    application_data JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- Current wizard step (1-6)
    current_step INT NOT NULL DEFAULT 1 CHECK (current_step >= 1 AND current_step <= 6),

    -- Application status
    status VARCHAR(50) NOT NULL DEFAULT 'draft',
    -- Possible values:
    --   'draft' - Still being filled out
    --   'submitted' - Completed and submitted for review
    --   'under_review' - Admin is reviewing
    --   'approved' - Application approved, cleaner profile created
    --   'rejected' - Application rejected
    --   'incomplete' - Missing required information

    -- Review metadata
    reviewed_by TEXT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at TIMESTAMPTZ,
    rejection_reason TEXT,
    admin_notes TEXT,

    -- Conversion tracking
    converted_to_cleaner_id TEXT REFERENCES cleaners(id) ON DELETE SET NULL,

    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    submitted_at TIMESTAMPTZ
);

-- Indexes for fast lookups
CREATE INDEX idx_cleaner_applications_session_id ON cleaner_applications(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX idx_cleaner_applications_user_id ON cleaner_applications(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_cleaner_applications_status ON cleaner_applications(status);
CREATE INDEX idx_cleaner_applications_submitted ON cleaner_applications(submitted_at) WHERE submitted_at IS NOT NULL;
CREATE INDEX idx_cleaner_applications_pending_review ON cleaner_applications(status) WHERE status = 'submitted';

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_cleaner_applications_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cleaner_applications_updated_at
    BEFORE UPDATE ON cleaner_applications
    FOR EACH ROW
    EXECUTE FUNCTION update_cleaner_applications_updated_at();

-- Constraint: Either session_id or user_id must be present
ALTER TABLE cleaner_applications
ADD CONSTRAINT cleaner_applications_identifier_check
CHECK (session_id IS NOT NULL OR user_id IS NOT NULL);

-- Constraint: Status must be valid
ALTER TABLE cleaner_applications
ADD CONSTRAINT cleaner_applications_status_check
CHECK (status IN ('draft', 'submitted', 'under_review', 'approved', 'rejected', 'incomplete'));

-- Constraint: If status is rejected, rejection_reason must be present
CREATE OR REPLACE FUNCTION check_rejection_reason()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.status = 'rejected' AND (NEW.rejection_reason IS NULL OR NEW.rejection_reason = '') THEN
        RAISE EXCEPTION 'rejection_reason is required when status is rejected';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER cleaner_applications_rejection_check
    BEFORE INSERT OR UPDATE ON cleaner_applications
    FOR EACH ROW
    WHEN (NEW.status = 'rejected')
    EXECUTE FUNCTION check_rejection_reason();

-- Comments for documentation
COMMENT ON TABLE cleaner_applications IS 'Stores cleaner applications from anonymous visitors through approval process';
COMMENT ON COLUMN cleaner_applications.session_id IS 'Anonymous UUID for steps 1-2 before authentication';
COMMENT ON COLUMN cleaner_applications.user_id IS 'Set after authentication in step 3';
COMMENT ON COLUMN cleaner_applications.application_data IS 'JSONB containing all application details collected during wizard';
COMMENT ON COLUMN cleaner_applications.current_step IS 'Current wizard step (1=eligibility, 2=availability, 3=auth, 4=profile, 5=legal, 6=documents)';
COMMENT ON COLUMN cleaner_applications.status IS 'Application lifecycle: draft → submitted → under_review → approved/rejected';
COMMENT ON COLUMN cleaner_applications.rejection_reason IS 'Admin explanation if application is rejected (shown to applicant)';
COMMENT ON COLUMN cleaner_applications.admin_notes IS 'Internal admin notes (not shown to applicant)';
COMMENT ON COLUMN cleaner_applications.converted_to_cleaner_id IS 'References cleaners.id after approval and profile creation';
