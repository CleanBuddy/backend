-- Rollback: Remove time_preferences column from bookings table

ALTER TABLE bookings
DROP COLUMN IF EXISTS time_preferences;
