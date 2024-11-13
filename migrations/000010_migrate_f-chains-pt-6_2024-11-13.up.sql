-- F001
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F001', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0066'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0066';

-- F006
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F006', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0080'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0080';

-- F00A
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F00A', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0048'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0048';

-- F00B
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F00B', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0065'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0065';

-- F010
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F010', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0049'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0049';

-- F017
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F017', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = ANY(ARRAY['0071', '0072'])
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = ANY(ARRAY['0071', '0072']);

-- F018
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F018', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = '0058'
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = '0058';

-- F01C
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F01C', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = ANY(ARRAY['0069', '0070'])
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = ANY(ARRAY['0069', '0070']);

-- F025
INSERT INTO chain_alias_domains (chain_id, alias, domains)
SELECT 'F025', alias, ARRAY [alias || '.gateway.grove.city']
FROM chain_alias_domains WHERE chain_id = ANY(ARRAY['0006', 'C006'])
ON CONFLICT (chain_id, alias) DO NOTHING;
UPDATE chains SET active = false WHERE id = ANY(ARRAY['0006', 'C006']);
