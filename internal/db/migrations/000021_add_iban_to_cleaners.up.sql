-- Add IBAN field to cleaners table for payout transfers
ALTER TABLE cleaners
ADD COLUMN iban VARCHAR(34);

-- Add comment explaining the field
COMMENT ON COLUMN cleaners.iban IS 'Romanian IBAN for bank transfers (RO + 2 digits + 24 characters, total 26 chars for RO IBANs)';

-- Note: IBAN is nullable during initial rollout, but should be required before production payouts
-- Romanian IBAN format: ROkk BBBB SSSS CCCC CCCC CCCC (26 characters total)
-- Example: RO49 AAAA 1B31 0075 9384 0000
