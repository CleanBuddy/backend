-- Remove additional addon fields from bookings table
ALTER TABLE bookings
DROP COLUMN IF EXISTS includes_fridge_cleaning,
DROP COLUMN IF EXISTS includes_oven_cleaning,
DROP COLUMN IF EXISTS includes_balcony_cleaning;
