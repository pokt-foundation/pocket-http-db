-- F001
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F001', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0066'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0066';

-- F00E
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F00E', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = ANY(ARRAY['0077', '0078'])
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = ANY(ARRAY['0077', '0078']);

-- F00F
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F00F', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0046'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0046';

-- F011
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F011', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = 'A088'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = 'A088';

-- F012
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F012', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0005'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0005';

-- F016
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F016', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0056'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0056';

-- F01A
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F01A', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0051'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0051';

-- F021
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F021', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = ANY(ARRAY['0009', '000B'])
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = ANY(ARRAY['0009', '000B']);

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