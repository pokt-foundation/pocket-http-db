-- Update Chains <-> Gigastakes Mapping for February 2025 new chains
BEGIN;
-- update the existing chains
UPDATE chains
SET
  blockchain = 'xrpl-evm-devnet',
  description = 'XRPL EVM Devnet'
WHERE id = 'F02C';

UPDATE chains
SET
  blockchain = 'berachain-bartio-testnet',
  description = 'Berachain bArtio Testnet'
WHERE id = 'F031';

-- add the new chain
INSERT INTO chains (id, blockchain, description, enforce_result, ticker, log_limit_blocks, active, 
scheme, node_request_timeout)
VALUES
  ('F035', 'berachain', 'Berachain', 'JSON', 'BERA', 100000, FALSE, 'https', 0);

-- update old aliases
UPDATE chain_alias_domains
SET 
  alias = 'berachain-bartio-testnet',
  domains = '{berachain-bartio-testnet.gateway.grove.city}'
WHERE alias = 'berachain';

UPDATE chain_alias_domains
SET 
  alias = 'xrpl-evm-devnet',
  domains = '{xrpl-evm-devnet.gateway.grove.city}'
WHERE alias = 'xrpl-evm';

-- add chain aliases
INSERT INTO chain_alias_domains (chain_id, alias, domains)
VALUES
  ('F035', 'berachain', '{berachain.gateway.grove.city}');

-- add chain configs (mostly legacy)
INSERT INTO chain_config (chain_id, log_limit_blocks)
VALUES
  ('F035', 100000);

-- add eth chain checks
INSERT INTO chain_checks (chain_id, type, payload, result_key, evm_chain_id)
VALUES
  ('F035', 'chain', '{"method":"eth_chainId","id":1,"jsonrpc":"2.0"}', 'id', 80085);

-- add sync checks
INSERT INTO chain_checks (chain_id, type, payload, result_key, allowance)
VALUES
  ('F035', 'sync', '{"method":"eth_blockNumber","id":1,"jsonrpc":"2.0"}', 'result', 2);

-- Create new  chain_id <-> gigastakes mapping in chains_gigastake_applications
INSERT INTO chains_gigastake_applications (chain_id, gigastake_application_id)
VALUES
  ('F035', 'd8b9739d'),
  ('F035', 'c71280d7'),
  ('F035', '149199e3');

COMMIT;
