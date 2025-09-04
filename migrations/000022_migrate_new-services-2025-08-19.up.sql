-- Add Kleomedes Cosmos Chains and Hey service to Portal
BEGIN;

-- Insert new chains:
-- - Soneium (EVM chain)
-- - 27 Cosmos chains
-- - Hey service (Grove internal)
INSERT INTO chains (id, blockchain, description, enforce_result, ticker, log_limit_blocks, active,
scheme, node_request_timeout)
VALUES
  -- EVM chain
  ('1237', 'soneium', 'Soneium', 'JSON', 'SONY', 100000, FALSE, 'https', 0),
  -- Cosmos chains
  ('1238', 'atomone', 'AtomOne', 'REST', 'ATONE', 100000, FALSE, 'https', 0),
  ('1239', 'akash', 'Akash Network', 'REST', 'AKT', 100000, FALSE, 'https', 0),
  ('1240', 'fetch', 'ASI Network', 'REST', 'FET', 100000, FALSE, 'https', 0),
  ('1241', 'persistence', 'Persistence', 'REST', 'XPRT', 100000, FALSE, 'https', 0),
  ('1242', 'router', 'Router Protocol', 'REST', 'ROUTE', 100000, FALSE, 'https', 0),
  ('1243', 'seda', 'Seda Protocol', 'REST', 'SEDA', 100000, FALSE, 'https', 0),
  ('1244', 'shentu', 'shentu', 'REST', 'CTK', 100000, FALSE, 'https', 0),
  ('1245', 'arkeo', 'Arkeo', 'REST', 'ARKEO', 100000, FALSE, 'https', 0),
  ('1246', 'babylon', 'Babylon Genesis', 'REST', 'BABY', 100000, FALSE, 'https', 0),
  ('1247', 'celestia', 'Celestia', 'REST', 'TIA', 100000, FALSE, 'https', 0),
  ('1248', 'cheqd', 'Cheqd', 'REST', 'CHEQ', 100000, FALSE, 'https', 0),
  ('1249', 'chihuahua', 'Chihuahua', 'REST', 'HUAHUA', 100000, FALSE, 'https', 0),
  ('1250', 'cosmoshub', 'Cosmos Hub', 'REST', 'ATOM', 100000, FALSE, 'https', 0),
  ('1251', 'elys-network', 'Elys Network', 'REST', 'ELYS', 100000, FALSE, 'https', 0),
  ('1252', 'jackal', 'Jackal Protocol', 'REST', 'JKL', 100000, FALSE, 'https', 0),
  ('1253', 'juno', 'Juno Network', 'REST', 'JUNO', 100000, FALSE, 'https', 0),
  ('1254', 'kyve', 'Kyve', 'REST', 'KYVE', 100000, FALSE, 'https', 0),
  ('1255', 'namada', 'Namada', 'REST', 'NAM', 100000, FALSE, 'https', 0),
  ('1256', 'neutron', 'Neutron', 'REST', 'NTRN', 100000, FALSE, 'https', 0),
  ('1257', 'nillion', 'Nillion', 'REST', 'NIL', 100000, FALSE, 'https', 0),
  ('1258', 'passage', 'Passage', 'REST', 'PASG', 100000, FALSE, 'https', 0),
  ('1259', 'provenance', 'Provenance', 'REST', 'HASH', 100000, FALSE, 'https', 0),
  ('1260', 'quicksilver', 'Quicksilver Protocol', 'REST', 'QCK', 100000, FALSE, 'https', 0),
  ('1261', 'side-protocol', 'Side Protocol', 'REST', 'SIDE', 100000, FALSE, 'https', 0),
  ('1262', 'stargaze', 'Stargaze', 'REST', 'STARS', 100000, FALSE, 'https', 0),
  ('1263', 'stride', 'Stride', 'REST', 'STRD', 100000, FALSE, 'https', 0),
  -- Internal Grove service
  ('1264', 'hey', 'Grove hey Service', 'REST', 'HEY', 100000, FALSE, 'https', 0);

-- Insert corresponding chain aliases for gateway domains
INSERT INTO chain_alias_domains (chain_id, alias, domains)
VALUES
  ('1237', 'soneium', '{soneium.gateway.grove.city}'),
  ('1238', 'atomone', '{atomone.gateway.grove.city}'),
  ('1239', 'akash', '{akash.gateway.grove.city}'),
  ('1240', 'fetch', '{fetch.gateway.grove.city}'),
  ('1241', 'persistence', '{persistence.gateway.grove.city}'),
  ('1242', 'router', '{router.gateway.grove.city}'),
  ('1243', 'seda', '{seda.gateway.grove.city}'),
  ('1244', 'shentu', '{shentu.gateway.grove.city}'),
  ('1245', 'arkeo', '{arkeo.gateway.grove.city}'),
  ('1246', 'babylon', '{babylon.gateway.grove.city}'),
  ('1247', 'celestia', '{celestia.gateway.grove.city}'),
  ('1248', 'cheqd', '{cheqd.gateway.grove.city}'),
  ('1249', 'chihuahua', '{chihuahua.gateway.grove.city}'),
  ('1250', 'cosmoshub', '{cosmoshub.gateway.grove.city}'),
  ('1251', 'elys-network', '{elys-network.gateway.grove.city}'),
  ('1252', 'jackal', '{jackal.gateway.grove.city}'),
  ('1253', 'juno', '{juno.gateway.grove.city}'),
  ('1254', 'kyve', '{kyve.gateway.grove.city}'),
  ('1255', 'namada', '{namada.gateway.grove.city}'),
  ('1256', 'neutron', '{neutron.gateway.grove.city}'),
  ('1257', 'nillion', '{nillion.gateway.grove.city}'),
  ('1258', 'passage', '{passage.gateway.grove.city}'),
  ('1259', 'provenance', '{provenance.gateway.grove.city}'),
  ('1260', 'quicksilver', '{quicksilver.gateway.grove.city}'),
  ('1261', 'side-protocol', '{side-protocol.gateway.grove.city}'),
  ('1262', 'stargaze', '{stargaze.gateway.grove.city}'),
  ('1263', 'stride', '{stride.gateway.grove.city}'),
  ('1264', 'hey', '{hey.gateway.grove.city}');

COMMIT;
