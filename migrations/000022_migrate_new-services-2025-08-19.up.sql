-- Add Kleomedes Cosmos Chains to Portal
BEGIN;

INSERT INTO chains (id, blockchain, description, enforce_result, ticker, log_limit_blocks, active, 
scheme, node_request_timeout)
VALUES
  ('1237', 'soneium', 'Soneium', 'JSON', 'SONY', 100000, FALSE, 'https', 0),
  ('1238', 'atomone', 'AtomOne', 'REST', 'ATONE', 100000, FALSE, 'https', 0),
  ('1239', 'akash', 'Akash Network', 'REST', 'AKT', 100000, FALSE, 'https', 0),
  ('1240', 'fetch', 'ASI Network', 'REST', 'FET', 100000, FALSE, 'https', 0),
  ('1241', 'persistence', 'Persistence', 'REST', 'XPRT', 100000, FALSE, 'https', 0),
  ('1242', 'router', 'Router Protocol', 'REST', 'ROUTE', 100000, FALSE, 'https', 0),
  ('1243', 'seda', 'Seda Protocol', 'REST', 'SEDA', 100000, FALSE, 'https', 0),
  ('1244', 'shentu', 'shentu', 'REST', 'CTK', 100000, FALSE, 'https', 0);

INSERT INTO chain_alias_domains (chain_id, alias, domains)
VALUES
  ('1237', 'soneium', '{soneium.gateway.grove.city}'),
  ('1238', 'atomone', '{atomone.gateway.grove.city}'),
  ('1239', 'akash', '{akash.gateway.grove.city}'),
  ('1240', 'fetch', '{fetch.gateway.grove.city}'),
  ('1241', 'persistence', '{persistence.gateway.grove.city}'),
  ('1242', 'router', '{router.gateway.grove.city}'),
  ('1243', 'seda', '{seda.gateway.grove.city}'),
  ('1244', 'shentu', '{shentu.gateway.grove.city}');

COMMIT;
