-- Update Chains <-> Gigastakes Mapping for February 2025 new chains
BEGIN;
-- add the new chains
INSERT INTO chains (id, blockchain, description, enforce_result, ticker, log_limit_blocks, active, 
scheme, node_request_timeout)
VALUES
  ('F02C', 'xrpl-evm', 'XRPL EVM Sidechain', 'JSON', 'XRP', 100000, FALSE, 'https', 0),
  ('F02D', 'sonic', 'Sonic', 'JSON', 'S', 100000, FALSE, 'https', 0),
  ('F02E', 'tron', 'TRON', 'JSON', 'TRX', 100000, FALSE, 'https', 0),
  ('F030', 'linea', 'Linea', 'JSON', 'ETH', 100000, FALSE, 'https', 0),
  ('F031', 'berachain', 'Berachain', 'JSON', 'BERA', 100000, FALSE, 'https', 0),
  ('F032', 'ink', 'Ink', 'JSON', 'ETH', 100000, FALSE, 'https', 0),
  ('F033', 'mantle', 'Mantle', 'JSON', 'MNT', 100000, FALSE, 'https', 0),
  ('F034', 'sei', 'Sei', 'JSON', 'SEI', 100000, FALSE, 'https', 0);

-- add chain aliases
INSERT INTO chain_alias_domains (chain_id, alias, domains)
VALUES
  ('F02C', 'xrpl-evm', '{xrpl-evm.gateway.grove.city}'),
  ('F02D', 'sonic', '{sonic.gateway.grove.city}'),
  ('F02E', 'tron', '{tron.gateway.grove.city}'),
  ('F030', 'linea', '{linea.gateway.grove.city}'),
  ('F031', 'berachain', '{berachain.gateway.grove.city}'),
  ('F032', 'ink', '{ink.gateway.grove.city}'),
  ('F033', 'mantle', '{mantle.gateway.grove.city}'),
  ('F034', 'sei', '{sei.gateway.grove.city}');

-- add chain configs (mostly legacy)
INSERT INTO chain_config (chain_id, log_limit_blocks)
VALUES
  ('F02C', 100000),
  ('F02D', 100000),
  ('F02E', 100000),
  ('F030', 100000),
  ('F031', 100000),
  ('F032', 100000),
  ('F033', 100000),
  ('F034', 100000);

-- add eth chain checks
INSERT INTO chain_checks (chain_id, type, payload, result_key, evm_chain_id)
VALUES
  ('F02C', 'chain', '{"method":"eth_chainId","id":1,"jsonrpc":"2.0"}', 'id', 1440002),
  ('F02D', 'chain', '{"method":"eth_chainId","id":1,"jsonrpc":"2.0"}', 'id', 146),
  ('F02E', 'chain', '{"method":"eth_chainId","id":1,"jsonrpc":"2.0"}', 'id', 728126428),
  ('F030', 'chain', '{"method":"eth_chainId","id":1,"jsonrpc":"2.0"}', 'id', 59144),
  ('F031', 'chain', '{"method":"eth_chainId","id":1,"jsonrpc":"2.0"}', 'id', 80084),
  ('F032', 'chain', '{"method":"eth_chainId","id":1,"jsonrpc":"2.0"}', 'id', 57073),
  ('F033', 'chain', '{"method":"eth_chainId","id":1,"jsonrpc":"2.0"}', 'id', 5000),
  ('F034', 'chain', '{"method":"eth_chainId","id":1,"jsonrpc":"2.0"}', 'id', 1329);

-- add sync checks
INSERT INTO chain_checks (chain_id, type, payload, result_key, allowance)
VALUES
  ('F02C', 'sync', '{"method":"eth_blockNumber","id":1,"jsonrpc":"2.0"}', 'result', 2),
  ('F02D', 'sync', '{"method":"eth_blockNumber","id":1,"jsonrpc":"2.0"}', 'result', 2),
  ('F02E', 'sync', '{"method":"eth_blockNumber","id":1,"jsonrpc":"2.0"}', 'result', 2),
  ('F030', 'sync', '{"method":"eth_blockNumber","id":1,"jsonrpc":"2.0"}', 'result', 2),
  ('F031', 'sync', '{"method":"eth_blockNumber","id":1,"jsonrpc":"2.0"}', 'result', 2),
  ('F032', 'sync', '{"method":"eth_blockNumber","id":1,"jsonrpc":"2.0"}', 'result', 2),
  ('F033', 'sync', '{"method":"eth_blockNumber","id":1,"jsonrpc":"2.0"}', 'result', 2),
  ('F034', 'sync', '{"method":"eth_blockNumber","id":1,"jsonrpc":"2.0"}', 'result', 2);

-- Update the chain_id in chains_gigastake_applications based on the mapping
UPDATE chains_gigastake_applications AS cga
SET chain_id = CASE
    WHEN ga.address = '3c50afb94de9d9719f9c23bd2976858c74ad44eb' THEN 'F02C'
    WHEN ga.address = '5266fa808ed9a6a5e915894570d46e3d451c1267' THEN 'F02C'
    WHEN ga.address = '62e344a58a141935e4f00766bce8f8c1c2736d61' THEN 'F02C'
    WHEN ga.address = '6a54693f8169ef5915215abfc3c0f1a44401f688' THEN 'F02D'
    WHEN ga.address = '832db8020cdfdbf65194f2276905950682b69121' THEN 'F02D'
    WHEN ga.address = '875d2e01a07c92f32a10e63cec0985fe6a9eb758' THEN 'F02D'
    WHEN ga.address = 'a2e4fee5687a6fa4012f7ba1449a2b4fd3bb32b9' THEN 'F02E'
    WHEN ga.address = 'a712d116669653f019b2773c8d4ca47956fa5492' THEN 'F02E'
    WHEN ga.address = 'b728fb333dfdc215aee85c99cebe6386895c743a' THEN 'F02E'
    WHEN ga.address = 'b90d686e3c20ca41d795fede44603f17eed04ca3' THEN 'F030'
    WHEN ga.address = 'bd78e8bc24b73fb7977e4db3a781d55e531e73e4' THEN 'F030'
    WHEN ga.address = 'ccd7e9e95ba790db57a9134124c0e728f42c6e6e' THEN 'F030'
    WHEN ga.address = 'd009929986304b7beb2e90fb560827678c39d056' THEN 'F031'
    WHEN ga.address = 'd5f279c2641e61d9404bb32c070dd3c798cef402' THEN 'F031'
    WHEN ga.address = 'daa46b19faa46e699f07e135bd2211d7ea32345f' THEN 'F031'
    WHEN ga.address = 'dbbf3693f4c2a13460a84c1c848878a727a34a93' THEN 'F032'
    WHEN ga.address = 'e0855387204668b75a7127996f2b52cbaa33786c' THEN 'F032'
    WHEN ga.address = 'ec335e16cf90e5d37ef02123b2fd476802cdad06' THEN 'F032'
    WHEN ga.address = 'f271965de3d862e63e2258d4618687f30c9cf5b9' THEN 'F033'
    WHEN ga.address = 'f743db5998ba07ae4cbbacb60e62ed194bedd084' THEN 'F033'
    WHEN ga.address = 'f9051e866a8185c0a25e7536ad54f8a12e99d312' THEN 'F033'
    WHEN ga.address = 'f929087a312d0d111118209f814a2bab03179bb6' THEN 'F034'
    WHEN ga.address = 'f950687c021ae45f2df1e5481d22217ea3073d37' THEN 'F034'
    WHEN ga.address = 'fec14897e8427bc025fe4ebc6f7ee6098eac85b8' THEN 'F034'
    ELSE chain_id -- Keep current value if not matched
END
FROM gigastake_applications AS ga
WHERE ga.id = cga.gigastake_application_id
  AND ga.address IN (
    '3c50afb94de9d9719f9c23bd2976858c74ad44eb',
    '5266fa808ed9a6a5e915894570d46e3d451c1267',
    '62e344a58a141935e4f00766bce8f8c1c2736d61',
    '6a54693f8169ef5915215abfc3c0f1a44401f688',
    '832db8020cdfdbf65194f2276905950682b69121',
    '875d2e01a07c92f32a10e63cec0985fe6a9eb758',
    'a2e4fee5687a6fa4012f7ba1449a2b4fd3bb32b9',
    'a712d116669653f019b2773c8d4ca47956fa5492',
    'b728fb333dfdc215aee85c99cebe6386895c743a',
    'b90d686e3c20ca41d795fede44603f17eed04ca3',
    'bd78e8bc24b73fb7977e4db3a781d55e531e73e4',
    'ccd7e9e95ba790db57a9134124c0e728f42c6e6e',
    'd009929986304b7beb2e90fb560827678c39d056',
    'd5f279c2641e61d9404bb32c070dd3c798cef402',
    'daa46b19faa46e699f07e135bd2211d7ea32345f',
    'dbbf3693f4c2a13460a84c1c848878a727a34a93',
    'e0855387204668b75a7127996f2b52cbaa33786c',
    'ec335e16cf90e5d37ef02123b2fd476802cdad06',
    'f271965de3d862e63e2258d4618687f30c9cf5b9',
    'f743db5998ba07ae4cbbacb60e62ed194bedd084',
    'f9051e866a8185c0a25e7536ad54f8a12e99d312',
    'f929087a312d0d111118209f814a2bab03179bb6',
    'f950687c021ae45f2df1e5481d22217ea3073d37',
    'fec14897e8427bc025fe4ebc6f7ee6098eac85b8'
);
COMMIT;
