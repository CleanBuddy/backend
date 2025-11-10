-- Drop addresses table
DROP TRIGGER IF EXISTS ensure_single_default_address_trigger ON addresses;
DROP FUNCTION IF EXISTS ensure_single_default_address;
DROP TRIGGER IF EXISTS update_addresses_updated_at ON addresses;
DROP INDEX IF EXISTS idx_addresses_default;
DROP INDEX IF EXISTS idx_addresses_user_id;
DROP TABLE IF EXISTS addresses;
