ALTER TABLE accounts
    DROP CONSTRAINT gcp_account_id_check;

ALTER TABLE accounts
    DROP COLUMN gcp_entitlement_id;

ALTER TABLE accounts
    DROP COLUMN gcp_account_id;

ALTER TABLE accounts
    DROP COLUMN billing_type;

DROP TYPE billing_type;
