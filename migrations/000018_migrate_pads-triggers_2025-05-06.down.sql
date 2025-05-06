-- Down migration to revert all changes made in the up migration

-- Drop the triggers first
DROP TRIGGER IF EXISTS portal_applications_change_trigger ON portal_applications;
DROP TRIGGER IF EXISTS portal_application_settings_change_trigger ON portal_application_settings;
DROP TRIGGER IF EXISTS accounts_change_trigger ON accounts;
DROP TRIGGER IF EXISTS pay_plans_change_trigger ON pay_plans;

-- Drop the trigger function
DROP FUNCTION IF EXISTS log_portal_application_changes();

-- Drop the changes table last
DROP TABLE IF EXISTS portal_application_changes;
