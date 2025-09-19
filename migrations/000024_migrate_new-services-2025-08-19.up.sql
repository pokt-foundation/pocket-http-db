-- Add hyperliquid, unichain, stellar, bittensor, and a few LLM services
BEGIN;

-- Insert 13 new services:
INSERT INTO 
	chains (id, blockchain, description, enforce_result, ticker, log_limit_blocks, active, scheme, node_request_timeout)
VALUES
  ('1267', 'hyperliquid', 'Hyperliquid', 'JSON', 'HYPE', 100000, FALSE, 'https', 0),
  ('1268', 'unichain', 'Unichain The DeFi Chain by Uniswap', 'JSON', 'ETH', 100000, FALSE, 'https', 0),
  ('1269', 'deepseek', 'Deepseek V3 0324', 'JSON', 'AI', 100000, FALSE, 'https', 0),
  ('1270', 'grok', 'Grok', 'JSON', 'AI', 100000, FALSE, 'https', 0),
  ('1271', 'qwen', 'Qwen', 'JSON', 'AI', 100000, FALSE, 'https', 0),
  ('1272', 'gpt-oss', 'GPT OSS', 'JSON', '', 100000, FALSE, 'https', 0),
  ('1273', 'bittensor', 'Bittensor Mainnet', 'REST', 'TAO', 100000, FALSE, 'https', 0),
  ('1274', 'bittensor-testnet', 'Bittensor Testnet', 'REST', 'TAO', 100000, FALSE, 'https', 0),
  ('1275', 'planq', 'Planq  Mainnet', 'JSON', 'PLQ', 100000, FALSE, 'https', 0),
  ('1276', 'stellar-soroban', 'Stellar Soroban Mainnet', 'REST', 'XLM', 100000, FALSE, 'https', 0),
  ('1277', 'stellar-soroban-testnet', 'Stellar Soroban Testnet', 'REST', 'XLM', 100000, FALSE, 'https', 0),
  ('1278', 'stellar-horizon', 'Stellar Horizon  Mainnet', 'REST', 'XLM', 100000, FALSE, 'https', 0),
  ('1279', 'stellar-horizon-testnet', 'Stellar Horizon Testnet', 'REST', 'XLM', 100000, FALSE, 'https', 0);

-- Insert corresponding chain aliases for gateway domains
INSERT INTO chain_alias_domains (chain_id, alias, domains)
VALUES
  ('1267', 'hyperliquid', '{hyperliquid.gateway.grove.city}'),
  ('1268', 'unichain', '{unichain.gateway.grove.city}'),
  ('1269', 'deepseek', '{deepseek.gateway.grove.city}'),
  ('1270', 'grok', '{grok.gateway.grove.city}'),
  ('1271', 'qwen', '{qwen.gateway.grove.city}'),
  ('1272', 'gpt-oss', '{gpt-oss.gateway.grove.city}'),
  ('1273', 'bittensor', '{bittensor.gateway.grove.city}'),
  ('1274', 'bittensor-testnet', '{bittensor-testnet.gateway.grove.city}'),
  ('1275', 'planq', '{planq.gateway.grove.city}'),
  ('1276', 'stellar-soroban', '{stellar-soroban.gateway.grove.city}'),
  ('1277', 'stellar-soroban-testnet', '{stellar-soroban-testnet.gateway.grove.city}'),
  ('1278', 'stellar-horizon', '{stellar-horizon.gateway.grove.city}'),
  ('1279', 'stellar-horizon-testnet', '{stellar-horizon-testnet.gateway.grove.city}');

COMMIT;
