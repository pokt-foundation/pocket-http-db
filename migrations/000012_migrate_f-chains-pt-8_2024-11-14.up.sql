-- F00E
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F00E', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = ANY(ARRAY['0077', '0078'])
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = ANY(ARRAY['0077', '0078']);

-- F012
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F012', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = ANY(ARRAY['0005', '000A'])
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = ANY(ARRAY['0005', '000A']);

-- F026
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F026', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0076'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0076';

-- F029
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F029', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0074'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0074';