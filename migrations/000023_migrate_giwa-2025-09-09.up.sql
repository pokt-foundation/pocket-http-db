-- Add GIWA Services
BEGIN;

INSERT INTO chains (id, blockchain, description, enforce_result, ticker, log_limit_blocks, active,
scheme, node_request_timeout)
VALUES
  ('1265', 'giwa', 'Global Infrastructure for Web3 Access', 'EVM', 'ETH', 100000, FALSE, 'https', 0),
  ('1266', 'giwa-sepolia-testnet', 'Global Infrastructure for Web3 Access Sepolia Testnet', 'EVM', 'ETH', 100000, FALSE, 'https', 0);

-- Insert corresponding chain aliases for gateway domains
INSERT INTO chain_alias_domains (chain_id, alias, domains)
VALUES
  ('1265', 'giwa', '{giwa.gateway.grove.city}'),
  ('1266', 'giwa-sepolia-testnet', '{giwa-sepolia-testnet.gateway.grove.city}');

COMMIT;
