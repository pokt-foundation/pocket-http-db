-- This migration improves the PADS triggers with better error handling, logging, validation, and batching
-- to prevent 401s and 404s in downstream PEAS applications.

-- Create a trigger logs table for monitoring and debugging
CREATE TABLE IF NOT EXISTS pads_trigger_logs (
    id SERIAL PRIMARY KEY,
    trigger_name VARCHAR(255) NOT NULL,
    table_name VARCHAR(255) NOT NULL,
    operation VARCHAR(10) NOT NULL,
    portal_app_ids TEXT[],
    error_message TEXT,
    execution_time_ms INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_pads_trigger_logs_created_at ON pads_trigger_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_pads_trigger_logs_trigger_name ON pads_trigger_logs(trigger_name);

-- Create a batch tracking table for notification management
CREATE TABLE IF NOT EXISTS pads_batch_tracker (
    id SERIAL PRIMARY KEY,
    last_notification_sent TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    batch_size INTEGER DEFAULT 0,
    next_batch_time TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert initial batch tracker row
INSERT INTO pads_batch_tracker (id, last_notification_sent, batch_size, next_batch_time) 
VALUES (1, CURRENT_TIMESTAMP, 0, CURRENT_TIMESTAMP)
ON CONFLICT (id) DO NOTHING;

-- Create function to handle batched notifications
CREATE OR REPLACE FUNCTION send_batched_notification() RETURNS VOID AS $$
DECLARE
    batch_interval INTEGER := 1; -- seconds between batches
    max_batch_size INTEGER := 100; -- max changes before forcing notification
    current_batch_size INTEGER;
    last_sent TIMESTAMP WITH TIME ZONE;
    should_send BOOLEAN := FALSE;
BEGIN
    -- Get current batch status
    SELECT batch_size, last_notification_sent 
    INTO current_batch_size, last_sent
    FROM pads_batch_tracker 
    WHERE id = 1;

    -- Check if we should send notification based on time or batch size
    IF current_batch_size >= max_batch_size 
       OR (current_batch_size > 0 AND CURRENT_TIMESTAMP - last_sent >= INTERVAL '1 second' * batch_interval) THEN
        should_send := TRUE;
    END IF;

    IF should_send THEN
        -- Send the notification
        PERFORM pg_notify('portal_application_changes', '');
        
        -- Reset the batch tracker
        UPDATE pads_batch_tracker 
        SET last_notification_sent = CURRENT_TIMESTAMP,
            batch_size = 0,
            next_batch_time = CURRENT_TIMESTAMP + INTERVAL '1 second' * batch_interval
        WHERE id = 1;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Create a background job function to ensure pending notifications are sent
-- This should be called periodically (e.g., every 5 seconds) by your application
CREATE OR REPLACE FUNCTION flush_pending_notifications() RETURNS INTEGER AS $$
DECLARE
    pending_count INTEGER;
BEGIN
    SELECT batch_size INTO pending_count
    FROM pads_batch_tracker 
    WHERE id = 1;
    
    IF pending_count > 0 THEN
        PERFORM send_batched_notification();
        RETURN pending_count;
    END IF;
    
    RETURN 0;
END;
$$ LANGUAGE plpgsql;

-- Create a function to clean up old trigger logs (for maintenance)
CREATE OR REPLACE FUNCTION cleanup_pads_trigger_logs(days_to_keep INTEGER DEFAULT 30) RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    DELETE FROM pads_trigger_logs 
    WHERE created_at < CURRENT_TIMESTAMP - INTERVAL '1 day' * days_to_keep;
    
    GET DIAGNOSTICS deleted_count = ROW_COUNT;
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Improved trigger function with error handling, logging, validation, and batching
CREATE OR REPLACE FUNCTION log_portal_application_changes() RETURNS trigger AS $$
DECLARE
    portal_app_ids TEXT[];
    is_delete BOOLEAN := FALSE;
    start_time TIMESTAMP WITH TIME ZONE;
    end_time TIMESTAMP WITH TIME ZONE;
    execution_time INTEGER;
    log_id INTEGER;
    error_msg TEXT;
BEGIN
    start_time := CURRENT_TIMESTAMP;
    portal_app_ids := ARRAY[]::TEXT[];

    -- Start logging
    INSERT INTO pads_trigger_logs (trigger_name, table_name, operation, portal_app_ids)
    VALUES (TG_NAME, TG_TABLE_NAME, TG_OP, portal_app_ids)
    RETURNING id INTO log_id;

    BEGIN
        IF TG_TABLE_NAME = 'portal_applications' THEN
            IF TG_OP = 'DELETE' THEN
                -- Validate OLD record exists
                IF OLD.id IS NULL THEN
                    RAISE EXCEPTION 'DELETE operation has NULL portal application ID';
                END IF;
                is_delete := TRUE;
                portal_app_ids := array_append(portal_app_ids, OLD.id);
            ELSIF TG_OP = 'UPDATE' AND NEW.deleted = true THEN
                -- Validate NEW record exists
                IF NEW.id IS NULL THEN
                    RAISE EXCEPTION 'UPDATE operation has NULL portal application ID';
                END IF;
                is_delete := TRUE;
                portal_app_ids := array_append(portal_app_ids, NEW.id);
            ELSE
                -- Validate NEW record exists for INSERT/UPDATE
                IF NEW.id IS NULL THEN
                    RAISE EXCEPTION 'INSERT/UPDATE operation has NULL portal application ID';
                END IF;
                portal_app_ids := array_append(portal_app_ids, NEW.id);
            END IF;

        ELSIF TG_TABLE_NAME = 'portal_application_settings' THEN
            -- Validate application_id exists
            IF COALESCE(NEW.application_id, OLD.application_id) IS NULL THEN
                RAISE EXCEPTION 'portal_application_settings operation has NULL application_id';
            END IF;
            
            -- Validate that the referenced portal application exists
            SELECT array_agg(pa.id) INTO portal_app_ids
            FROM portal_applications pa
            WHERE pa.id = COALESCE(NEW.application_id, OLD.application_id)
              AND pa.deleted = false;
            
            -- Check if no valid applications found
            IF portal_app_ids IS NULL OR array_length(portal_app_ids, 1) = 0 THEN
                RAISE WARNING 'No valid portal applications found for application_id: %', 
                    COALESCE(NEW.application_id, OLD.application_id);
                portal_app_ids := ARRAY[]::TEXT[];
            END IF;

        ELSIF TG_TABLE_NAME = 'accounts' THEN
            -- Validate account ID exists
            IF COALESCE(NEW.id, OLD.id) IS NULL THEN
                RAISE EXCEPTION 'accounts operation has NULL account ID';
            END IF;
            
            -- Find all non-deleted portal applications for this account
            SELECT array_agg(pa.id) INTO portal_app_ids
            FROM portal_applications pa
            WHERE pa.account_id = COALESCE(NEW.id, OLD.id)
              AND pa.deleted = false;
            
            -- Handle case where account has no applications
            IF portal_app_ids IS NULL THEN
                portal_app_ids := ARRAY[]::TEXT[];
            END IF;

        ELSIF TG_TABLE_NAME = 'pay_plans' THEN
            -- Validate plan_type exists
            IF COALESCE(NEW.plan_type, OLD.plan_type) IS NULL THEN
                RAISE EXCEPTION 'pay_plans operation has NULL plan_type';
            END IF;
            
            -- Find all non-deleted portal applications affected by this plan change
            SELECT array_agg(pa.id) INTO portal_app_ids
            FROM portal_applications pa
            JOIN accounts a ON pa.account_id = a.id
            WHERE a.plan_type = COALESCE(NEW.plan_type, OLD.plan_type)
              AND pa.deleted = false;
            
            -- Handle case where plan has no applications
            IF portal_app_ids IS NULL THEN
                portal_app_ids := ARRAY[]::TEXT[];
            END IF;
        END IF;

        -- Remove duplicates and NULL values
        SELECT ARRAY(
            SELECT DISTINCT unnest(portal_app_ids) 
            WHERE unnest(portal_app_ids) IS NOT NULL
        ) INTO portal_app_ids;

        -- Insert into changes table with 'is_delete' flag only if we have valid app IDs
        IF array_length(portal_app_ids, 1) > 0 THEN
            INSERT INTO portal_application_changes (portal_app_id, is_delete)
            SELECT unnest(portal_app_ids), is_delete;
            
            -- Update batch tracker instead of sending immediate notification
            UPDATE pads_batch_tracker 
            SET batch_size = batch_size + array_length(portal_app_ids, 1)
            WHERE id = 1;
            
            -- Try to send batched notification (will only send if conditions are met)
            PERFORM send_batched_notification();
        END IF;

        -- Calculate execution time
        end_time := CURRENT_TIMESTAMP;
        execution_time := EXTRACT(EPOCH FROM (end_time - start_time)) * 1000;

        -- Update log with success
        UPDATE pads_trigger_logs 
        SET portal_app_ids = log_portal_application_changes.portal_app_ids,
            execution_time_ms = execution_time
        WHERE id = log_id;

    EXCEPTION WHEN OTHERS THEN
        -- Log the error
        error_msg := SQLERRM;
        end_time := CURRENT_TIMESTAMP;
        execution_time := EXTRACT(EPOCH FROM (end_time - start_time)) * 1000;
        
        UPDATE pads_trigger_logs 
        SET error_message = error_msg,
            execution_time_ms = execution_time,
            portal_app_ids = log_portal_application_changes.portal_app_ids
        WHERE id = log_id;
        
        -- Re-raise the exception to maintain trigger behavior
        RAISE;
    END;

    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Create a monitoring view for easy analysis
CREATE OR REPLACE VIEW pads_trigger_monitoring AS
SELECT 
    trigger_name,
    table_name,
    operation,
    COUNT(*) as execution_count,
    COUNT(CASE WHEN error_message IS NOT NULL THEN 1 END) as error_count,
    AVG(execution_time_ms) as avg_execution_time_ms,
    MAX(execution_time_ms) as max_execution_time_ms,
    MAX(created_at) as last_execution
FROM pads_trigger_logs
WHERE created_at >= CURRENT_TIMESTAMP - INTERVAL '24 hours'
GROUP BY trigger_name, table_name, operation
ORDER BY error_count DESC, avg_execution_time_ms DESC;