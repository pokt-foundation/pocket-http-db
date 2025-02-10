-- Rollback script for MEV Config Changes
BEGIN;

DELETE FROM config_gateway
WHERE version = 8
;

COMMIT;
