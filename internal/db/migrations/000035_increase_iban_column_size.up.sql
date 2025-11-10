-- Increase IBAN column size to accommodate encrypted data
-- Encrypted IBANs (AES-256-GCM + base64) require ~100-150 characters
ALTER TABLE cleaners
    ALTER COLUMN iban TYPE VARCHAR(255);

COMMENT ON COLUMN cleaners.iban IS 'Encrypted IBAN for payouts (AES-256-GCM encrypted, base64 encoded)';
