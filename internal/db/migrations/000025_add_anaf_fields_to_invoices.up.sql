-- Add ANAF e-Factura integration fields to invoices table
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS anaf_upload_index VARCHAR(255);
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS anaf_status VARCHAR(50) DEFAULT 'pending';
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS anaf_submitted_at TIMESTAMP;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS anaf_processed_at TIMESTAMP;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS anaf_download_id VARCHAR(255);
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS anaf_confirmation_url TEXT;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS anaf_errors JSONB;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS anaf_retry_count INTEGER DEFAULT 0;
ALTER TABLE invoices ADD COLUMN IF NOT EXISTS anaf_last_retry_at TIMESTAMP;

-- Create index for ANAF status queries
CREATE INDEX IF NOT EXISTS idx_invoices_anaf_status ON invoices(anaf_status);
CREATE INDEX IF NOT EXISTS idx_invoices_anaf_submitted_at ON invoices(anaf_submitted_at);

-- Add comments for documentation
COMMENT ON COLUMN invoices.anaf_upload_index IS 'Unique identifier from ANAF after upload';
COMMENT ON COLUMN invoices.anaf_status IS 'ANAF submission status: pending, processing, accepted, rejected, failed';
COMMENT ON COLUMN invoices.anaf_submitted_at IS 'Timestamp when invoice was submitted to ANAF';
COMMENT ON COLUMN invoices.anaf_processed_at IS 'Timestamp when ANAF finished processing';
COMMENT ON COLUMN invoices.anaf_download_id IS 'ANAF download ID for confirmation PDF';
COMMENT ON COLUMN invoices.anaf_confirmation_url IS 'URL to ANAF confirmation document';
COMMENT ON COLUMN invoices.anaf_errors IS 'JSON array of errors from ANAF if rejected';
COMMENT ON COLUMN invoices.anaf_retry_count IS 'Number of submission retry attempts';
COMMENT ON COLUMN invoices.anaf_last_retry_at IS 'Timestamp of last retry attempt';
