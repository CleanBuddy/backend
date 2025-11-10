-- Rollback: Remove geolocation columns from cleaners table

DROP INDEX IF EXISTS idx_cleaners_lat_lng;

ALTER TABLE cleaners
DROP COLUMN IF EXISTS latitude,
DROP COLUMN IF EXISTS longitude;
