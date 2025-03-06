-- Adding Archival Checks for Kaia and NEAR F Chains
BEGIN;

INSERT INTO chain_checks (chain_id, type, payload)
VALUES
  ('F016', 'archival', '{"method":"eth_getBalance","params":["0xe5Fb31A5CaEE6a96de393bdBF89FBe65fe125Bb3", "0x1"],"id":1,"jsonrpc":"2.0"}'),
  ('F01B', 'archival', '{"jsonrpc": "2.0","id": "1","method": "query","params": {"request_type": "view_account","block_id": 14999888,"account_id": "zavodil.poolv1.near"}}');

COMMIT;
