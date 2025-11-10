-- Revert IBAN column size back to original
-- Note: This will fail if there are encrypted IBANs longer than 34 characters
ALTER TABLE cleaners
    ALTER COLUMN iban TYPE VARCHAR(34);
