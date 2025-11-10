-- Remove reservation code from bookings table
DROP INDEX IF EXISTS idx_bookings_reservation_code;
ALTER TABLE bookings DROP COLUMN IF EXISTS reservation_code;
