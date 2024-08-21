ALTER TABLE accounts
ADD COLUMN monthly_user_limit INT,
    ADD CONSTRAINT monthly_user_limit_check CHECK (
        plan_type = 'PLAN_UNLIMITED'
        OR monthly_user_limit IS NULL
    );
