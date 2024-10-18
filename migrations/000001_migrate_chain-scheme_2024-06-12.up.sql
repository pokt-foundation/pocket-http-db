-- Re-create the enum type for chain_scheme
CREATE TYPE chain_scheme AS ENUM ('https', 'wss');
-- Add the scheme column to the chains table, allowing NULL initially
ALTER TABLE chains
ADD COLUMN scheme chain_scheme;
-- Set default values for the newly added column
UPDATE chains
SET scheme = 'https';
-- Alter the scheme column to disallow NULL values
ALTER TABLE chains
ALTER COLUMN scheme
SET NOT NULL;
