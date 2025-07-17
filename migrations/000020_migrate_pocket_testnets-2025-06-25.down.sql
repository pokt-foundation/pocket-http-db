-- Undo Pocket Testnets (Alpha, Beta)
BEGIN;

-- Remove chain aliases
DELETE FROM chain_alias_domains 
WHERE chain_id in ('1235', '1236');

-- Finally, remove the chain itself
DELETE FROM chains 
WHERE id in ('1235', '1236');

COMMIT;
