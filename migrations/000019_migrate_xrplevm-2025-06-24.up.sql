-- Add XRPL EVM Testnet to Portal
BEGIN;

-- TODO #357 - Using placeholder ID 1234 instead of xrplevm
INSERT INTO chains (id, blockchain, description, enforce_result, ticker, log_limit_blocks, active, 
scheme, node_request_timeout)
VALUES
  ('1234', 'xrplevm', 'XRPLEVM', 'JSON', 'XRP', 100000, FALSE, 'https', 0);

INSERT INTO chain_alias_domains (chain_id, alias, domains)
VALUES
  ('1234', 'xrplevm', '{xrplevm.gateway.grove.city}');

COMMIT;
