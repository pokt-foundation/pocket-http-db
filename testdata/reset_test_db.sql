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

-- Seed Test DB
\i /docker-entrypoint-initdb.d/seed_test_db.sql;
