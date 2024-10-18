-- Enums
CREATE TYPE auth_provider AS ENUM ('auth0');
CREATE TYPE auth_type AS ENUM ('auth0_github', 'auth0_username', 'auth0_google');
CREATE TYPE chain_auth_type AS ENUM ('basic_auth', 'bearer_token', 'none');
CREATE TYPE chain_check_type AS ENUM ('archival', 'chain', 'merge', 'sync');
CREATE TYPE environment AS ENUM (
    'production',
    'canary',
    'staging',
    'test',
    'local'
);
CREATE TYPE notification_event AS ENUM (
    'full',
    'half',
    'quarter',
    'signedUp',
    'threeQuarters'
);
CREATE TYPE notification_type AS ENUM ('email', 'portal', 'webhook');
CREATE TYPE permissions AS ENUM (
    'read:endpoint',
    'write:endpoint',
    'delete:endpoint',
    'transfer:endpoint'
);
CREATE TYPE whitelist_type AS ENUM (
    'blockchains',
    'contracts',
    'methods',
    'origins',
    'userAgents'
);
CREATE TYPE validator_method_str AS ENUM ('blockRange', 'logLimitBlocks', 'ethAddress');
CREATE TYPE protocol_id AS ENUM (
    'morse-mainnet',
    'shannon-mainnet',
    'shannon-testnet'
);
CREATE TYPE chain_scheme AS ENUM ('https', 'wss');
CREATE TYPE billing_type AS ENUM ('stripe', 'google');
-- Plans Tables
CREATE TABLE pay_plans (
    name VARCHAR(255),
    description VARCHAR,
    plan_type VARCHAR(25) PRIMARY KEY,
    chain_ids VARCHAR(4) ARRAY,
    monthly_relay_limit INT NOT NULL,
    throughput_limit INT NOT NULL,
    application_limit INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    daily_limit INT
);
-- Users Tables
CREATE TABLE users (
    id VARCHAR(10) PRIMARY KEY,
    email VARCHAR(255) NULL UNIQUE,
    signed_up BOOLEAN NOT NULL,
    icon_url VARCHAR,
    updates_product BOOLEAN,
    updates_marketing BOOLEAN,
    beta_tester BOOLEAN,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE user_auth_providers (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(10) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type auth_type NOT NULL,
    provider auth_provider NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    federated BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, type)
);
CREATE TABLE user_roles (
    role_name VARCHAR(25) PRIMARY KEY,
    permissions permissions ARRAY NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Accounts Tables
CREATE TABLE accounts (
    id VARCHAR(10) PRIMARY KEY,
    name VARCHAR(255),
    billing_type billing_type NOT NULL DEFAULT 'stripe',
    gcp_account_id VARCHAR(255),
    gcp_entitlement_id VARCHAR(255),
    icon_url VARCHAR,
    plan_type VARCHAR(25) NOT NULL REFERENCES pay_plans(plan_type),
    partner_chain_ids VARCHAR(4) ARRAY,
    partner_throughput_limit INT,
    partner_application_limit INT,
    monthly_user_limit INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted BOOLEAN NOT NULL DEFAULT false,
    deleted_at TIMESTAMPTZ NULL,
    -- Non-zero monthly user limit must be a multiple of 1M
    CONSTRAINT monthly_user_limit_check CHECK (
        plan_type = 'PLAN_UNLIMITED'
        OR monthly_user_limit IS NULL
        OR monthly_user_limit = 0
        OR monthly_user_limit % 1000000 = 0
    ),
    CONSTRAINT gcp_account_id_check CHECK (
        (billing_type = 'google' AND gcp_account_id IS NOT NULL AND gcp_entitlement_id IS NOT NULL) 
        OR
        (billing_type != 'google' AND gcp_account_id IS NULL AND gcp_entitlement_id IS NULL)
    )
);
CREATE TABLE IF NOT EXISTS account_integrations (
    id INT GENERATED ALWAYS AS IDENTITY,
    account_id VARCHAR(10) NOT NULL UNIQUE REFERENCES accounts(id) ON DELETE CASCADE,
    stripe_subscription_id VARCHAR(255) UNIQUE,
    covalent_api_key_free VARCHAR(255) UNIQUE,
    covalent_api_key_paid VARCHAR(255) UNIQUE,
    created_at TIMESTAMPTZ NULL,
    updated_at TIMESTAMPTZ NULL,
    PRIMARY KEY (id)
);
CREATE TABLE account_users (
    id SERIAL PRIMARY KEY,
    user_id VARCHAR(10) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id VARCHAR(10) NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    role_name VARCHAR(25) NOT NULL REFERENCES user_roles(role_name),
    accepted BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, account_id)
);
CREATE UNIQUE INDEX idx_unique_owner_per_account ON account_users (account_id)
WHERE role_name = 'OWNER';
CREATE TABLE account_notifications (
    id SERIAL PRIMARY KEY,
    account_id VARCHAR(24) NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    active BOOLEAN NOT NULL,
    type notification_type NOT NULL,
    destination VARCHAR(255),
    trigger VARCHAR(255),
    events VARCHAR(100) ARRAY,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (account_id, type)
);
-- Chains Tables
CREATE TABLE chains (
    id VARCHAR(4) PRIMARY KEY,
    blockchain VARCHAR(100) NOT NULL,
    description VARCHAR(100) NOT NULL,
    enforce_result VARCHAR(4) NOT NULL,
    ticker VARCHAR(100) NOT NULL,
    path VARCHAR(100),
    icon_url VARCHAR,
    request_timeout INT,
    node_request_timeout INT,
    log_limit_blocks INT,
    allowed_methods VARCHAR(10) ARRAY,
    active BOOLEAN NOT NULL DEFAULT false,
    scheme chain_scheme NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted BOOLEAN DEFAULT false,
    deleted_at TIMESTAMPTZ NULL
);
CREATE TABLE chain_altruists (
    id SERIAL PRIMARY KEY,
    chain_id VARCHAR(4) NOT NULL REFERENCES chains(id) ON DELETE CASCADE,
    url VARCHAR(255) NOT NULL,
    auth VARCHAR(100),
    auth_type chain_auth_type NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (chain_id, url)
);
CREATE TABLE chain_checks (
    id SERIAL PRIMARY KEY,
    chain_id VARCHAR(4) NOT NULL REFERENCES chains(id) ON DELETE CASCADE,
    type chain_check_type NOT NULL,
    payload VARCHAR(255),
    result_key VARCHAR(100),
    allowance INT,
    evm_chain_id INT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (chain_id, type),
    CONSTRAINT sync_allowance_check CHECK (
        (
            type = 'sync'
            AND allowance IS NOT NULL
        )
        OR (
            type <> 'sync'
            AND allowance IS NULL
        )
    ),
    CONSTRAINT chain_id_check CHECK (
        (
            type = 'chain'
            AND evm_chain_id IS NOT NULL
        )
        OR (
            type <> 'chain'
            AND evm_chain_id IS NULL
        )
    )
);
-- DEPRECATED - TODO remove when move to only store aliases is complete
CREATE TABLE IF NOT EXISTS chain_alias_domains (
    chain_id VARCHAR(4) NOT NULL REFERENCES chains(id) ON DELETE CASCADE,
    alias VARCHAR(255) NOT NULL,
    domains VARCHAR(100) ARRAY NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY(chain_id, alias)
);
CREATE TABLE chain_aliases (
    chain_id VARCHAR(4) REFERENCES chains(id) ON DELETE CASCADE,
    alias VARCHAR(100) NOT NULL CHECK (alias <> ''),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (chain_id, alias)
);
-- Portal Application Tables
CREATE TABLE portal_applications (
    id VARCHAR(24) PRIMARY KEY UNIQUE,
    account_id VARCHAR(10) REFERENCES accounts(id),
    name VARCHAR(255) NOT NULL,
    app_emoji VARCHAR(255),
    description VARCHAR,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted BOOLEAN NOT NULL DEFAULT false,
    deleted_at TIMESTAMPTZ NULL
);
CREATE TABLE portal_application_settings (
    id SERIAL PRIMARY KEY,
    application_id VARCHAR(24) NOT NULL UNIQUE REFERENCES portal_applications(id) ON DELETE CASCADE,
    monthly_relay_limit INT NOT NULL,
    environment environment NOT NULL,
    favorited_chain_ids VARCHAR(4) ARRAY,
    secret_key VARCHAR(64),
    secret_key_required BOOLEAN,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE portal_application_whitelists (
    id SERIAL PRIMARY KEY,
    application_id VARCHAR(24) NOT NULL REFERENCES portal_applications(id) ON DELETE CASCADE,
    type whitelist_type NOT NULL,
    value VARCHAR(255) NOT NULL,
    chain_id VARCHAR(4),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (application_id, value, type, chain_id),
    CONSTRAINT check_blockchain_id_for_methods_contracts CHECK (
        (
            type NOT IN ('methods', 'contracts')
            AND chain_id IS NULL
        )
        OR (
            type IN ('methods', 'contracts')
            AND chain_id IS NOT NULL
        )
    )
);
CREATE UNIQUE INDEX portal_application_whitelists_null_chain_idx ON portal_application_whitelists (application_id, value, type)
WHERE chain_id IS NULL;
CREATE TABLE IF NOT EXISTS portal_application_aats (
    id VARCHAR(24) PRIMARY KEY,
    portal_application_id VARCHAR(24) NOT NULL REFERENCES portal_applications(id) ON DELETE CASCADE,
    address VARCHAR(40) NOT NULL,
    public_key VARCHAR(64) NOT NULL,
    client_public_key VARCHAR(64) NOT NULL,
    signature VARCHAR(128) NOT NULL,
    private_key VARCHAR(400) NULL,
    version VARCHAR(10) NOT NULL
);
-- Protocols Table
CREATE TABLE protocols (
    id protocol_id PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted BOOLEAN NOT NULL DEFAULT false,
    deleted_at TIMESTAMPTZ NULL
);
-- Gigastake Applications Table
CREATE TABLE IF NOT EXISTS gigastake_applications (
    id VARCHAR(24) PRIMARY KEY,
    protocol_id protocol_id NOT NULL REFERENCES protocols(id),
    name VARCHAR(255) NOT NULL UNIQUE,
    address VARCHAR(40) NOT NULL UNIQUE,
    public_key VARCHAR(64) NOT NULL,
    client_public_key VARCHAR(64) NOT NULL,
    signature VARCHAR(128) NOT NULL,
    version VARCHAR(10) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted BOOLEAN NOT NULL DEFAULT false,
    deleted_at TIMESTAMPTZ NULL
);
-- Chains and Gigastake Applications Join Table
CREATE TABLE chains_gigastake_applications (
    chain_id VARCHAR(4) NOT NULL REFERENCES chains(id) ON DELETE CASCADE,
    gigastake_application_id VARCHAR(24) NOT NULL REFERENCES gigastake_applications(id) ON DELETE CASCADE,
    PRIMARY KEY(chain_id, gigastake_application_id)
);
-- Blocked Contracts Tables
CREATE TABLE global_blocked_contracts (
    id SERIAL PRIMARY KEY,
    blocked_address VARCHAR(255) NOT NULL UNIQUE,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- Gateway Config Tables
CREATE TABLE chain_config (
    id SERIAL PRIMARY KEY,
    chain_id VARCHAR(4) REFERENCES chains(id),
    valid_range INT,
    log_limit_blocks INT,
    archival_chain_id VARCHAR(4) NULL REFERENCES chains(id),
    UNIQUE (chain_id)
);
CREATE TABLE method_validators (
    id SERIAL PRIMARY KEY,
    method VARCHAR(255) NOT NULL,
    validator validator_method_str,
    UNIQUE (method, validator)
);
CREATE TABLE config_openapi (
    id SERIAL PRIMARY KEY,
    chain_id VARCHAR(4) REFERENCES chains(id),
    open_api_spec JSONB NOT NULL,
    UNIQUE (chain_id)
);
CREATE TABLE config_gateway (
    version INT NOT NULL,
    environment environment NOT NULL,
    description VARCHAR NOT NULL,
    configuration JSONB NOT NULL,
    PRIMARY KEY (version, environment)
);
-- Auto Increment Trigger Function
CREATE OR REPLACE FUNCTION increment_version() RETURNS TRIGGER AS $$
DECLARE latest_version INT;
BEGIN -- Find the latest version for the current environment
SELECT COALESCE(MAX(version), 0) INTO latest_version
FROM config_gateway
WHERE environment = NEW.environment;
-- Increment the version for the new row
NEW.version := latest_version + 1;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- Create Trigger
CREATE TRIGGER before_insert_config_gateway BEFORE
INSERT ON config_gateway FOR EACH ROW EXECUTE FUNCTION increment_version();
-- Do not allow updates to config_gateway table
CREATE OR REPLACE RULE config_gateway_no_update AS ON UPDATE TO config_gateway DO INSTEAD NOTHING;
-- Do not allow inserts if the config is the same as the latest config for the same environment, regardless of order
CREATE OR REPLACE FUNCTION check_json_change() RETURNS TRIGGER AS $$
DECLARE latest_config JSONB;
BEGIN
SELECT configuration INTO latest_config
FROM config_gateway
WHERE environment = NEW.environment
ORDER BY version DESC
LIMIT 1;
-- Use the JSONB equality operator to compare the JSONB objects directly
IF (
    jsonb_strip_nulls(NEW.configuration) = jsonb_strip_nulls(latest_config)
) THEN RAISE EXCEPTION 'Blocking config_gateway row insert: Gateway configuration JSON has not changed for the environment %',
NEW.environment;
END IF;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER config_gateway_check_change BEFORE
INSERT ON config_gateway FOR EACH ROW EXECUTE PROCEDURE check_json_change();
-- Listener Notification Function
CREATE OR REPLACE FUNCTION notify_event() RETURNS TRIGGER AS $$
DECLARE data json;
notification json;
BEGIN -- Convert the old or new row to JSON, based on the kind of action.
-- Action = DELETE?             -> OLD row
-- Action = INSERT or UPDATE?   -> NEW row
IF (TG_OP = 'DELETE') THEN data = row_to_json(OLD);
ELSE data = row_to_json(NEW);
END IF;
-- Contruct the notification as a JSON string.
notification = json_build_object(
    'table',
    TG_TABLE_NAME,
    'action',
    TG_OP,
    'data',
    data
);
-- Execute pg_notify(channel, notification)
PERFORM pg_notify('events', notification::text);
-- Result is ignored since this is an AFTER trigger
RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- Listener Notification Triggers
CREATE TRIGGER portal_applications_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON portal_applications FOR EACH ROW EXECUTE PROCEDURE notify_event();
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
CREATE TRIGGER portal_application_aats_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON portal_application_aats FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER gigastake_applications_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON gigastake_applications FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER accounts_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON accounts FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER account_users_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON account_users FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER account_integrations_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON account_integrations FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER account_notifications_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON account_notifications FOR EACH ROW EXECUTE PROCEDURE notify_event();
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
CREATE TRIGGER chain_aliases_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON chain_aliases FOR EACH ROW EXECUTE PROCEDURE notify_event();
-- DEPRECATED - TODO remove when move to only store aliases is complete
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
CREATE TRIGGER chain_config_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON chain_config FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER method_validators_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON method_validators FOR EACH ROW EXECUTE PROCEDURE notify_event();
CREATE TRIGGER config_gateway_notify_event
AFTER
INSERT
    OR
UPDATE
    OR DELETE ON config_gateway FOR EACH ROW EXECUTE PROCEDURE notify_event();
-- Create a dummy portalapprw user
CREATE ROLE portalapprw WITH LOGIN PASSWORD 'password_123';
GRANT SELECT,
    INSERT,
    UPDATE,
    DELETE ON ALL TABLES IN SCHEMA public TO portalapprw;
GRANT USAGE,
    SELECT ON ALL SEQUENCES IN SCHEMA public TO portalapprw;
