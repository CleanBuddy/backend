-- Remove supplies column from bookings table
ALTER TABLE bookings
DROP COLUMN IF EXISTS supplies;
