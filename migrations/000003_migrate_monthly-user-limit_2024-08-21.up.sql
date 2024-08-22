ALTER TABLE accounts
ADD COLUMN monthly_user_limit INT,
    ADD CONSTRAINT monthly_user_limit_check CHECK (
        plan_type = 'PLAN_UNLIMITED'
        OR monthly_user_limit IS NULL
        OR monthly_user_limit = 0
        OR monthly_user_limit % 1000000 = 0
    );
