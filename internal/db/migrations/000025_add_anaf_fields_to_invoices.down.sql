-- Remove ANAF integration fields from invoices table
DROP INDEX IF EXISTS idx_invoices_anaf_submitted_at;
DROP INDEX IF EXISTS idx_invoices_anaf_status;

ALTER TABLE invoices DROP COLUMN IF EXISTS anaf_last_retry_at;
ALTER TABLE invoices DROP COLUMN IF EXISTS anaf_retry_count;
ALTER TABLE invoices DROP COLUMN IF EXISTS anaf_errors;
ALTER TABLE invoices DROP COLUMN IF EXISTS anaf_confirmation_url;
ALTER TABLE invoices DROP COLUMN IF EXISTS anaf_download_id;
ALTER TABLE invoices DROP COLUMN IF EXISTS anaf_processed_at;
ALTER TABLE invoices DROP COLUMN IF EXISTS anaf_submitted_at;
ALTER TABLE invoices DROP COLUMN IF EXISTS anaf_status;
ALTER TABLE invoices DROP COLUMN IF EXISTS anaf_upload_index;
