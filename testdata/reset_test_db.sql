-- Drop all tables, sequences, and constraints
DO $$
DECLARE r RECORD;
BEGIN FOR r IN (
    SELECT tablename
    FROM pg_tables
    WHERE schemaname = current_schema()
) LOOP EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
END LOOP;
FOR r IN (
    SELECT sequence_name
    FROM information_schema.sequences
    WHERE sequence_schema = current_schema()
) LOOP EXECUTE 'DROP SEQUENCE IF EXISTS ' || quote_ident(r.sequence_name) || ' CASCADE';
END LOOP;

END $$;
DROP TYPE IF EXISTS auth_provider CASCADE;
DROP TYPE IF EXISTS auth_type CASCADE;
DROP TYPE IF EXISTS chain_auth_type CASCADE;
DROP TYPE IF EXISTS chain_check_type CASCADE;
DROP TYPE IF EXISTS environment CASCADE;
DROP TYPE IF EXISTS notification_event CASCADE;
DROP TYPE IF EXISTS notification_type CASCADE;
DROP TYPE IF EXISTS permissions CASCADE;
DROP TYPE IF EXISTS whitelist_type CASCADE;

-- Reset Schema
\i /docker-entrypoint-initdb.d/schema.sql;
-- Drop Triggers to Avoid Breaking PHD Listener
DROP TRIGGER IF EXISTS portal_applications_notify_event ON portal_applications;
DROP TRIGGER IF EXISTS gigastake_applications_notify_event ON gigastake_applications;
DROP TRIGGER IF EXISTS portal_application_settings_notify_event ON portal_application_settings;
DROP TRIGGER IF EXISTS portal_application_whitelists_notify_event ON portal_application_whitelists;
DROP TRIGGER IF EXISTS portal_application_notifications_notify_event ON portal_application_notifications;
DROP TRIGGER IF EXISTS portal_application_aats_notify_event ON portal_application_aats;
DROP TRIGGER IF EXISTS accounts_notify_event ON accounts;
DROP TRIGGER IF EXISTS account_user_access_notify_event ON account_user_access;
DROP TRIGGER IF EXISTS users_notify_event ON users;
DROP TRIGGER IF EXISTS user_auth_providers_notify_event ON user_auth_providers;
DROP TRIGGER IF EXISTS user_roles_notify_event ON user_roles;
DROP TRIGGER IF EXISTS pay_plans_notify_event ON pay_plans;
DROP TRIGGER IF EXISTS chains_notify_event ON chains;
DROP TRIGGER IF EXISTS chain_altruists_notify_event ON chain_altruists;
DROP TRIGGER IF EXISTS chain_alias_domains_notify_event ON chain_alias_domains;
DROP TRIGGER IF EXISTS chain_checks_notify_event ON chain_checks;
DROP TRIGGER IF EXISTS chains_gigastake_applications_notify_event ON chains_gigastake_applications;
DROP TRIGGER IF EXISTS global_blocked_contracts_notify_event ON global_blocked_contracts;
-- Seed Test DB
\i /docker-entrypoint-initdb.d/seed_test_db.sql;
-- Recreate Triggers
CREATE TRIGGER portal_applications_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON portal_applications FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER gigastake_applications_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON gigastake_applications FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER portal_application_aats_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON portal_application_aats FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER portal_application_settings_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON portal_application_settings FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER portal_application_whitelists_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON portal_application_whitelists FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER portal_application_notifications_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON portal_application_notifications FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER accounts_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON accounts FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER account_user_access_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON account_user_access FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER users_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON users FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER user_auth_providers_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON user_auth_providers FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER user_roles_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON user_roles FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER pay_plans_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON pay_plans FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER chains_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON chains FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER chain_altruists_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON chain_altruists FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER chain_alias_domains_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON chain_alias_domains FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER chain_checks_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON chain_checks FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER chains_gigastake_applications_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON chains_gigastake_applications FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER global_blocked_contracts_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON global_blocked_contracts FOR EACH ROW EXECUTE PROCEDURE notify_event();
