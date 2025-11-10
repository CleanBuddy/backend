-- Create companies table for cleaning company management
CREATE TABLE IF NOT EXISTS companies (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,

    -- Company details
    name VARCHAR(255) NOT NULL,
    cui VARCHAR(20) NOT NULL UNIQUE, -- Romanian tax ID
    registration_number VARCHAR(50),

    -- Banking
    iban VARCHAR(34),
    bank_name VARCHAR(255),

    -- Legal
    legal_address TEXT,
    contact_email VARCHAR(255),
    contact_phone VARCHAR(20),

    -- Documents
    id_document_url TEXT,
    registration_document_url TEXT,
    id_document_verified BOOLEAN NOT NULL DEFAULT false,
    registration_document_verified BOOLEAN NOT NULL DEFAULT false,

    -- Approval
    approval_status VARCHAR(20) NOT NULL DEFAULT 'PENDING', -- PENDING, APPROVED, REJECTED
    rejected_reason TEXT,
    approved_by TEXT REFERENCES users(id),
    approved_at TIMESTAMP WITH TIME ZONE,

    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_companies_approval_status ON companies(approval_status);
CREATE INDEX idx_companies_cui ON companies(cui);

-- Trigger for updated_at
CREATE TRIGGER update_companies_updated_at
    BEFORE UPDATE ON companies
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON TABLE companies IS 'Cleaning companies with their legal and banking details';
COMMENT ON COLUMN companies.cui IS 'Romanian unique tax identification number (CUI)';
COMMENT ON COLUMN companies.approval_status IS 'Admin approval status: PENDING, APPROVED, REJECTED';
