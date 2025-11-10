-- Add reservation code to bookings table
-- Reservation code is a unique, human-friendly identifier (e.g., CB-2024-A1B2C3)
ALTER TABLE bookings
    ADD COLUMN reservation_code VARCHAR(20) UNIQUE;

-- Create index for fast lookups
CREATE INDEX idx_bookings_reservation_code ON bookings(reservation_code);

-- Comment
COMMENT ON COLUMN bookings.reservation_code IS 'Human-friendly unique identifier for customer reference (e.g., CB-2024-A1B2C3)';
