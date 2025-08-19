-- Undo Chain Addition
BEGIN;

-- Remove chain aliases
DELETE FROM chain_alias_domains 
WHERE chain_id in ('1237', '1238', '1239', '1240', '1241', '1242', '1243', '1244');

-- Finally, remove the chains
DELETE FROM chains 
WHERE id in ('1237', '1238', '1239', '1240', '1241', '1242', '1243', '1244');

COMMIT;
