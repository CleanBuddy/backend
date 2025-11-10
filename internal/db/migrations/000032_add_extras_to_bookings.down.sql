-- Remove extras column from bookings table
ALTER TABLE bookings
DROP COLUMN IF EXISTS extras;
