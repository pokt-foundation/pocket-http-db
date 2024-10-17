CREATE TYPE billing_type AS ENUM ('stripe', 'google');

ALTER TABLE accounts
    ADD COLUMN billing_type billing_type;

ALTER TABLE accounts
    ADD COLUMN gcp_account_id VARCHAR(255);

ALTER TABLE accounts
    ADD COLUMN gcp_entitlement_id VARCHAR(255);

UPDATE accounts
    SET billing_type = 'stripe';

ALTER TABLE accounts
    ALTER COLUMN billing_type SET NOT NULL,
    ALTER COLUMN billing_type SET DEFAULT 'stripe';

ALTER TABLE accounts
    ADD CONSTRAINT gcp_account_id_check CHECK (
        (billing_type = 'google' AND gcp_account_id IS NOT NULL AND gcp_entitlement_id IS NOT NULL) 
        OR
        (billing_type != 'google' AND gcp_account_id IS NULL AND gcp_entitlement_id IS NULL)
    );
