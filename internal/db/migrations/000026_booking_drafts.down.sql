-- Drop booking_drafts table and related objects

DROP TRIGGER IF EXISTS booking_drafts_updated_at ON booking_drafts;
DROP FUNCTION IF EXISTS update_booking_drafts_updated_at();

DROP INDEX IF EXISTS idx_booking_drafts_converted;
DROP INDEX IF EXISTS idx_booking_drafts_expires_at;
DROP INDEX IF EXISTS idx_booking_drafts_user_id;
DROP INDEX IF EXISTS idx_booking_drafts_session_id;

DROP TABLE IF EXISTS booking_drafts;
