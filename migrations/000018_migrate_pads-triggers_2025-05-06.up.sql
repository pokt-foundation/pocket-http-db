-- This file contains the triggers that listen for changes to the Grove Portal DB and stream updates to PEAS over gRPC.

-- These triggers are used to listen for changes to Portal Applications and their associated tables.
-- When updates are detected, the triggers insert a row into the `portal_application_changes` table.
-- The `log_portal_application_changes` function is then called to handle the update.
-- The function sends a minimal notification to the `portal_application_changes` channel, which is handled by the `portalApplicationChangesChannel` in the Postgres data source.
-- See implementation here: https://github.com/buildwithgrove/path-auth-data-server/blob/main/postgres/grove/data_source.go#L163
-- PADS then streams the updated data to PEAS over gRPC, which updates its Gateway Endpoints store used to authorize requests to PATH.

-- /*-------------------- Listener Updates --------------------*/

-- Create the changes table with an 'is_delete' field
CREATE TABLE portal_application_changes (
    id SERIAL PRIMARY KEY,
    portal_app_id VARCHAR(24) NOT NULL,
    is_delete BOOLEAN NOT NULL DEFAULT FALSE,
    changed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create the trigger function with 'is_delete' handling
CREATE OR REPLACE FUNCTION log_portal_application_changes() RETURNS trigger AS $$
DECLARE
    portal_app_ids TEXT[];
    is_delete BOOLEAN := FALSE;
BEGIN
    portal_app_ids := ARRAY[]::TEXT[];

    IF TG_TABLE_NAME = 'portal_applications' THEN
        IF TG_OP = 'DELETE' THEN
            is_delete := TRUE;
            portal_app_ids := array_append(portal_app_ids, OLD.id);
        ELSIF TG_OP = 'UPDATE' AND NEW.deleted = true THEN
            is_delete := TRUE;
            portal_app_ids := array_append(portal_app_ids, NEW.id);
        ELSE
            portal_app_ids := array_append(portal_app_ids, NEW.id);
        END IF;

    ELSIF TG_TABLE_NAME = 'portal_application_settings' THEN
        SELECT array_agg(pa.id) INTO portal_app_ids
        FROM portal_applications pa
        WHERE pa.id = COALESCE(NEW.application_id, OLD.application_id);

    ELSIF TG_TABLE_NAME = 'accounts' THEN
        SELECT array_agg(pa.id) INTO portal_app_ids
        FROM portal_applications pa
        WHERE pa.account_id = COALESCE(NEW.id, OLD.id);

    ELSIF TG_TABLE_NAME = 'pay_plans' THEN
        SELECT array_agg(pa.id) INTO portal_app_ids
        FROM portal_applications pa
        JOIN accounts a ON pa.account_id = a.id
        WHERE a.plan_type = COALESCE(NEW.plan_type, OLD.plan_type);
    END IF;

    -- Remove duplicates
    SELECT ARRAY(SELECT DISTINCT unnest(portal_app_ids)) INTO portal_app_ids;

    -- Insert into changes table with 'is_delete' flag
    IF array_length(portal_app_ids, 1) > 0 THEN
        INSERT INTO portal_application_changes (portal_app_id, is_delete)
        SELECT unnest(portal_app_ids), is_delete;
    END IF;

    -- Send minimal notification
    PERFORM pg_notify('portal_application_changes', '');

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create triggers for each table

CREATE TRIGGER portal_applications_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON portal_applications
FOR EACH ROW EXECUTE FUNCTION log_portal_application_changes();

CREATE TRIGGER portal_application_settings_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON portal_application_settings
FOR EACH ROW EXECUTE FUNCTION log_portal_application_changes();

CREATE TRIGGER accounts_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON accounts
FOR EACH ROW EXECUTE FUNCTION log_portal_application_changes();

CREATE TRIGGER pay_plans_change_trigger
AFTER INSERT OR UPDATE OR DELETE ON pay_plans
FOR EACH ROW EXECUTE FUNCTION log_portal_application_changes();
