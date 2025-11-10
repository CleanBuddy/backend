-- Add time_preferences column to bookings table
-- This allows flexible scheduling where client provides multiple date/time options
-- and cleaner chooses the most convenient one

ALTER TABLE bookings
ADD COLUMN IF NOT EXISTS time_preferences JSONB;

-- Comment for documentation
COMMENT ON COLUMN bookings.time_preferences IS 'JSONB array of preferred date/time slots: [{"date": "2025-10-15", "timeSlots": ["morning", "afternoon"]}, ...]';

-- If time_preferences is provided, scheduled_date and scheduled_time may be NULL initially
-- Cleaner will select from the preferences and update scheduled_date/scheduled_time
