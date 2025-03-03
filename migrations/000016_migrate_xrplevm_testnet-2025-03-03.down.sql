-- Undo XRPL EVM Testnet chain additions
BEGIN;

-- First, remove the chain-gigastake application mappings
DELETE FROM chains_gigastake_applications
WHERE chain_id = 'F036';

-- Remove sync checks
DELETE FROM chain_checks 
WHERE chain_id = 'F036';

-- Remove eth chain checks
DELETE FROM chain_checks 
WHERE chain_id = 'F036';

-- Remove chain configs
DELETE FROM chain_config 
WHERE chain_id = 'F036';

-- Remove chain aliases
DELETE FROM chain_alias_domains 
WHERE chain_id = 'F036';

-- Finally, remove the chain itself
DELETE FROM chains 
WHERE id = 'F036';

COMMIT;
