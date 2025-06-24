-- Undo XRPL EVM Testnet chain additions
BEGIN;

-- Remove chain aliases
DELETE FROM chain_alias_domains 
WHERE chain_id = '1234';

-- Finally, remove the chain itself
DELETE FROM chains 
WHERE id = '1234';

COMMIT;
