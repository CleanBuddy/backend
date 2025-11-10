-- Drop booking_drafts table and related objects
-- Draft bookings are now managed in frontend localStorage only

-- Drop indexes
DROP INDEX IF EXISTS idx_booking_drafts_session_id;
DROP INDEX IF EXISTS idx_booking_drafts_user_id;
DROP INDEX IF EXISTS idx_booking_drafts_expires_at;
DROP INDEX IF EXISTS idx_booking_drafts_converted;

-- Drop trigger
DROP TRIGGER IF EXISTS booking_drafts_updated_at ON booking_drafts;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_booking_drafts_updated_at();

-- Drop table
DROP TABLE IF EXISTS booking_drafts;
