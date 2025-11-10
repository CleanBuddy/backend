-- Add supplies column to bookings table
ALTER TABLE bookings
ADD COLUMN IF NOT EXISTS supplies VARCHAR(50);

COMMENT ON COLUMN bookings.supplies IS 'Who provides cleaning supplies: client_provides, cleaner_provides, or platform_provides';
