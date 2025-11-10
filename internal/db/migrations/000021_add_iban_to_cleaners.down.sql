-- Remove IBAN field from cleaners table
ALTER TABLE cleaners
DROP COLUMN IF EXISTS iban;
