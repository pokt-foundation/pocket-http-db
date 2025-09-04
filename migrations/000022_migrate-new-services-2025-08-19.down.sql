-- Rollback: Remove Kleomedes Cosmos chains and Hey service
BEGIN;

-- First remove chain aliases to maintain referential integrity
DELETE FROM chain_alias_domains
WHERE chain_id IN ('1237', '1238', '1239', '1240', '1241', '1242', '1243', '1244', '1245', '1246',
                   '1247', '1248', '1249', '1250', '1251', '1252', '1253', '1254', '1255', '1256',
                   '1257', '1258', '1259', '1260', '1261', '1262', '1263', '1264');

-- Then remove the chains themselves
DELETE FROM chains
WHERE id IN ('1237', '1238', '1239', '1240', '1241', '1242', '1243', '1244', '1245', '1246',
             '1247', '1248', '1249', '1250', '1251', '1252', '1253', '1254', '1255', '1256',
             '1257', '1258', '1259', '1260', '1261', '1262', '1263', '1264');

COMMIT;
