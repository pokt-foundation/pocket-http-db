-- Insert pay_plans
INSERT INTO pay_plans (
        name,
        description,
        plan_type,
        chain_ids,
        monthly_relay_limit,
        throughput_limit,
        application_limit,
        created_at,
        updated_at,
        daily_limit
    )
VALUES (
        'Free Tier',
        'Ideal for small projects and testing',
        'FREETIER_V0',
        ARRAY ['0001', '0053'],
        5000000,
        5000,
        5,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        250000
    ),
    (
        'Pay As You Go',
        'Pay only for what you use',
        'PAY_AS_YOU_GO_V0',
        ARRAY ['0001', '0053'],
        10000000,
        10000,
        5,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        0
    ),
    (
        'Enterprise',
        'Premium plan for large businesses',
        'ENTERPRISE',
        ARRAY ['0001', '0053'],
        10000000,
        10000,
        5,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        0
    ),
    (
        'Test Plan',
        'For testing purposes',
        'TEST_PLAN_V0',
        ARRAY ['0001', '0053'],
        10000000,
        10000,
        5,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        0
    ),
    (
        'Test Plan 10K',
        'Test plan with 10K daily limit for testing purposes',
        'TEST_PLAN_10K',
        ARRAY ['0001', '0053'],
        10000000,
        10000,
        5,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        10000
    ),
    (
        'Test Plan 90K',
        'Test plan with 90K daily limit for testing purposes',
        'TEST_PLAN_90K',
        ARRAY ['0001', '0053'],
        90000000,
        10000,
        5,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        90000
    ),
    (
        'Basic Plan',
        'Basic plan for testing purposes',
        'basic_plan',
        ARRAY ['0001', '0053'],
        5000000,
        5000,
        2,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        1000
    ),
    (
        'Pro Plan',
        'Pro plan for testing purposes',
        'pro_plan',
        ARRAY ['0001', '0053', '0021', '0064'],
        10000000,
        10000,
        5,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        5000
    ),
    (
        'Enterprise Plan',
        'Enterprise plan for testing purposes',
        'enterprise_plan',
        ARRAY ['0001', '0053', '0021', '0064', '0034'],
        20000000,
        20000,
        10,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        10000
    ),
    (
        'Developer Plan',
        'Developer plan for testing purposes',
        'developer_plan',
        ARRAY ['0001', '0053', '0021', '0034'],
        500000,
        500,
        1,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        100
    ),
    (
        'Startup Plan',
        'Startup plan for testing purposes',
        'startup_plan',
        ARRAY ['0001', '0053', '0064', '0034'],
        1000000,
        1000,
        5,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        500
    );
-- Insert users
INSERT INTO users (
        id,
        email,
        signed_up,
        created_at,
        updated_at,
        icon_url,
        updates_product,
        updates_marketing,
        beta_tester
    )
VALUES (
        'user_1',
        'james.holden123@test.com',
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        'https://picsum.photos/200',
        true,
        false,
        true
    ),
    (
        'user_2',
        'paul.atreides456@test.com',
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        'https://picsum.photos/200',
        false,
        true,
        false
    ),
    (
        'user_3',
        'ellen.ripley789@test.com',
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        'https://picsum.photos/200',
        true,
        true,
        false
    ),
    (
        'user_4',
        'ulfric.stormcloak123@test.com',
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        'https://picsum.photos/200',
        false,
        false,
        true
    ),
    (
        'user_5',
        'chrisjen.avasarala1@test.com',
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        'https://picsum.photos/200',
        true,
        false,
        false
    ),
    (
        'user_6',
        'amos.burton789@test.com',
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        'https://picsum.photos/200',
        false,
        true,
        true
    ),
    (
        'user_7',
        'frodo.baggins123@test.com',
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        'https://picsum.photos/200',
        true,
        true,
        true
    ),
    (
        'user_8',
        'rick.deckard456@test.com',
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        'https://picsum.photos/200',
        true,
        true,
        false
    ),
    (
        'user_9',
        'tyrion.lannister789@test.com',
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        'https://picsum.photos/200',
        true,
        true,
        false
    ),
    (
        'user_10',
        'daenerys.targaryen123@test.com',
        false,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        'https://picsum.photos/200',
        false,
        false,
        true
    ),
    (
        'user_11',
        'bernard.marx@test.com',
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        'https://picsum.photos/200',
        true,
        true,
        false
    ),
    (
        'user_12',
        'george.foreman@test.com',
        false,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000',
        'https://picsum.photos/200',
        false,
        true,
        false
    );
INSERT INTO user_auth_providers (
        user_id,
        provider,
        type,
        provider_user_id,
        federated,
        created_at
    )
VALUES (
        'user_1',
        'auth0',
        'auth0_username',
        'auth0|james_holden',
        false,
        '2022-11-11 11:11:11.000000'
    ),
    (
        'user_1',
        'auth0',
        'auth0_github',
        'github|james_holden',
        true,
        '2022-11-11 11:11:11.000000'
    ),
    (
        'user_2',
        'auth0',
        'auth0_username',
        'auth0|paul_atreides',
        false,
        '2022-11-11 11:11:11.000000'
    ),
    (
        'user_2',
        'auth0',
        'auth0_github',
        'github|paul_atreides',
        true,
        '2022-11-11 11:11:11.000000'
    ),
    (
        'user_3',
        'auth0',
        'auth0_username',
        'auth0|ellen_ripley',
        false,
        '2022-11-11 11:11:11.000000'
    ),
    (
        'user_4',
        'auth0',
        'auth0_username',
        'auth0|ulfric_stormcloak',
        false,
        '2022-11-11 11:11:11.000000'
    ),
    (
        'user_5',
        'auth0',
        'auth0_username',
        'auth0|chrisjen_avasarala',
        false,
        '2022-11-11 11:11:11.000000'
    ),
    (
        'user_6',
        'auth0',
        'auth0_username',
        'auth0|amos_burton',
        false,
        '2022-11-11 11:11:11.000000'
    ),
    (
        'user_7',
        'auth0',
        'auth0_username',
        'auth0|frodo_baggins',
        false,
        '2022-11-11 11:11:11.000000'
    ),
    (
        'user_8',
        'auth0',
        'auth0_username',
        'auth0|rick_deckard',
        false,
        '2022-11-11 11:11:11.000000'
    ),
    (
        'user_9',
        'auth0',
        'auth0_username',
        'auth0|tyrion_lannister',
        false,
        '2022-11-11 11:11:11.000000'
    ),
    (
        'user_11',
        'auth0',
        'auth0_username',
        'auth0|bernard_marx',
        false,
        '2022-11-11 11:11:11.000000'
    );
-- Insert accounts
INSERT INTO accounts (
        id,
        name,
        icon_url,
        plan_type,
        partner_chain_ids,
        partner_throughput_limit,
        partner_application_limit,
        created_at,
        updated_at
    )
VALUES (
        'account_1',
        'The Brave Voyager',
        'https://picsum.photos/200',
        'basic_plan',
        ARRAY ['0001', '0053'],
        2000,
        1,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_2',
        'Dragonborn Explorer',
        'https://picsum.photos/200',
        'pro_plan',
        ARRAY ['0001', '0053','0021', '0064'],
        5000,
        3,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_3',
        'Fellowship Startup',
        'https://picsum.photos/200',
        'startup_plan',
        ARRAY ['0001', '0053', '0064', '0034'],
        1000,
        2,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_4',
        'Iron Throne Enterprise',
        'https://picsum.photos/200',
        'enterprise_plan',
        ARRAY ['0001'],
        1000,
        2,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_5',
        'Nomad Wanderer',
        'https://picsum.photos/200',
        'basic_plan',
        ARRAY ['0006', '0040'],
        6000,
        1,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    );
INSERT INTO account_integrations (
        account_id,
        covalent_api_key_free,
        created_at,
        updated_at
    )
VALUES (
        'account_1',
        'covalent_api_key_1',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_2',
        'covalent_api_key_2',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_3',
        'covalent_api_key_3',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_4',
        'covalent_api_key_4',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    );
-- Insert user_roles
INSERT INTO user_roles (
        role_name,
        permissions,
        created_at,
        updated_at
    )
VALUES (
        'OWNER',
        ARRAY ['read:endpoint'::permissions, 'write:endpoint'::permissions, 'delete:endpoint'::permissions, 'transfer:endpoint'::permissions],
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'ADMIN',
        ARRAY ['read:endpoint'::permissions, 'write:endpoint'::permissions],
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'MEMBER',
        ARRAY ['read:endpoint'::permissions],
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    );
INSERT INTO portal_applications (
        id,
        account_id,
        name,
        app_emoji,
        description,
        request_timeout,
        first_date_surpassed,
        plan_type,
        daily_limit,
        custom_limit,
        stripe_subscription_id,
        created_at,
        updated_at
    )
VALUES (
        'test_app_1',
        'account_1',
        'pokt_app_123',
        '1F336',
        'Embark on an interstellar journey with our powerful application.',
        5000,
        '2022-11-11 11:11:11.000000',
        'FREETIER_V0',
        250000,
        NULL,
        'stripe_id_1',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'test_app_2',
        'account_2',
        'pokt_app_456',
        '1F336',
        'Unveil the mysteries of the multiverse with our enchanting application.',
        10000,
        '2022-11-11 11:11:11.000000',
        'PAY_AS_YOU_GO_V0',
        0,
        NULL,
        'stripe_id_2',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'test_app_3',
        'account_3',
        'pokt_app_789',
        '1F336',
        'Harness the power of a connected world with our revolutionary application.',
        10000,
        '2022-11-11 11:11:11.000000',
        'ENTERPRISE',
        NULL,
        '4200000',
        'stripe_id_3',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    );
-- Insert account_user_access
INSERT INTO account_user_access (
        account_id,
        user_id,
        portal_application_id,
        role_name,
        owner,
        accepted,
        created_at,
        updated_at
    )
VALUES (
        'account_1',
        'user_1',
        NULL,
        'OWNER',
        true,
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_1',
        'user_2',
        'test_app_1',
        'ADMIN',
        false,
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_1',
        'user_8',
        'test_app_1',
        'ADMIN',
        false,
        false,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_2',
        'user_3',
        NULL,
        'OWNER',
        true,
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_2',
        'user_4',
        'test_app_2',
        'MEMBER',
        false,
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_2',
        'user_9',
        'test_app_2',
        'MEMBER',
        false,
        false,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_2',
        'user_2',
        'test_app_2',
        'MEMBER',
        false,
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_3',
        'user_5',
        NULL,
        'OWNER',
        true,
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_3',
        'user_6',
        'test_app_3',
        'ADMIN',
        false,
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_3',
        'user_7',
        'test_app_3',
        'MEMBER',
        false,
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_3',
        'user_10',
        'test_app_3',
        'MEMBER',
        false,
        false,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_4',
        'user_4',
        NULL,
        'OWNER',
        true,
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'account_5',
        'user_4',
        NULL,
        'OWNER',
        true,
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    );
INSERT INTO portal_application_settings (
        application_id,
        secret_key,
        secret_key_required,
        monthly_relay_limit,
        environment,
        favorited_chain_ids,
        updated_at
    )
VALUES (
        'test_app_1',
        'test_40f482d91a5ef2300ebb4e2308c',
        true,
        2500000,
        'production',
        ARRAY ['0001', '0053'],
        '2022-11-11 11:11:11.000000'
    ),
    (
        'test_app_2',
        'test_9c9e3b193cfba5348f93bb2f3e3fb794',
        false,
        1500000,
        'production',
        ARRAY ['0021', '0064'],
        '2022-11-11 11:11:11.000000'
    ),
    (
        'test_app_3',
        'test_9f48b13e2bc5fd31ab367841f11495c1',
        false,
        4500000,
        'production',
        ARRAY ['0001', '0034'],
        '2022-11-11 11:11:11.000000'
    );
INSERT INTO portal_application_whitelists (
        application_id,
        type,
        value,
        chain_id,
        created_at
    )
VALUES (
        'test_app_1',
        'blockchains',
        '0053',
        NULL,
        '2022-11-11 11:11:11.000000'
    ),
    (
        'test_app_1',
        'contracts',
        '0x1234567890abcdef',
        '0001',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'test_app_1',
        'methods',
        'GET',
        '0001',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'test_app_1',
        'origins',
        'https://test.com',
        NULL,
        '2022-11-11 11:11:11.000000'
    ),
    (
        'test_app_1',
        'userAgents',
        'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        NULL,
        '2022-11-11 11:11:11.000000'
    ),
    (
        'test_app_2',
        'blockchains',
        '0021',
        NULL,
        '2022-11-11 11:11:11.000000'
    ),
    (
        'test_app_2',
        'contracts',
        '0x0987654321abcdef',
        '0064',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'test_app_2',
        'methods',
        'POST',
        '0064',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'test_app_2',
        'origins',
        'https://example.com',
        NULL,
        '2022-11-11 11:11:11.000000'
    ),
    (
        'test_app_2',
        'userAgents',
        'Mozilla/5.0 (Linux; Android 10; SM-A205U) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.120 Mobile Safari/537.36',
        NULL,
        '2022-11-11 11:11:11.000000'
    );
INSERT INTO portal_application_notifications (
        application_id,
        active,
        type,
        destination,
        trigger,
        events,
        updated_at
    )
VALUES (
        'test_app_1',
        true,
        'email',
        'test@test.com',
        'trigger123',
        ARRAY ['quarter'::notification_event,'threeQuarters'::notification_event,'full'::notification_event],
        '2022-11-11 11:11:11.000000'
    ),
    (
        'test_app_2',
        true,
        'email',
        'email@pokt.network',
        'trigger456',
        ARRAY ['full'::notification_event,'half'::notification_event],
        '2022-11-11 11:11:11.000000'
    );
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
VALUES (
        'test_protocol_app_1',
        'test_app_1',
        'test_34715cae753e67c75fbb340442e7de8e',
        'test_34715cae753e67c75fbb340442e7de8e',
        'test_11b8d394ca331d7c7a71ca1896d630f6',
        'test_89a3af6a587aec02cfade6f5000424c2',
        'test_1dc39a2e5a84a35bf030969a0b3231f7',
        '0.0.1'
    ),
    (
        'test_protocol_app_2',
        'test_app_2',
        'test_8237c72345f12d1b1a8b64a1a7f66fa4',
        'test_8237c72345f12d1b1a8b64a1a7f66fa4',
        'test_2e83c836a29b423a47d8e18c779fd422',
        'test_04c71d90a92f40416b6f1d7d8af17e02',
        'test_f48d33b30ddaf60a1e5bb50d2ba8da5a',
        '0.0.1'
    ),
    (
        'test_protocol_app_3',
        'test_app_3',
        'test_b5e07928fc80083c13ad0201b81bae9b',
        'test_f608500e4fe3e09014fe2411b4a560b5',
        'test_8663e187c19f3c6e27317eab4ed6d7d5',
        'test_328a9cf1b35085eeaa669aa858f6fba9',
        'test_c3cd8be16ba32e24dd49fdb0247fc9b8',
        '0.0.1'
    ),
    (
        'test_protocol_app_4',
        'test_app_3',
        'test_eb2e5bcba557cfe8fa76fd7fff54f9d1',
        'test_f6a5d8690ecb669865bd752b7796a920',
        'test_838d29d61a65401f7d56d084cb6e4783',
        'test_6ee5ea553408f0895923fd1569dc5072',
        'test_cf05cf9bb26111c548e88fb6157af708',
        '0.0.1'
    );
INSERT INTO chains (
        id,
        icon_url,
        blockchain,
        description,
        enforce_result,
        path,
        ticker,
        log_limit_blocks,
        request_timeout,
        active,
        created_at,
        updated_at
    )
VALUES (
        '0001',
        'https://picsum.photos/200',
        'pokt-mainnet',
        'Pocket Network Mainnet',
        'JSON',
        '/v1/query/height',
        'POKT',
        0,
        0,
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0053',
        'https://picsum.photos/200',
        'optimism-mainnet',
        'Optimism Mainnet',
        'JSON',
        '',
        'OP',
        100000,
        0,
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0021',
        'https://picsum.photos/200',
        'eth-mainnet',
        'Ethereum Mainnet',
        'JSON',
        '',
        'ETH',
        100000,
        0,
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0064',
        'https://picsum.photos/200',
        'sui-testnet',
        'Sui Testnet',
        'JSON',
        '',
        'SUI-TESTNET',
        100000,
        60000,
        false,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0040',
        'https://picsum.photos/200',
        'harmony-0',
        'Harmony Shard 0',
        'JSON',
        '',
        'HMY',
        0,
        0,
        true,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    );
INSERT INTO chain_altruists (
        chain_id,
        url,
        auth,
        auth_type,
        created_at,
        updated_at
    )
VALUES (
        '0001',
        'https://altruist-0001.com:1234',
        'test_pocket:auth123456',
        'basic_auth',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0053',
        'https://altruist-0053.com:1234',
        'test_pocket:auth123456',
        'basic_auth',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0021',
        'https://altruist-0021.com:1234',
        'test_pocket:auth123456',
        'basic_auth',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0064',
        'https://altruist-0064.com:1234',
        'test_pocket:auth123456',
        'basic_auth',
        '2022-11-11 11:11:11.000000',
        '2023-01-23T23:52:00.176019Z'
    ),
    (
        '0040',
        'https://altruist-0040.com:1234',
        'test_pocket:auth123456',
        'basic_auth',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    );
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
        '0001',
        'sync',
        '{"id":1,"jsonrpc":"2.0","method":"query"}',
        'result.sync_info',
        1,
        NULL,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0053',
        'sync',
        '{"id":1,"jsonrpc":"2.0","method":"eth_blockNumber","params":[]}',
        'result',
        2,
        NULL,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0021',
        'sync',
        '{"id":1,"jsonrpc":"2.0","method":"eth_blockNumber","params":[]}',
        'result',
        5,
        NULL,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0021',
        'chain',
        '{"method":"eth_chainId","id":1,"jsonrpc":"2.0"}',
        'id',
        NULL,
        1,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0064',
        'sync',
        '{"id":1,"jsonrpc":"2.0","method":"sui_blockNumber","params":[]}',
        'result',
        7,
        NULL,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0040',
        'sync',
        '{"id":1,"jsonrpc":"2.0","method":"hmy_blockNumber","params":[]}',
        'result',
        8,
        NULL,
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    );
-- DEPRECATED - TODO remove when move to only store aliases is complete
INSERT INTO chain_alias_domains (
        chain_id,
        alias,
        domains
    )
VALUES (
        '0001',
        'pokt-mainnet',
        ARRAY ['pokt-rpc.gateway.pokt.network']
    ),
    (
        '0053',
        'optimism-mainnet',
        ARRAY ['op-rpc.gateway.pokt.network']
    ),
    (
        '0021',
        'eth-mainnet',
        ARRAY ['eth-rpc.gateway.pokt.network']
    ),
    (
        '0040',
        'harmony-0',
        ARRAY ['hmy-rpc.gateway.pokt.network']
    );
INSERT INTO chain_aliases (
        chain_id,
        alias,
        created_at
    )
VALUES (
        '0001',
        'pokt-mainnet',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0053',
        'optimism-mainnet',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0053',
        'optimism-rpc',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0021',
        'eth-mainnet',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0021',
        'eth-rpc',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0064',
        'sui-testnet',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0040',
        'harmony-0',
        '2022-11-11 11:11:11.000000'
    );
-- Insert data into aats
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
VALUES (
        'test_gigastake_app_1',
        'pokt_gigastake',
        'test_8d4f6a5b0c6e9f1db12c1f662e5ec8c5',
        'test_37a0e8437f5149dc98a9a5b207efc2d0',
        'test_65c29f0cc82e418b81a528a0c0682a9f',
        'test_f22651fb566346fca30b605e5f46e3ca',
        '0.0.1',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'test_gigastake_app_2',
        'optimism_gigastake',
        'test_5c60d434db4e42d2b5d2ea6eeb8933c4',
        'test_a7e28f8d716541a0a332a5dc6b7e4e6e',
        'test_ba4e53dada8f4f939048e56dc8f88f37',
        'test_52e991c26da841bc882ad3a3ee9ee964',
        '0.0.1',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        'test_gigastake_app_3',
        'harmony_gigastake',
        'test_e570c841d5cd4f6197e0428ed7c517fd',
        'test_4f805bbbf96c4a649efc3f4f95616f2e',
        'test_789f9d6adcc846f1a079bf68237b5f5c',
        'test_01eac46efc9242a2be73879f1d09f1dc',
        '0.0.1',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    );
-- Insert into chains_gigastake_applications
INSERT INTO chains_gigastake_applications (chain_id, gigastake_application_id)
VALUES ('0001', 'test_gigastake_app_1'),
    ('0053', 'test_gigastake_app_2'),
    ('0040', 'test_gigastake_app_3');
INSERT INTO global_blocked_contracts (blocked_address, created_at, updated_at)
VALUES (
        '0xtest_6789abcdef0123456789abcdef01234567',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0xtest_f0123456789abcdef0123456789abcdef01',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0xtest_cdef0123456789abcdef0123456789abcdef',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0xtest_56789abcdef0123456789abcdef01234567',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    ),
    (
        '0xtest_789abcdef0123456789abcdef0123456789',
        '2022-11-11 11:11:11.000000',
        '2022-11-11 11:11:11.000000'
    );
