-- Drop disputes table and related objects
DROP TRIGGER IF EXISTS set_disputes_updated_at ON disputes;
DROP INDEX IF EXISTS idx_disputes_created_at;
DROP INDEX IF EXISTS idx_disputes_assigned_to;
DROP INDEX IF EXISTS idx_disputes_status;
DROP INDEX IF EXISTS idx_disputes_created_by;
DROP INDEX IF EXISTS idx_disputes_booking_id;
DROP TABLE IF EXISTS disputes;
