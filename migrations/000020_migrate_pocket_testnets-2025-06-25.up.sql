-- Add Pocket Alpha and Pocket Beta Testnets to Portal
BEGIN;

INSERT INTO chains (id, blockchain, description, enforce_result, ticker, log_limit_blocks, active, 
scheme, node_request_timeout)
VALUES
  ('1235', 'pocket-alpha-testnet', 'pocket-alpha-testnet', 'REST', 'POKT', 100000, FALSE, 'https', 0),
  ('1236', 'pocket-beta-testnet', 'pocket-beta-testnet', 'REST', 'POKT', 100000, FALSE, 'https', 0);

INSERT INTO chain_alias_domains (chain_id, alias, domains)
VALUES
  ('1235', 'pocket-alpha-testnet', '{pocket-alpha-testnet.gateway.grove.city}'),
  ('1236', 'pocket-beta-testnet', '{pocket-beta-testnet.gateway.grove.city}');

COMMIT;
