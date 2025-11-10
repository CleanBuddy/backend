-- Migration: Add geolocation (latitude/longitude) to addresses table
-- This enables accurate distance calculations for cleaner matching

-- Add latitude and longitude columns
ALTER TABLE addresses
ADD COLUMN latitude DECIMAL(10, 8),
ADD COLUMN longitude DECIMAL(11, 8);

-- Add index for geospatial queries (PostGIS-like queries if needed in future)
CREATE INDEX idx_addresses_lat_lng ON addresses(latitude, longitude);

-- Add comment explaining the geolocation
COMMENT ON COLUMN addresses.latitude IS 'Latitude in decimal degrees (-90 to 90). Populated via geocoding API.';
COMMENT ON COLUMN addresses.longitude IS 'Longitude in decimal degrees (-180 to 180). Populated via geocoding API.';

-- Note: Existing addresses will have NULL lat/lng
-- They will be populated on next address update or via backfill script
