-- Drop availability table
DROP TRIGGER IF EXISTS update_availability_updated_at ON availability;
DROP INDEX IF EXISTS idx_availability_is_active;
DROP INDEX IF EXISTS idx_availability_specific_date;
DROP INDEX IF EXISTS idx_availability_day_of_week;
DROP INDEX IF EXISTS idx_availability_cleaner_id;
DROP TABLE IF EXISTS availability;
