-- Add rejected_reason column to cleaners table
ALTER TABLE cleaners ADD COLUMN rejected_reason TEXT;

-- Add comment
COMMENT ON COLUMN cleaners.rejected_reason IS 'Reason for rejection if approval_status is REJECTED';
