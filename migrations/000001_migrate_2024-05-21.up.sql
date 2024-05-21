-- Migration SQL to transition from schema.sql to new-schema.sql with the specified changes
-- Adding new table for chain_override_sessions
CREATE TABLE chain_override_sessions (
    chain_id VARCHAR(4) NOT NULL REFERENCES chains(id) ON DELETE CASCADE,
    protocol_id protocol_id NOT NULL,
    gigastake_application_id VARCHAR(64) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    node_urls VARCHAR(255) ARRAY NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (chain_id, protocol_id),
    CHECK (array_length(node_urls, 1) = 24)
);
-- Adding triggers for the new table
CREATE TRIGGER chain_override_sessions_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON chain_override_sessions FOR EACH ROW EXECUTE PROCEDURE notify_event();
-- Adding new field 'override' to gigastake_applications table
ALTER TABLE gigastake_applications
ADD COLUMN override BOOLEAN NOT NULL DEFAULT false;
