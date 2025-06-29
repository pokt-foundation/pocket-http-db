-- Rollback the PADS trigger improvements

-- Drop the monitoring view
DROP VIEW IF EXISTS pads_trigger_monitoring;

-- Drop the background job functions
DROP FUNCTION IF EXISTS flush_pending_notifications();
DROP FUNCTION IF EXISTS send_batched_notification();
DROP FUNCTION IF EXISTS cleanup_pads_trigger_logs(INTEGER);

-- Restore the original trigger function (from migration 000018)
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

-- Drop the new tables and indexes
DROP INDEX IF EXISTS idx_pads_trigger_logs_created_at;
DROP INDEX IF EXISTS idx_pads_trigger_logs_trigger_name;
DROP TABLE IF EXISTS pads_batch_tracker;
DROP TABLE IF EXISTS pads_trigger_logs;