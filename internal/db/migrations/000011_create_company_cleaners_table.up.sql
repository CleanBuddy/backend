-- Create company_cleaners junction table for company employees
CREATE TABLE IF NOT EXISTS company_cleaners (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,

    -- Relations
    company_id TEXT NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    cleaner_id TEXT NOT NULL REFERENCES cleaners(id) ON DELETE CASCADE,

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE', -- ACTIVE, INACTIVE

    -- Audit
    joined_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP WITH TIME ZONE,

    -- Constraints
    UNIQUE(company_id, cleaner_id)
);

-- Indexes
CREATE INDEX idx_company_cleaners_company_id ON company_cleaners(company_id);
CREATE INDEX idx_company_cleaners_cleaner_id ON company_cleaners(cleaner_id);

-- Comments
COMMENT ON TABLE company_cleaners IS 'Junction table linking cleaners to companies they work for';
COMMENT ON COLUMN company_cleaners.status IS 'Employment status: ACTIVE, INACTIVE';
