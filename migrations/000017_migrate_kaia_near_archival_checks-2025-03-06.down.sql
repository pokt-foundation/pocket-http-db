-- Undo Kaia and NEAR archival checks
BEGIN;

-- Remove newly added archival checks
DELETE FROM chain_checks
WHERE chain_id in ('F016','F01B')
AND type = 'archival';

COMMIT;
