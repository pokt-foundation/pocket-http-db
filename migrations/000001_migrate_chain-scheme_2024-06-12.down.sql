-- Remove the NOT NULL constraint from the scheme column
ALTER TABLE chains
ALTER COLUMN scheme DROP NOT NULL;
-- Remove the scheme column from the chains table
ALTER TABLE chains DROP COLUMN scheme;
-- Drop the enum type chain_scheme
DROP TYPE chain_scheme;
