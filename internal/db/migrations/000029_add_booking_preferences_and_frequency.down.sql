-- Rollback: Remove booking preferences and frequency fields

ALTER TABLE bookings DROP COLUMN IF EXISTS frequency;
ALTER TABLE bookings DROP COLUMN IF EXISTS time_preferences;
ALTER TABLE bookings DROP COLUMN IF EXISTS extras;
ALTER TABLE bookings DROP COLUMN IF EXISTS supplies;
