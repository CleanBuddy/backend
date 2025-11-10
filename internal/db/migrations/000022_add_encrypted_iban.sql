-- Migration: Add encrypted_iban column to cleaners table
-- This migration adds encrypted IBAN storage while maintaining backward compatibility

-- Add encrypted_iban column (will store base64-encoded AES-256-GCM ciphertext)
ALTER TABLE cleaners ADD COLUMN encrypted_iban TEXT;

-- Create index for faster lookups (even though encrypted data can't be searched efficiently)
CREATE INDEX idx_cleaners_encrypted_iban ON cleaners(encrypted_iban) WHERE encrypted_iban IS NOT NULL;

-- Add comment explaining the encryption
COMMENT ON COLUMN cleaners.encrypted_iban IS 'AES-256-GCM encrypted IBAN (base64-encoded). Use encryption service to decrypt.';
COMMENT ON COLUMN cleaners.iban IS 'DEPRECATED: Plaintext IBAN. Will be migrated to encrypted_iban and removed in future release.';

-- Migration strategy:
-- 1. Add encrypted_iban column (this migration)
-- 2. Update application code to write to encrypted_iban and read from both
-- 3. Backfill encrypted_iban from iban (separate migration script)
-- 4. Update code to only use encrypted_iban
-- 5. Drop iban column (future migration)
