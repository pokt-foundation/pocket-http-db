-- This DOWN migration reverses the changes made by the UP migration 000008_migrate_f-chains-pt-4_2024-11-13.up.sql.

-- To execute on the local test Postgres database:
-- migrate -path migrations/ -database "postgresql://postgres:pgpassword@localhost:5432/postgres?sslmode=disable" -verbose down 8

-- Delete F-chain Ethereum chain alias domains (except 'eth')
DELETE FROM chain_alias_domains WHERE chain_id = 'F00C' and alias IN ('eth-mainnet', 'eth-trace', 'eth-archival');

-- Disable F-chain Ethereum chain
UPDATE chains SET active = false WHERE id = 'F00C';

-- Re-create new chain aliases for the old chains
INSERT into chain_alias_domains (chain_id, alias, domains)
VALUES 
('0021', 'eth-mainnet', ARRAY['eth-mainnet.gateway.grove.city']),
('0028', 'eth-trace', ARRAY['eth-trace.gateway.grove.city']),
('0022', 'eth-archival', ARRAY['eth-archival.gateway.grove.city']);

-- Re-enable old chains
UPDATE chains SET active = true WHERE id = '0021'; -- re-enable eth-mainnet
UPDATE chains SET active = true WHERE id = '0028'; -- re-enable eth-trace
UPDATE chains SET active = true WHERE id = '0022'; -- re-enable eth-archival
