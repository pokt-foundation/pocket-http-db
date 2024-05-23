-- Migration SQL to revert changes made by the UP migration file
BEGIN;
-- Removing triggers for chain_override_sessions table
DROP TRIGGER IF EXISTS chain_override_sessions_notify_event ON chain_override_sessions;
-- Dropping new table chain_override_sessions
DROP TABLE IF EXISTS chain_override_sessions;
-- Removing the composite foreign key and 'override_chain_id' column from gigastake_applications table
ALTER TABLE gigastake_applications DROP CONSTRAINT IF EXISTS fk_override_chain_protocol_id;
ALTER TABLE gigastake_applications DROP COLUMN IF EXISTS override_chain_id;
COMMIT;
