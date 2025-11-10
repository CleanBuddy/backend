-- Add additional addon fields to bookings table
ALTER TABLE bookings
ADD COLUMN IF NOT EXISTS includes_fridge_cleaning BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN IF NOT EXISTS includes_oven_cleaning BOOLEAN NOT NULL DEFAULT false,
ADD COLUMN IF NOT EXISTS includes_balcony_cleaning BOOLEAN NOT NULL DEFAULT false;

COMMENT ON COLUMN bookings.includes_fridge_cleaning IS 'Whether booking includes fridge cleaning addon';
COMMENT ON COLUMN bookings.includes_oven_cleaning IS 'Whether booking includes oven cleaning addon';
COMMENT ON COLUMN bookings.includes_balcony_cleaning IS 'Whether booking includes balcony cleaning addon';
