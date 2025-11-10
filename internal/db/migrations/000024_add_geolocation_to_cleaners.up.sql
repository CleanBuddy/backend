-- Migration: Add geolocation (latitude/longitude) to cleaners table
-- This enables accurate distance calculations for cleaner matching

-- Add latitude and longitude columns
ALTER TABLE cleaners
ADD COLUMN latitude DECIMAL(10, 8),
ADD COLUMN longitude DECIMAL(11, 8);

-- Add index for geospatial queries
CREATE INDEX idx_cleaners_lat_lng ON cleaners(latitude, longitude);

-- Add comment explaining the geolocation
COMMENT ON COLUMN cleaners.latitude IS 'Latitude in decimal degrees (-90 to 90). Populated from cleaner address via geocoding API.';
COMMENT ON COLUMN cleaners.longitude IS 'Longitude in decimal degrees (-180 to 180). Populated from cleaner address via geocoding API.';

-- Note: Existing cleaners will have NULL lat/lng
-- They will be populated when cleaner updates their address
