-- Add extras column to bookings table
ALTER TABLE bookings
ADD COLUMN IF NOT EXISTS extras TEXT[];

COMMENT ON COLUMN bookings.extras IS 'Array of extra services: windows, fridge_oven, balcony, ironing, etc.';
