-- Rollback script for February 2025 chain updates
BEGIN;

-- Remove chain_id <-> gigastakes mappings
DELETE FROM chains_gigastake_applications
WHERE (chain_id, gigastake_application_id) IN (
    ('F035', 'd8b9739d'),
    ('F035', 'c71280d7'),
    ('F035', '149199e3')
);

-- Remove sync checks
DELETE FROM chain_checks 
WHERE chain_id = 'F035'; 

-- Remove eth chain checks
DELETE FROM chain_checks 
WHERE chain_id = 'F035'; 

-- Remove chain configs
DELETE FROM chain_config 
WHERE chain_id = 'F035';

-- Remove chain aliases
DELETE FROM chain_alias_domains 
WHERE chain_id = 'F035';

-- Revert alias updates back to original values
UPDATE chain_alias_domains
SET 
    alias = 'berachain',
    domains = '{berachain.gateway.grove.city}'
WHERE alias = 'berachain-bartio-testnet';

UPDATE chain_alias_domains
SET 
    alias = 'xrpl-evm',
    domains = '{xrpl-evm.gateway.grove.city}'
WHERE alias = 'xrpl-evm-devnet';

-- Remove the new chain
DELETE FROM chains 
WHERE id = 'F035';

COMMIT;
