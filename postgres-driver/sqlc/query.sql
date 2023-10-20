-- name: CheckIDExists :one
SELECT EXISTS (
        SELECT 1
        FROM (
                SELECT id
                FROM portal_applications
                WHERE portal_applications.id = @id::VARCHAR
                UNION ALL
                SELECT id
                FROM portal_application_aats
                WHERE portal_application_aats.id = @id::VARCHAR
                UNION ALL
                SELECT id
                FROM gigastake_applications
                WHERE gigastake_applications.id = @id::VARCHAR
                UNION ALL
                SELECT id
                FROM accounts
                WHERE accounts.id = @id::VARCHAR
                UNION ALL
                SELECT id
                FROM users
                WHERE users.id = @id::VARCHAR
            ) AS id_table
    );
-- name: SelectPlans :many
SELECT name,
    description,
    plan_type,
    chain_ids,
    monthly_relay_limit,
    throughput_limit,
    application_limit,
    created_at,
    daily_limit
FROM pay_plans;
-- name: SelectGigastakeApplications :many
SELECT ga.id,
    ga.name,
    ARRAY_AGG(cga.chain_id)::VARCHAR [] AS chain_ids,
    ga.created_at,
    ga.updated_at,
    ga.deleted,
    ga.address,
    ga.public_key,
    ga.client_public_key,
    ga.signature,
    ga.version
FROM gigastake_applications AS ga
    LEFT JOIN chains_gigastake_applications AS cga ON ga.id = cga.gigastake_application_id
GROUP BY ga.id,
    ga.name,
    ga.created_at,
    ga.updated_at,
    ga.deleted,
    ga.address,
    ga.public_key,
    ga.client_public_key,
    ga.signature,
    ga.version;
-- name: SelectPortalApplications :many
WITH aats_agg AS (
    SELECT paa.portal_application_id,
        json_object_agg(
            paa.id,
            json_build_object(
                'address',
                paa.address,
                'public_key',
                paa.public_key,
                'client_public_key',
                paa.client_public_key,
                'signature',
                paa.signature,
                'version',
                paa.version
            )
        ) AS aats
    FROM portal_application_aats paa
    GROUP BY paa.portal_application_id
),
notifications_agg AS (
    SELECT pn.application_id,
        json_object_agg(
            pn.type,
            json_build_object(
                'type',
                pn.type,
                'active',
                pn.active,
                'destination',
                pn.destination,
                'trigger',
                pn.trigger,
                'events',
                (
                    SELECT json_object_agg(
                            event,
                            true
                        )
                    FROM (
                            SELECT unnest(pn.events) AS event
                        ) subquery
                    WHERE event IS NOT NULL
                )
            )
        ) AS notifications
    FROM portal_application_notifications pn
    GROUP BY pn.application_id
),
whitelists_agg AS (
    SELECT paw.application_id,
        json_agg(
            json_build_object(
                'type',
                paw.type,
                'value',
                paw.value,
                'chain_id',
                paw.chain_id
            )
        ) AS whitelists
    FROM portal_application_whitelists paw
    GROUP BY paw.application_id
)
SELECT p.*,
    pas.secret_key,
    pas.secret_key_required,
    pas.monthly_relay_limit,
    pas.environment,
    pas.favorited_chain_ids,
    COALESCE(aats_agg.aats, '[]'::json) AS aats,
    COALESCE(notifications_agg.notifications, '[]'::json) AS notifications,
    COALESCE(whitelists_agg.whitelists, '[]'::json) AS whitelists
FROM portal_applications p
    LEFT JOIN portal_application_settings pas ON p.id = pas.application_id
    LEFT JOIN aats_agg ON p.id = aats_agg.portal_application_id
    LEFT JOIN notifications_agg ON p.id = notifications_agg.application_id
    LEFT JOIN whitelists_agg ON p.id = whitelists_agg.application_id
WHERE (
        $1::BOOLEAN
        OR p.deleted = false
    );
-- name: InsertPortalApplication :one
INSERT INTO portal_applications (
        id,
        account_id,
        name,
        description,
        app_emoji,
        created_at,
        updated_at,
        request_timeout,
        first_date_surpassed,
        -- legacy field
        plan_type,
        -- legacy field
        daily_limit,
        -- legacy field
        custom_limit
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12
    )
RETURNING *;
-- name: InsertPortalApplicationAAT :one
INSERT INTO portal_application_aats (
        id,
        portal_application_id,
        address,
        public_key,
        private_key,
        client_public_key,
        signature,
        version
    )
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;
-- name: InsertPortalApplicationSetting :one
INSERT INTO portal_application_settings (
        application_id,
        secret_key,
        secret_key_required,
        monthly_relay_limit,
        environment
    )
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
-- name: GetUserFields :one
SELECT email,
    signed_up,
    icon_url,
    updates_product,
    updates_marketing,
    beta_tester
FROM users
WHERE users.id = $1;
-- name: GetAccountOwner :one
SELECT aua.account_id,
    aua.owner,
    aua.user_id,
    users.email,
    users.icon_url,
    users.updates_product,
    users.updates_marketing,
    users.beta_tester
FROM account_user_access AS aua
    JOIN users ON aua.user_id = users.id
WHERE aua.account_id = $1
    AND aua.owner = true;
-- name: UpdatePortalAppFields :exec
UPDATE portal_applications
SET name = COALESCE(NULLIF(@name::VARCHAR, ''), name),
    description = COALESCE(NULLIF(@description::VARCHAR, ''), description),
    app_emoji = COALESCE(NULLIF(@app_emoji::VARCHAR, ''), app_emoji),
    plan_type = COALESCE(NULLIF(@plan_type::VARCHAR, ''), plan_type),
    daily_limit = COALESCE($2, daily_limit),
    custom_limit = COALESCE($3, custom_limit),
    stripe_subscription_id = COALESCE($4, stripe_subscription_id),
    updated_at = $5
WHERE id = $1;
-- name: GetPlanDailyLimit :one
SELECT daily_limit
FROM pay_plans
WHERE plan_type = $1;
-- name: UpdatePortalAppSettings :exec
UPDATE portal_application_settings
SET secret_key = COALESCE($2, secret_key),
    secret_key_required = COALESCE($3, secret_key_required),
    monthly_relay_limit = COALESCE($4, monthly_relay_limit),
    environment = CASE
        WHEN @environment::environment IS NOT NULL THEN COALESCE(@environment::environment, environment)
        ELSE environment
    END::environment,
    favorited_chain_ids = COALESCE($5, favorited_chain_ids),
    updated_at = COALESCE($6, updated_at)
WHERE application_id = $1;
-- name: UpdateUpsertPortalAppNotification :exec
INSERT INTO portal_application_notifications (
        application_id,
        type,
        active,
        destination,
        trigger,
        events,
        updated_at
    )
SELECT $1,
    @type::notification_type,
    @active::BOOLEAN,
    @destination::VARCHAR(255),
    @trigger::VARCHAR(255),
    @events::notification_event [],
    $2
WHERE @active IS true ON CONFLICT (application_id, type) DO
UPDATE
SET active = EXCLUDED.active,
    destination = EXCLUDED.destination,
    trigger = EXCLUDED.trigger,
    events = EXCLUDED.events,
    updated_at = EXCLUDED.updated_at
WHERE EXCLUDED.active IS true;
-- name: UpdateDeletePortalAppNotification :exec
DELETE FROM portal_application_notifications
WHERE application_id = $1
    and type = $2;
-- name: CheckIfAppWhitelistExists :one
SELECT EXISTS(
        SELECT 1
        FROM portal_application_whitelists
        WHERE application_id = $1
            AND type = $2
            AND value = $3
            AND chain_id IS NULL
    );
-- name: UpdateInsertWhitelists :exec
INSERT INTO portal_application_whitelists (
        application_id,
        type,
        chain_id,
        value,
        created_at
    )
VALUES(
        $1,
        unnest(@types::whitelist_type []),
        NULLIF(unnest(@chain_ids::VARCHAR []), ''),
        unnest(@values::VARCHAR []),
        @created_at::TIMESTAMPTZ
    ) ON CONFLICT (application_id, chain_id, type, value) DO NOTHING;
-- name: UpdateDeleteWhitelists :exec
DELETE FROM portal_application_whitelists
WHERE (
        type IN ('methods', 'contracts')
        AND application_id = $1
        AND NOT EXISTS (
            SELECT 1
            FROM unnest(@types::whitelist_type []) t1(type),
                unnest(@values::VARCHAR []) t2(value),
                unnest(@chain_ids::VARCHAR []) t3(chain_id)
            WHERE t1.type = portal_application_whitelists.type
                AND t2.value = portal_application_whitelists.value
                AND t3.chain_id = portal_application_whitelists.chain_id
        )
    )
    OR (
        type IN ('blockchains', 'origins', 'userAgents')
        AND application_id = $1
        AND NOT EXISTS (
            SELECT 1
            FROM unnest(@types::whitelist_type []) t1(type),
                unnest(@values::VARCHAR []) t2(value)
            WHERE t1.type = portal_application_whitelists.type
                AND t2.value = portal_application_whitelists.value
        )
    );
-- name: UpdateFirstDatesSurpassed :exec
UPDATE portal_applications
SET first_date_surpassed = @first_date_surpassed
WHERE id = ANY (@portal_app_ids::VARCHAR []);
-- name: DeletePortalApp :exec
UPDATE portal_applications
SET deleted = true,
    deleted_at = $2
WHERE id = $1;
-- name: SelectAccounts :many
WITH filtered_portal_applications AS (
    SELECT *
    FROM portal_applications
    WHERE deleted = false
),
app_role_agg_owner AS (
    SELECT aua.user_id,
        aua.account_id,
        jsonb_object_agg(fpa.id, 'OWNER') AS portal_application_roles,
        jsonb_object_agg(fpa.id, aua.accepted) AS portal_applications_accepted
    FROM account_user_access AS aua
        JOIN filtered_portal_applications AS fpa ON aua.account_id = fpa.account_id
    WHERE aua.owner = true
    GROUP BY aua.user_id,
        aua.account_id
),
app_role_agg_non_owner AS (
    SELECT aua.user_id,
        aua.account_id,
        jsonb_object_agg(aua.portal_application_id, aua.role_name) AS portal_application_roles,
        jsonb_object_agg(aua.portal_application_id, aua.accepted) AS portal_applications_accepted
    FROM account_user_access AS aua
        JOIN filtered_portal_applications AS fpa ON aua.portal_application_id = fpa.id
    WHERE aua.owner = false
    GROUP BY aua.user_id,
        aua.account_id
),
app_role_agg AS (
    SELECT COALESCE(aro.user_id, aroo.user_id) as user_id,
        COALESCE(aro.account_id, aroo.account_id) as account_id,
        COALESCE(aro.portal_application_roles, '{}'::jsonb) || COALESCE(aroo.portal_application_roles, '{}'::jsonb) as portal_application_roles,
        COALESCE(aro.portal_applications_accepted, '{}'::jsonb) || COALESCE(aroo.portal_applications_accepted, '{}'::jsonb) as portal_applications_accepted
    FROM app_role_agg_owner aro
        FULL JOIN app_role_agg_non_owner aroo ON aro.user_id = aroo.user_id
        AND aro.account_id = aroo.account_id
)
SELECT a.*,
    ai.covalent_api_key_free,
    ai.covalent_api_key_paid,
    json_agg(
        json_build_object(
            'user_id',
            u.id,
            'email',
            u.email,
            'owner',
            aua.owner,
            'portal_application_roles',
            ara.portal_application_roles,
            'portal_applications_accepted',
            ara.portal_applications_accepted
        )
    ) AS users
FROM accounts AS a
    LEFT JOIN account_user_access AS aua ON a.id = aua.account_id
    LEFT JOIN portal_applications AS pa ON aua.portal_application_id = pa.id
    LEFT JOIN account_integrations AS ai ON a.id = ai.account_id
    LEFT JOIN users AS u ON aua.user_id = u.id
    LEFT JOIN user_roles AS ur ON aua.role_name = ur.role_name
    LEFT JOIN app_role_agg AS ara ON u.id = ara.user_id
    AND a.id = ara.account_id
WHERE (
        $1::BOOLEAN
        OR a.deleted = false
    )
GROUP BY a.id,
    ai.covalent_api_key_free,
    ai.covalent_api_key_paid;
-- name: SelectAccount :one
WITH app_role_agg_owner AS (
    SELECT aua.user_id,
        aua.account_id,
        jsonb_object_agg(pa.id, 'OWNER') AS portal_application_roles
    FROM account_user_access AS aua
        JOIN portal_applications AS pa ON aua.account_id = pa.account_id
    WHERE aua.owner = true
    GROUP BY aua.user_id,
        aua.account_id
),
app_role_agg_non_owner AS (
    SELECT aua.user_id,
        aua.account_id,
        jsonb_object_agg(aua.portal_application_id, aua.role_name) AS portal_application_roles
    FROM account_user_access AS aua
        JOIN portal_applications AS pa ON aua.portal_application_id = pa.id
    WHERE aua.owner = false
    GROUP BY aua.user_id,
        aua.account_id
),
app_role_agg AS (
    SELECT COALESCE(aro.user_id, aroo.user_id) as user_id,
        COALESCE(aro.account_id, aroo.account_id) as account_id,
        COALESCE(aro.portal_application_roles, '{}'::jsonb) || COALESCE(aroo.portal_application_roles, '{}'::jsonb) as portal_application_roles
    FROM app_role_agg_owner aro
        FULL JOIN app_role_agg_non_owner aroo ON aro.user_id = aroo.user_id
        AND aro.account_id = aroo.account_id
)
SELECT a.*,
    ai.covalent_api_key_free,
    ai.covalent_api_key_paid,
    json_agg(
        json_build_object(
            'user_id',
            u.id,
            'email',
            u.email,
            'accepted',
            aua.accepted,
            'owner',
            aua.owner,
            'portal_application_roles',
            ara.portal_application_roles
        )
    ) AS users
FROM accounts AS a
    LEFT JOIN account_user_access AS aua ON a.id = aua.account_id
    LEFT JOIN portal_applications AS pa ON aua.portal_application_id = pa.id
    LEFT JOIN account_integrations AS ai ON a.id = ai.account_id
    LEFT JOIN users AS u ON aua.user_id = u.id
    LEFT JOIN user_roles AS ur ON aua.role_name = ur.role_name
    LEFT JOIN app_role_agg AS ara ON u.id = ara.user_id
    AND a.id = ara.account_id
WHERE a.id = $1
GROUP BY a.id,
    ai.covalent_api_key_free,
    ai.covalent_api_key_paid;
-- name: UpdateAccountFields :exec
UPDATE accounts
SET name = COALESCE($1, name),
    icon_url = COALESCE($2, icon_url),
    plan_type = COALESCE($3, plan_type),
    partner_chain_ids = COALESCE($4, partner_chain_ids),
    partner_throughput_limit = COALESCE($5, partner_throughput_limit),
    partner_application_limit = COALESCE($6, partner_application_limit),
    updated_at = $7
WHERE id = $8;
-- name: UpsertAccountIntegrations :one
INSERT INTO account_integrations (
        account_id,
        covalent_api_key_free,
        covalent_api_key_paid,
        created_at,
        updated_at
    )
VALUES ($1, $2, $3, $4, $5) ON CONFLICT (account_id) DO
UPDATE
SET covalent_api_key_free = CASE
        WHEN EXCLUDED.covalent_api_key_free IS NOT NULL THEN EXCLUDED.covalent_api_key_free
        ELSE account_integrations.covalent_api_key_free
    END,
    covalent_api_key_paid = CASE
        WHEN EXCLUDED.covalent_api_key_paid IS NOT NULL THEN EXCLUDED.covalent_api_key_paid
        ELSE account_integrations.covalent_api_key_paid
    END,
    created_at = CASE
        WHEN EXCLUDED.created_at IS NOT NULL THEN EXCLUDED.created_at
        ELSE account_integrations.created_at
    END,
    updated_at = EXCLUDED.updated_at
RETURNING *;
-- name: SelectUserPermissions :many
WITH filtered_portal_applications AS (
    SELECT *
    FROM portal_applications
    WHERE deleted = false
)
SELECT aua.user_id,
    aua.account_id,
    aua.role_name,
    aua.owner,
    ur.permissions::permissions [] AS permissions,
    CASE
        WHEN aua.owner THEN array_agg(fpa.id)::VARCHAR []
        ELSE ARRAY [aua.portal_application_id]::VARCHAR []
    END AS portal_application_ids
FROM account_user_access AS aua
    LEFT JOIN user_roles AS ur ON aua.role_name = ur.role_name
    LEFT JOIN filtered_portal_applications AS fpa ON aua.account_id = fpa.account_id
WHERE aua.accepted = true
    AND aua.user_id IS NOT NULL
    AND fpa.id IS NOT NULL
GROUP BY aua.user_id,
    aua.account_id,
    aua.role_name,
    aua.owner,
    ur.permissions,
    aua.portal_application_id;
-- name: SelectUserIDs :many
SELECT user_id,
    provider_user_id
FROM user_auth_providers;
-- name: InsertAccount :one
INSERT INTO accounts (
        id,
        plan_type,
        name,
        icon_url,
        created_at,
        updated_at
    )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
-- name: DeleteAccount :exec
UPDATE accounts
SET deleted = true,
    deleted_at = $2
WHERE id = $1;
-- name: CheckChainExists :one
SELECT EXISTS(
        SELECT 1
        FROM chains
        WHERE id = $1
            AND deleted = false
    );
-- name: CheckAliasExists :one
SELECT EXISTS(
        SELECT 1
        FROM chain_alias_domains
        WHERE chain_id = $1
            AND alias = $2
    );
-- name: CheckPlanTypeExists :one
SELECT EXISTS(
        SELECT 1
        FROM pay_plans
        WHERE plan_type = $1
    );
-- name: CheckUserEmail :one
SELECT email
FROM users
WHERE id = $1;
-- name: CheckUserIDFromEmail :one
SELECT id
FROM users
WHERE email = $1;
-- name: CheckAccountExists :one
SELECT EXISTS(
        SELECT 1
        FROM accounts
        WHERE id = $1
            AND deleted = false
    );
-- name: CheckPlanExists :one
SELECT EXISTS(
        SELECT 1
        FROM pay_plans
        WHERE plan_type = $1
    );
-- name: CheckPortalAppExists :one
SELECT EXISTS(
        SELECT 1
        FROM portal_applications
        WHERE id = $1
            AND deleted = false
    );
-- name: CheckUserExists :one
SELECT EXISTS(
        SELECT 1
        FROM users
        WHERE id = $1
    );
-- name: CheckUserAccountExists :one
SELECT EXISTS (
        SELECT 1
        FROM account_user_access
        WHERE user_id = $1
    );
-- name: CheckAccountUserExists :one
SELECT EXISTS (
        SELECT 1
        FROM account_user_access
        WHERE (
                user_id = $1
                AND portal_application_id = $2
            )
            OR (
                user_id = $1
                AND role_name = 'OWNER'
                AND account_id = $3
            )
    );
-- name: CheckUserProviderExists :one
SELECT EXISTS (
        SELECT 1
        FROM users
            JOIN user_auth_providers ON users.id = user_auth_providers.user_id
        WHERE users.email = $1
            AND user_auth_providers.type = $2
    );
-- name: CheckAccountUserRole :one
SELECT role_name
FROM account_user_access
WHERE user_id = $1
    AND account_id = $2;
-- name: CheckAccountUserAcceptedOrOwner :one
SELECT accepted
FROM account_user_access
WHERE user_id = $1
    AND (
        portal_application_id = $2
        OR owner = true
    );
-- name: CheckAccountUserIsOwner :one
SELECT owner
FROM account_user_access
WHERE user_id = $1
    AND account_id = $2;
-- name: CheckAccountUserAccepted :one
SELECT accepted
FROM account_user_access
WHERE user_id = $1
    AND portal_application_id = $2;
-- name: InsertAccountUserAccess :one
WITH updated_user AS (
    UPDATE users
    SET email = $7
    WHERE id = $1
    RETURNING id,
        email
) -- Must update email to trigger listener notification for users table
INSERT INTO account_user_access (
        user_id,
        account_id,
        portal_application_id,
        role_name,
        owner,
        accepted,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        NULLIF(@portal_application_id::VARCHAR, ''),
        $3,
        $4,
        $5,
        $6,
        $8
    )
RETURNING user_id;
-- name: InsertAccountUserAccessNoUser :one
WITH inserted_user AS (
    INSERT INTO users (
            id,
            email,
            signed_up,
            created_at,
            updated_at
        )
    VALUES ($1, $2, false, $3, $4)
    RETURNING id,
        email
)
INSERT INTO account_user_access (
        user_id,
        account_id,
        portal_application_id,
        role_name,
        owner,
        accepted,
        created_at,
        updated_at
    )
VALUES (
        (
            SELECT id
            FROM inserted_user
        ),
        $6,
        $7,
        $5,
        false,
        false,
        $3,
        $4
    )
RETURNING user_id;
-- name: UpdateAccountUserRole :exec
UPDATE account_user_access
SET role_name = $3,
    updated_at = $4
WHERE portal_application_id = $1
    AND user_id = $2;
-- name: TransferOwnerDeleteOldRows :one
WITH delete_old_owner_row AS (
    DELETE FROM account_user_access
    WHERE account_user_access.account_id = $1
        AND role_name = 'OWNER'
    RETURNING user_id
),
delete_new_owner_non_owner_rows AS (
    DELETE FROM account_user_access
    WHERE account_user_access.account_id = $1
        AND user_id = @new_owner_id::VARCHAR(24)
        AND role_name <> 'OWNER'
)
SELECT user_id
FROM delete_old_owner_row;
-- name: TransferOwnerCreateRows :exec
WITH updated_user AS (
    UPDATE users
    SET email = users.email
    WHERE id = @new_owner_id::VARCHAR(24)
),
insert_old_owner_admin_rows AS (
    INSERT INTO account_user_access (
            portal_application_id,
            user_id,
            account_id,
            role_name,
            owner,
            accepted,
            created_at,
            updated_at
        )
    SELECT DISTINCT on (pa.id) pa.id,
        @old_owner_id::VARCHAR(24),
        aua.account_id,
        'ADMIN',
        false,
        true,
        aua.created_at,
        aua.updated_at
    FROM account_user_access AS aua
        LEFT JOIN portal_applications AS pa ON aua.account_id = pa.account_id
    WHERE aua.account_id = $1
        AND pa.deleted = false -- Only include non-deleted portal applications
        AND NOT EXISTS (
            SELECT 1
            FROM account_user_access
            WHERE user_id = @old_owner_id::VARCHAR(24)
                AND portal_application_id = pa.id
        )
),
insert_new_owner_row AS (
    INSERT INTO account_user_access (
            user_id,
            account_id,
            role_name,
            owner,
            accepted,
            created_at,
            updated_at
        )
    SELECT @new_owner_id::VARCHAR(24),
        $1,
        'OWNER',
        true,
        true,
        $2,
        $3
)
SELECT 1;
-- name: CreateUserNewSignUp :one
WITH inserted_user AS (
    INSERT INTO users (id, email, signed_up, created_at, updated_at)
    VALUES ($1, $2, true, $3, $4) ON CONFLICT (email) DO NOTHING
    RETURNING id
)
INSERT INTO user_auth_providers (
        user_id,
        type,
        provider,
        provider_user_id,
        federated
    )
VALUES (
        (
            SELECT COALESCE(
                    (
                        SELECT id
                        FROM inserted_user
                    ),
                    (
                        SELECT id
                        FROM users
                        WHERE users.email = $2
                    )
                )
        ),
        $5,
        $6,
        $7,
        $8
    )
RETURNING user_id;
-- name: UpdateUserFields :exec
UPDATE users
SET icon_url = COALESCE($1, icon_url),
    updates_product = COALESCE($2, updates_product),
    updates_marketing = COALESCE($3, updates_marketing),
    beta_tester = COALESCE($4, beta_tester),
    updated_at = $5
WHERE id = $6;
-- name: UpdateUserAcceptedInvite :exec
WITH inserted_or_existing_provider AS (
    INSERT INTO user_auth_providers (
            user_id,
            type,
            provider,
            provider_user_id,
            federated
        )
    VALUES ($1, $2, $3, $4, $5) ON CONFLICT (user_id, type) DO
    UPDATE
    SET user_id = EXCLUDED.user_id
    RETURNING user_id
),
updated_access AS (
    UPDATE account_user_access
    SET accepted = true
    WHERE user_id = (
            SELECT user_id
            FROM inserted_or_existing_provider
        )
        AND portal_application_id = $6
)
UPDATE users
SET signed_up = true
WHERE id = (
        SELECT user_id
        FROM inserted_or_existing_provider
    );
-- name: UpdateUserDeclinedInvite :exec
WITH user_check AS (
    SELECT user_id
    FROM user_auth_providers uap
    WHERE uap.user_id = $1
    UNION ALL
    SELECT user_id
    FROM account_user_access aua
    WHERE aua.user_id = $1
        AND aua.portal_application_id != $2
),
delete_aua AS (
    DELETE FROM account_user_access aua
    WHERE aua.user_id = $1
        AND aua.portal_application_id = $2
)
DELETE FROM users u
WHERE u.id = $1
    AND NOT EXISTS (
        SELECT user_id
        FROM user_check
    );
-- name: DeleteAccountUser :exec
DELETE FROM account_user_access
WHERE portal_application_id = $1
    AND user_id = $2;
-- name: GetPortalUserID :one
SELECT user_id
FROM user_auth_providers
WHERE provider_user_id = $1;
-- name: SelectAllUsers :many
SELECT users.*,
    json_agg(user_auth_providers.*) AS auth_providers
FROM users
    LEFT JOIN user_auth_providers ON users.id = user_auth_providers.user_id
GROUP BY users.id;
-- name: GetUserDataFromPortalUserID :one
SELECT users.*,
    json_agg(user_auth_providers.*) AS auth_providers
FROM users
    LEFT JOIN user_auth_providers ON users.id = user_auth_providers.user_id
WHERE users.id = $1
GROUP BY users.id;
-- name: GetUserDataFromAuthProviderUserID :one
SELECT users.*,
    json_agg(user_auth_providers.*) AS auth_providers
FROM users
    LEFT JOIN user_auth_providers ON users.id = user_auth_providers.user_id
WHERE user_auth_providers.provider_user_id = $1
GROUP BY users.id;
-- name: DeleteUser :one
DELETE FROM users
WHERE id = $1
RETURNING id;
-- name: SelectChains :many
SELECT c.*,
    COALESCE(
        json_agg(DISTINCT ca) FILTER (
            WHERE ca.id IS NOT NULL
        ),
        '[]'
    )::json AS chain_altruists,
    COALESCE(
        json_agg(DISTINCT cc) FILTER (
            WHERE cc.id IS NOT NULL
        ),
        '[]'
    )::json AS chain_checks,
    -- DEPRECATED - TODO remove when move to only store aliases is complete
    COALESCE(
        json_object_agg(COALESCE(cga.alias, 'null'), cga.domains) FILTER (
            WHERE cga.alias IS NOT NULL
        ),
        '{}'
    )::json AS alias_domains_map,
    ARRAY(
        SELECT DISTINCT alias
        FROM chain_aliases
        WHERE chain_id = c.id
    )::VARCHAR [] AS chain_aliases
FROM chains c
    LEFT JOIN chain_altruists ca ON c.id = ca.chain_id
    LEFT JOIN chain_checks cc ON c.id = cc.chain_id
    LEFT JOIN chain_alias_domains cga ON c.id = cga.chain_id
WHERE (
        @include_deleted::BOOLEAN
        OR c.deleted = false
    )
GROUP BY c.id;
-- name: SelectChain :one
SELECT c.*,
    COALESCE(
        json_agg(DISTINCT ca) FILTER (
            WHERE ca.id IS NOT NULL
        ),
        '[]'
    )::json AS chain_altruists,
    COALESCE(
        json_agg(DISTINCT cc) FILTER (
            WHERE cc.id IS NOT NULL
        ),
        '[]'
    )::json AS chain_checks,
    -- DEPRECATED - TODO remove when move to only store aliases is complete
    COALESCE(
        json_object_agg(COALESCE(cga.alias, 'null'), cga.domains) FILTER (
            WHERE cga.alias IS NOT NULL
        ),
        '{}'
    )::json AS alias_domains_map,
    ARRAY(
        SELECT DISTINCT alias
        FROM chain_aliases
        WHERE chain_id = c.id
    )::VARCHAR [] AS chain_aliases
FROM chains c
    LEFT JOIN chain_altruists ca ON c.id = ca.chain_id
    LEFT JOIN chain_checks cc ON c.id = cc.chain_id
    LEFT JOIN chain_alias_domains cga ON c.id = cga.chain_id
WHERE c.id = $1
GROUP BY c.id;
-- name: InsertChain :one
INSERT INTO chains (
        id,
        icon_url,
        blockchain,
        description,
        enforce_result,
        path,
        ticker,
        request_timeout,
        log_limit_blocks,
        allowed_methods,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12
    )
RETURNING id;
-- name: UpdateChain :one
UPDATE chains
SET blockchain = COALESCE($2, chains.blockchain),
    icon_url = COALESCE($3, chains.icon_url),
    description = COALESCE($4, chains.description),
    enforce_result = COALESCE($5, chains.enforce_result),
    path = COALESCE($6, chains.path),
    ticker = COALESCE($7, chains.ticker),
    request_timeout = COALESCE($8, chains.request_timeout),
    log_limit_blocks = COALESCE($9, chains.log_limit_blocks),
    allowed_methods = COALESCE($10, chains.allowed_methods),
    updated_at = $11
WHERE id = $1
RETURNING id;
-- name: UpsertChainAltruist :exec
INSERT INTO chain_altruists (
        chain_id,
        url,
        auth,
        auth_type,
        created_at,
        updated_at
    )
VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (chain_id, url) DO
UPDATE
SET auth = COALESCE(EXCLUDED.auth, chain_altruists.auth),
    auth_type = COALESCE(EXCLUDED.auth_type, chain_altruists.auth_type),
    updated_at = EXCLUDED.updated_at;
-- name: DeleteUnusedChainAltruists :exec
DELETE FROM chain_altruists
WHERE chain_id = $1
    AND url NOT IN (
        SELECT unnest(@urls::VARCHAR [])
    );
-- name: InsertChainAlias :exec
INSERT INTO chain_aliases (
        chain_id,
        alias,
        created_at
    )
VALUES ($1, $2, $3) ON CONFLICT (chain_id, alias) DO NOTHING;
-- name: DeleteUnusedChainAlias :exec
DELETE FROM chain_aliases
WHERE chain_id = $1
    AND alias NOT IN (
        SELECT unnest(@aliases::VARCHAR [])
    );
-- name: DeleteChainAlias :exec
DELETE FROM chain_aliases
WHERE chain_id = $1
    AND alias = $2;
-- name: UpsertChainAliasDomains :exec
INSERT INTO chain_alias_domains (
        chain_id,
        alias,
        domains,
        updated_at
    )
VALUES ($1, $2, $3, $4) ON CONFLICT (chain_id, alias) DO
UPDATE
SET domains = COALESCE(
        EXCLUDED.domains,
        chain_alias_domains.domains
    ),
    updated_at = EXCLUDED.updated_at;
-- name: DeleteUnusedChainAliasDomains :exec
DELETE FROM chain_alias_domains
WHERE chain_id = $1
    AND alias NOT IN (
        SELECT unnest(@aliases::VARCHAR [])
    );
-- name: DeleteChainAliasDomain :exec
DELETE FROM chain_alias_domains
WHERE chain_id = $1
    AND alias = $2;
-- name: UpsertChainCheck :exec
INSERT INTO chain_checks (
        chain_id,
        type,
        payload,
        result_key,
        allowance,
        evm_chain_id,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8
    ) ON CONFLICT (chain_id, type) DO
UPDATE
SET payload = COALESCE(EXCLUDED.payload, chain_checks.payload),
    result_key = COALESCE(EXCLUDED.result_key, chain_checks.result_key),
    allowance = COALESCE(EXCLUDED.allowance, chain_checks.allowance),
    evm_chain_id = COALESCE(EXCLUDED.evm_chain_id, chain_checks.evm_chain_id),
    updated_at = EXCLUDED.updated_at;
-- name: DeleteUnusedChainChecks :exec
DELETE FROM chain_checks
WHERE chain_id = $1
    AND type NOT IN (
        SELECT unnest(@types::chain_check_type [])
    );
-- name: UpdateChainActive :one
UPDATE chains
SET active = $2,
    updated_at = $3
WHERE id = $1
RETURNING active;
-- name: InsertGigastakeApp :exec
WITH new_gigastake_application AS (
    INSERT INTO gigastake_applications (
            id,
            name,
            address,
            public_key,
            client_public_key,
            signature,
            version,
            created_at,
            updated_at
        )
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    RETURNING id
)
INSERT INTO chains_gigastake_applications (chain_id, gigastake_application_id)
SELECT unnest(@chain_ids::VARCHAR []),
    (
        SELECT id
        FROM new_gigastake_application
    ) ON CONFLICT (chain_id, gigastake_application_id) DO NOTHING;
-- name: UpdateGigastakeAppNameAndChainIDs :exec
WITH updated_gigastake_application AS (
    UPDATE gigastake_applications
    SET name = $2,
        updated_at = $3
    WHERE id = $1
    RETURNING id
),
inserted AS (
    INSERT INTO chains_gigastake_applications (chain_id, gigastake_application_id)
    SELECT unnest(@chain_ids::VARCHAR []),
        updated_gigastake_application.id
    FROM updated_gigastake_application ON CONFLICT (chain_id, gigastake_application_id) DO NOTHING
)
DELETE FROM chains_gigastake_applications
WHERE gigastake_application_id IN (
        SELECT id
        FROM updated_gigastake_application
    )
    AND chain_id NOT IN (
        SELECT unnest(@chain_ids::VARCHAR [])
    );
-- name: SelectGlobalBlockedContract :many
SELECT id,
    blocked_address
FROM global_blocked_contracts
WHERE active = true;
-- name: AddGlobalBlockedContract :exec
INSERT INTO global_blocked_contracts (blocked_address, created_at, updated_at)
VALUES ($1, $2, $3);
-- name: SetGlobalBlockedContractActive :one
UPDATE global_blocked_contracts
SET active = $2,
    updated_at = $3
WHERE blocked_address = $1
RETURNING id;
-- name: RemoveGlobalBlockedContract :one
DELETE FROM global_blocked_contracts
WHERE blocked_address = $1
RETURNING id;
-- name: GetLastCreatedUserID :one
SELECT id
FROM users
ORDER BY created_at DESC
LIMIT 1;
