-- Drop clients table
DROP TRIGGER IF EXISTS update_clients_updated_at ON clients;
DROP INDEX IF EXISTS idx_clients_user_id;
DROP TABLE IF EXISTS clients;
