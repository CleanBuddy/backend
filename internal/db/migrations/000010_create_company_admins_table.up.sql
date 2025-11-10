-- Create company_admins junction table for company administrators
CREATE TABLE IF NOT EXISTS company_admins (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,

    -- Relations
    company_id TEXT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Role
    role VARCHAR(20) NOT NULL DEFAULT 'ADMIN', -- OWNER, ADMIN

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Constraints
    UNIQUE(company_id, user_id)
);

-- Indexes
CREATE INDEX idx_company_admins_company_id ON company_admins(company_id);
CREATE INDEX idx_company_admins_user_id ON company_admins(user_id);

-- Comments
COMMENT ON TABLE company_admins IS 'Junction table linking users to companies they administer';
COMMENT ON COLUMN company_admins.role IS 'Admin role: OWNER (creator), ADMIN (can manage cleaners)';
