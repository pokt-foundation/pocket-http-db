-- This UP migration adds all aliases for the Ethereum chains to the new F-chain (F00C),
-- then disables the old chains (0021, 0028, 0022).

-- To execute on the local test Postgres database:
-- migrate -path migrations/ -database "postgresql://postgres:pgpassword@localhost:5432/postgres?sslmode=disable" -verbose up 8

-- Add new F-chain chain alias domains (in addition to 'eth')
INSERT into chain_alias_domains (chain_id, alias, domains)
VALUES 
('F00C', 'eth-mainnet', ARRAY['eth-mainnet.gateway.grove.city']),
('F00C', 'eth-trace', ARRAY['eth-trace.gateway.grove.city']),
('F00C', 'eth-archival', ARRAY['eth-archival.gateway.grove.city']);

-- Delete old chain aliases
DELETE FROM chain_alias_domains WHERE chain_id = '0021';
DELETE FROM chain_alias_domains WHERE chain_id = '0028';
DELETE FROM chain_alias_domains WHERE chain_id = '0022';

-- Disable old chains
UPDATE chains SET active = false WHERE id = '0021'; -- disable eth-mainnet
UPDATE chains SET active = false WHERE id = '0028'; -- disable eth-trace
UPDATE chains SET active = false WHERE id = '0022'; -- disable eth-archival
