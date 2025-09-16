-- Rollback: Remove new services
BEGIN;

-- First remove chain aliases to maintain referential integrity
DELETE FROM chain_alias_domains
WHERE chain_id IN ('1267', '1268', '1269', '1270', '1271', '1272', '1273', '1274', '1275', '1276', '1277', '1278', '1279');

-- Then remove the chains themselves
DELETE FROM chains
WHERE id IN ('1267', '1268', '1269', '1270', '1271', '1272', '1273', '1274', '1275', '1276', '1277', '1278', '1279');

COMMIT;
