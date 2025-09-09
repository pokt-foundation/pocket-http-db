-- Rollback: Remove GIWA Services
BEGIN;

-- First remove chain aliases to maintain referential integrity
DELETE FROM chain_alias_domains
WHERE chain_id IN ('1265','1266');

-- Then remove the chains themselves
DELETE FROM chains
WHERE id IN ('1265','1266');

COMMIT;
