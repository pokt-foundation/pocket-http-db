-- Add XRPL EVM Testnet to Portal
BEGIN;
INSERT INTO chains (id, blockchain, description, enforce_result, ticker, log_limit_blocks, active, 
scheme, node_request_timeout)
VALUES
  ('F036', 'xrpl-evm-testnet', 'XRPL EVM Testnet', 'JSON', 'XRP', 100000, FALSE, 'https', 0);

INSERT INTO chain_alias_domains (chain_id, alias, domains)
VALUES
  ('F036', 'xrpl-evm-testnet', '{xrpl-evm-testnet.gateway.grove.city}');

INSERT INTO chain_config (chain_id, log_limit_blocks)
VALUES
  ('F036', 100000);

INSERT INTO chain_checks (chain_id, type, payload, result_key, evm_chain_id)
VALUES
  ('F036', 'chain', '{"method":"eth_chainId","id":1,"jsonrpc":"2.0"}', 'id', 1329);

INSERT INTO chain_checks (chain_id, type, payload, result_key, allowance)
VALUES
  ('F036', 'sync', '{"method":"eth_blockNumber","id":1,"jsonrpc":"2.0"}', 'result', 2);

INSERT INTO chains_gigastake_applications
VALUES
  ('F036', 'e8ef9d09'),
  ('F036', '99fd95a8'),
  ('F036', 'c002a1f3');

COMMIT;
