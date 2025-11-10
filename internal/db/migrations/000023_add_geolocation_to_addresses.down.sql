-- Rollback: Remove geolocation columns from addresses table

DROP INDEX IF EXISTS idx_addresses_lat_lng;

ALTER TABLE addresses
DROP COLUMN IF EXISTS latitude,
DROP COLUMN IF EXISTS longitude;
