BEGIN;

ALTER TABLE chains
DROP COLUMN websockets_enabled;

COMMIT;